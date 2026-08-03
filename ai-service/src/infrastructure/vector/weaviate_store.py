import asyncio
import json
import time
import uuid as uuid_lib
from typing import Any

import httpx
import weaviate
from weaviate.classes.init import AdditionalConfig, Auth, Timeout
from weaviate.client import WeaviateAsyncClient

from application.errors import IndexingFailed, SimilaritySearchFailed
from domain.entities import (
    Chunk,
    ContentKind,
    Embedding,
    ScoredChunk,
    SimilarPost,
)
from infrastructure.config.settings import WeaviateSettings
from infrastructure.observability import WEAVIATE_CALLS, get_logger, observe
from infrastructure.vector.schema import (
    content_chunk_properties,
    content_chunk_vector_config,
)

logger = get_logger(__name__)

_NAMESPACE = uuid_lib.UUID("6f1a4c2a-aaaa-4b21-9b07-9a8c2f1c5e21")


def _chunk_uuid(kind: ContentKind, source_id: str, segment_idx: int) -> str:
    return str(uuid_lib.uuid5(_NAMESPACE, f"{kind}:{source_id}:{segment_idx}"))


def _properties(chunk: Chunk) -> dict[str, Any]:
    meta = chunk.metadata or {}
    props: dict[str, Any] = {
        "sourceId": chunk.source_id,
        "kind": chunk.kind,
        "text": chunk.text,
        "segmentIdx": chunk.segment_idx,
        "category": meta.get("category", ""),
        "tags": _split_tags(meta.get("tags", "")),
        "parentId": meta.get("parent_id", ""),
        "parentText": meta.get("parent_text", ""),
        "attachedPostId": meta.get("attached_post_id", ""),
        "placeId": meta.get("place_id", ""),
    }
    coord = _coordinate(meta)
    if coord is not None:
        props["coordinate"] = coord
    return props


def _split_tags(raw: str) -> list[str]:
    if not raw:
        return []
    return [t.strip() for t in raw.split(",") if t.strip()]


def _coordinate(meta: dict[str, str]) -> dict[str, float] | None:
    lat = meta.get("lat")
    lon = meta.get("lon")
    if lat is None or lon is None:
        return None
    try:
        return {"latitude": float(lat), "longitude": float(lon)}
    except (TypeError, ValueError):
        return None


def _combine_where(operands: list[str]) -> str | None:
    if not operands:
        return None
    if len(operands) == 1:
        return operands[0]
    return f"{{ operator: And, operands: [{', '.join(operands)}] }}"


def _near_vector(embedding: Embedding, distance: float | None) -> str:
    parts = [f"vector: {json.dumps(list(embedding))}"]
    if distance is not None:
        parts.append(f"distance: {distance}")
    return f"nearVector: {{ {', '.join(parts)} }}"


_QUERY_LOG_LIMIT = 800
_RETRY_HEADER = "x-retry-count"
_MAX_RETRY_BACKOFF_S = 30.0
_CONNECT_ATTEMPTS = 3
_CONNECT_BACKOFF_S = 1.0


def _truncate_query(query: str) -> str:
    if len(query) <= _QUERY_LOG_LIMIT:
        return query
    return query[:_QUERY_LOG_LIMIT] + "...(truncated)"


def _summarize_graphql_errors(errors: Any) -> list[dict[str, Any]]:
    if not isinstance(errors, list):
        return [{"message": str(errors)}]
    summary: list[dict[str, Any]] = []
    for item in errors:
        if isinstance(item, dict):
            summary.append(
                {
                    "message": item.get("message"),
                    "path": item.get("path"),
                    "locations": item.get("locations"),
                }
            )
        else:
            summary.append({"message": str(item)})
    return summary


def _http_error_transient(exc: httpx.HTTPError) -> bool:
    if isinstance(exc, httpx.TimeoutException):
        return True
    if isinstance(exc, httpx.HTTPStatusError):
        status = exc.response.status_code
        if status >= 500:
            return True
        if status == 429:
            return True
        return False
    return True


def _connect_error_transient(exc: BaseException) -> bool:
    if isinstance(exc, (httpx.TimeoutException, TimeoutError, asyncio.TimeoutError, ConnectionError, OSError)):
        return True
    name = type(exc).__name__.lower()
    return "timeout" in name or "connect" in name


def _response_body_preview(response: httpx.Response, *, limit: int = 500) -> str:
    text = response.text
    if len(text) <= limit:
        return text
    return text[:limit] + "...(truncated)"


class WeaviateVectorStore:
    def __init__(self, settings: WeaviateSettings) -> None:
        self._settings = settings
        self._client: WeaviateAsyncClient | None = None
        self._http: httpx.AsyncClient | None = None
        self._connect_lock = asyncio.Lock()

    @property
    def is_connected(self) -> bool:
        return self._client is not None and self._http is not None

    async def connect(self, *, required: bool = True) -> bool:
        """Connect to Weaviate. If required=False, log and return False on failure."""
        auth = Auth.api_key(self._settings.api_key) if self._settings.api_key else None
        logger.info("weaviate.connect.start", url=self._settings.url)
        client = weaviate.use_async_with_weaviate_cloud(
            cluster_url=self._settings.url,
            auth_credentials=auth,
            additional_config=AdditionalConfig(
                timeout=Timeout(init=30, query=10, insert=120),
            ),
            skip_init_checks=True,
        )
        headers = {}
        if self._settings.api_key:
            headers["Authorization"] = f"Bearer {self._settings.api_key}"
        http = httpx.AsyncClient(
            base_url=self._settings.url,
            headers=headers,
            timeout=httpx.Timeout(20.0),
        )
        try:
            await client.connect()
            self._client = client
            self._http = http
            await self._ensure_class()
            logger.info("weaviate.connect.done", url=self._settings.url)
            return True
        except BaseException as exc:
            log = logger.error if required else logger.warning
            log(
                "weaviate.connect.failed",
                url=self._settings.url,
                required=required,
                transient=_connect_error_transient(exc),
                error=str(exc),
            )
            try:
                await client.close()
            except Exception:
                pass
            await http.aclose()
            self._client = None
            self._http = None
            if required:
                raise
            return False

    async def _ensure_connected(self) -> None:
        if self.is_connected:
            return
        async with self._connect_lock:
            if self.is_connected:
                return
            last_exc: BaseException | None = None
            for attempt in range(1, _CONNECT_ATTEMPTS + 1):
                try:
                    await self.connect(required=True)
                    return
                except BaseException as exc:
                    last_exc = exc
                    transient = _connect_error_transient(exc)
                    if not transient or attempt >= _CONNECT_ATTEMPTS:
                        raise
                    delay = min(_CONNECT_BACKOFF_S * (2 ** (attempt - 1)), _MAX_RETRY_BACKOFF_S)
                    logger.warning(
                        "weaviate.connect.retry",
                        attempt=attempt,
                        max_attempts=_CONNECT_ATTEMPTS,
                        delay_s=delay,
                        error=str(exc),
                    )
                    await asyncio.sleep(delay)
            if last_exc is not None:
                raise last_exc

    async def close(self) -> None:
        if self._client is not None:
            await self._client.close()
            self._client = None
        if self._http is not None:
            await self._http.aclose()
            self._http = None

    async def _graphql(self, query: str, *, op: str) -> list[dict[str, Any]]:
        assert self._http is not None
        class_name = self._settings.class_name
        try:
            resp = await self._http.post("/v1/graphql", json={"query": query})
            resp.raise_for_status()
        except httpx.HTTPError as exc:
            transient = _http_error_transient(exc)
            details: dict[str, Any] = {
                "op": op,
                "class_name": class_name,
                "transient": transient,
                "error": str(exc),
            }
            if isinstance(exc, httpx.HTTPStatusError):
                details["status_code"] = exc.response.status_code
                details["response_body"] = _response_body_preview(exc.response)
            logger.error("weaviate.graphql.http_failed", **details)
            raise SimilaritySearchFailed(
                f"weaviate graphql {op} failed: {exc}",
                cause=exc,
                transient=transient,
            ) from exc
        body = resp.json()
        errors = body.get("errors")
        if errors:
            summary = _summarize_graphql_errors(errors)
            logger.error(
                "weaviate.graphql.errors",
                op=op,
                class_name=class_name,
                transient=False,
                errors=summary,
                query=_truncate_query(query),
            )
            raise SimilaritySearchFailed(
                f"weaviate graphql {op} errors: {summary}",
                transient=False,
            )
        data = body.get("data") or {}
        get = data.get("Get") or {}
        return get.get(class_name) or []

    async def _ensure_class(self) -> None:
        assert self._client is not None
        name = self._settings.class_name
        if not await self._client.collections.exists(name):
            await self._client.collections.create(
                name=name,
                properties=content_chunk_properties(),
                vector_config=content_chunk_vector_config(),
            )
            logger.info("weaviate.class_created", name=name)
            return
        collection = self._client.collections.get(name)
        config = await collection.config.get()
        existing = {p.name for p in (config.properties or [])}
        if "placeId" not in existing:
            from weaviate.classes.config import DataType, Property

            await collection.config.add_property(
                Property(name="placeId", data_type=DataType.TEXT, skip_vectorization=True)
            )
            logger.info("weaviate.property_added", name=name, property="placeId")

    async def _batch_delete(self, source_id: str, kind: str) -> None:
        assert self._http is not None
        body = {
            "match": {
                "class": self._settings.class_name,
                "where": {
                    "operator": "And",
                    "operands": [
                        {"path": ["sourceId"], "operator": "Equal", "valueText": source_id},
                        {"path": ["kind"], "operator": "Equal", "valueText": kind},
                    ],
                },
            }
        }
        resp = await self._http.request("DELETE", "/v1/batch/objects", json=body)
        resp.raise_for_status()

    async def _batch_insert(self, objects: list[dict[str, Any]]) -> None:
        assert self._http is not None
        resp = await self._http.post("/v1/batch/objects", json={"objects": objects})
        resp.raise_for_status()
        errors = [
            item["result"]["errors"]
            for item in resp.json()
            if item.get("result", {}).get("errors")
        ]
        if errors:
            raise IndexingFailed(f"weaviate batch insert errors: {errors}")

    async def upsert(self, chunks: list[Chunk], embeddings: list[Embedding]) -> None:
        if len(chunks) != len(embeddings):
            raise ValueError("chunks and embeddings length mismatch")
        if not chunks:
            return
        await self._ensure_connected()
        first = chunks[0]
        logger.info(
            "weaviate.upsert.start",
            source_id=first.source_id,
            kind=first.kind,
            chunks=len(chunks),
        )
        objects = [
            {
                "class": self._settings.class_name,
                "id": _chunk_uuid(chunk.kind, chunk.source_id, chunk.segment_idx),
                "properties": _properties(chunk),
                "vectors": {"default": list(vector)},
            }
            for chunk, vector in zip(chunks, embeddings, strict=True)
        ]
        started = time.monotonic()
        async with observe(WEAVIATE_CALLS, op="upsert"):
            try:
                await self._batch_delete(first.source_id, first.kind)
                await self._batch_insert(objects)
            except httpx.HTTPError as exc:
                logger.error(
                    "weaviate.upsert.failed",
                    source_id=first.source_id,
                    elapsed_s=round(time.monotonic() - started, 3),
                    error=str(exc),
                )
                raise IndexingFailed(f"weaviate upsert failed: {exc}", cause=exc) from exc
            except IndexingFailed:
                logger.error(
                    "weaviate.upsert.failed",
                    source_id=first.source_id,
                    elapsed_s=round(time.monotonic() - started, 3),
                )
                raise
        logger.info(
            "weaviate.upsert.done",
            source_id=first.source_id,
            elapsed_s=round(time.monotonic() - started, 3),
        )

    async def delete_by_source(self, source_id: str, kind: ContentKind) -> None:
        await self._ensure_connected()
        async with observe(WEAVIATE_CALLS, op="delete_by_source"):
            try:
                await self._batch_delete(source_id, kind)
            except httpx.HTTPError as exc:
                raise IndexingFailed(
                    f"weaviate delete_by_source failed: {exc}", cause=exc
                ) from exc

    async def search(
        self,
        query_embedding: Embedding,
        limit: int,
        *,
        kinds: list[ContentKind] | None = None,
        place_id: str | None = None,
        lat: float | None = None,
        lon: float | None = None,
        radius_m: float | None = None,
    ) -> list[ScoredChunk]:
        return await self._search(
            query_embedding,
            limit=limit,
            kinds=kinds,
            place_id=place_id,
            lat=lat,
            lon=lon,
            radius_m=radius_m,
        )

    async def search_dedup(
        self,
        query_embedding: Embedding,
        distance_threshold: float,
        limit: int,
        *,
        exclude_source_id: str | None = None,
        lat: float | None = None,
        lon: float | None = None,
        radius_m: float | None = None,
    ) -> list[ScoredChunk]:
        return await self._search(
            query_embedding,
            limit=limit,
            kinds=["post"],
            distance=distance_threshold,
            exclude_source_id=exclude_source_id,
            lat=lat,
            lon=lon,
            radius_m=radius_m,
        )

    async def find_similar_posts(
        self,
        query_embedding: Embedding,
        *,
        exclude_post_id: str,
        top_n: int,
        oversample: int,
    ) -> list[SimilarPost]:
        if top_n <= 0:
            return []
        await self._ensure_connected()
        candidate_limit = max(top_n * max(oversample, 1), top_n)
        where = _combine_where(
            [
                '{ path: ["kind"], operator: Equal, valueText: "post" }',
                f'{{ path: ["sourceId"], operator: NotEqual, valueText: {json.dumps(exclude_post_id)} }}',
            ]
        )
        query = (
            "{ Get { " + self._settings.class_name + "("
            f"limit: {candidate_limit}, {_near_vector(query_embedding, None)}, where: {where}"
            ") { sourceId _additional { distance } } } }"
        )
        logger.info(
            "weaviate.find_similar_posts.start",
            exclude_post_id=exclude_post_id,
            top_n=top_n,
            candidate_limit=candidate_limit,
        )
        started = time.monotonic()
        async with observe(WEAVIATE_CALLS, op="find_similar_posts"):
            try:
                objs = await self._graphql(query, op="find_similar_posts")
            except SimilaritySearchFailed as exc:
                logger.error(
                    "weaviate.find_similar_posts.failed",
                    exclude_post_id=exclude_post_id,
                    elapsed_s=round(time.monotonic() - started, 3),
                    transient=exc.transient,
                    error=exc.message,
                )
                raise
        logger.info(
            "weaviate.find_similar_posts.done",
            exclude_post_id=exclude_post_id,
            elapsed_s=round(time.monotonic() - started, 3),
            hits=len(objs),
        )

        best: dict[str, float] = {}
        for obj in objs:
            source_id = str(obj.get("sourceId", ""))
            if not source_id:
                continue
            distance = (obj.get("_additional") or {}).get("distance")
            if distance is None:
                continue
            score = 1.0 - float(distance)
            current = best.get(source_id)
            if current is None or score > current:
                best[source_id] = score
        ranked = sorted(best.items(), key=lambda kv: kv[1], reverse=True)
        return [SimilarPost(post_id=pid, score=score) for pid, score in ranked[:top_n]]

    async def get_post_embedding(self, post_id: str) -> Embedding | None:
        await self._ensure_connected()
        where = _combine_where(
            [
                '{ path: ["kind"], operator: Equal, valueText: "post" }',
                f'{{ path: ["sourceId"], operator: Equal, valueText: {json.dumps(post_id)} }}',
            ]
        )
        query = (
            "{ Get { " + self._settings.class_name + "("
            f"limit: 200, where: {where}"
            ") { _additional { vectors { default } } } } }"
        )
        logger.info("weaviate.get_post_embedding.start", post_id=post_id)
        started = time.monotonic()
        async with observe(WEAVIATE_CALLS, op="get_post_embedding"):
            try:
                objs = await self._graphql(query, op="get_post_embedding")
            except SimilaritySearchFailed as exc:
                logger.error(
                    "weaviate.get_post_embedding.failed",
                    post_id=post_id,
                    elapsed_s=round(time.monotonic() - started, 3),
                    transient=exc.transient,
                    error=exc.message,
                )
                raise
        logger.info(
            "weaviate.get_post_embedding.done",
            post_id=post_id,
            elapsed_s=round(time.monotonic() - started, 3),
            chunks=len(objs),
        )

        vectors: list[list[float]] = []
        for obj in objs:
            named = (obj.get("_additional") or {}).get("vectors") or {}
            vec = _extract_vector(named.get("default"))
            if vec:
                vectors.append(vec)
        if not vectors:
            return None
        width = len(vectors[0])
        summed = [0.0] * width
        for vec in vectors:
            if len(vec) != width:
                continue
            for i, value in enumerate(vec):
                summed[i] += value
        n = float(len(vectors))
        return [v / n for v in summed]

    async def _search(
        self,
        query_embedding: Embedding,
        *,
        limit: int,
        kinds: list[ContentKind] | None = None,
        distance: float | None = None,
        exclude_source_id: str | None = None,
        place_id: str | None = None,
        lat: float | None = None,
        lon: float | None = None,
        radius_m: float | None = None,
    ) -> list[ScoredChunk]:
        await self._ensure_connected()
        operands: list[str] = []
        if kinds:
            arr = ", ".join(json.dumps(k) for k in kinds)
            operands.append(f'{{ path: ["kind"], operator: ContainsAny, valueText: [{arr}] }}')
        if exclude_source_id:
            operands.append(
                f'{{ path: ["sourceId"], operator: NotEqual, valueText: {json.dumps(exclude_source_id)} }}'
            )
        if place_id:
            operands.append(
                f'{{ path: ["placeId"], operator: Equal, valueText: {json.dumps(place_id)} }}'
            )
        if lat is not None and lon is not None and radius_m is not None:
            operands.append(
                '{ path: ["coordinate"], operator: WithinGeoRange, valueGeoRange: '
                f"{{ geoCoordinates: {{ latitude: {json.dumps(lat)}, longitude: {json.dumps(lon)} }}, "
                f"distance: {{ max: {json.dumps(radius_m)} }} }} }}"
            )
        where = _combine_where(operands)
        args = [f"limit: {limit}", _near_vector(query_embedding, distance)]
        if where is not None:
            args.append(f"where: {where}")
        query = (
            "{ Get { " + self._settings.class_name + "("
            + ", ".join(args)
            + ") { sourceId kind text segmentIdx category tags parentId parentText attachedPostId "
            "placeId coordinate { latitude longitude } _additional { distance } } } }"
        )
        op = "search_dedup" if distance is not None else "search"
        logger.info(
            "weaviate.search.start",
            op=op,
            limit=limit,
            kinds=kinds,
            place_id=place_id,
            distance=distance,
            exclude_source_id=exclude_source_id,
        )
        started = time.monotonic()
        async with observe(WEAVIATE_CALLS, op=op):
            try:
                objs = await self._graphql(query, op=op)
            except SimilaritySearchFailed as exc:
                logger.error(
                    "weaviate.search.failed",
                    op=op,
                    elapsed_s=round(time.monotonic() - started, 3),
                    transient=exc.transient,
                    error=exc.message,
                )
                raise
        logger.info(
            "weaviate.search.done",
            op=op,
            elapsed_s=round(time.monotonic() - started, 3),
            hits=len(objs),
        )

        out: list[ScoredChunk] = []
        for obj in objs:
            meta_distance = (obj.get("_additional") or {}).get("distance")
            score = 1.0 - float(meta_distance) if meta_distance is not None else 0.0
            chunk = Chunk(
                source_id=str(obj.get("sourceId", "")),
                kind=str(obj.get("kind", "post")),  # type: ignore[arg-type]
                text=str(obj.get("text", "")),
                segment_idx=int(obj.get("segmentIdx", 0) or 0),
                metadata=_metadata_from_props(obj),
            )
            out.append(ScoredChunk(chunk=chunk, score=score))
        return out

    async def list_by_place(self, place_id: str, limit: int) -> list[Chunk]:
        await self._ensure_connected()
        where = f'{{ path: ["placeId"], operator: Equal, valueText: {json.dumps(place_id)} }}'
        query = (
            "{ Get { " + self._settings.class_name + "("
            f"limit: {limit}, where: {where}"
            ") { sourceId kind text segmentIdx category tags parentId parentText attachedPostId "
            "placeId coordinate { latitude longitude } } } }"
        )
        logger.info("weaviate.list_by_place.start", place_id=place_id, limit=limit)
        started = time.monotonic()
        async with observe(WEAVIATE_CALLS, op="list_by_place"):
            try:
                objs = await self._graphql(query, op="list_by_place")
            except SimilaritySearchFailed as exc:
                logger.error(
                    "weaviate.list_by_place.failed",
                    place_id=place_id,
                    elapsed_s=round(time.monotonic() - started, 3),
                    transient=exc.transient,
                    error=exc.message,
                )
                raise
        logger.info(
            "weaviate.list_by_place.done",
            place_id=place_id,
            elapsed_s=round(time.monotonic() - started, 3),
            hits=len(objs),
        )
        out: list[Chunk] = []
        for obj in objs:
            out.append(
                Chunk(
                    source_id=str(obj.get("sourceId", "")),
                    kind=str(obj.get("kind", "post")),  # type: ignore[arg-type]
                    text=str(obj.get("text", "")),
                    segment_idx=int(obj.get("segmentIdx", 0) or 0),
                    metadata=_metadata_from_props(obj),
                )
            )
        return out


def _extract_vector(stored: Any) -> list[float]:
    if not stored:
        return []
    if isinstance(stored, list):
        return [float(v) for v in stored]
    if isinstance(stored, dict):
        for value in stored.values():
            if isinstance(value, list) and value and isinstance(value[0], (int, float)):
                return [float(v) for v in value]
    return []


def _metadata_from_props(props: dict[str, Any]) -> dict[str, str]:
    meta: dict[str, str] = {}
    for src, dst in (
        ("category", "category"),
        ("parentId", "parent_id"),
        ("parentText", "parent_text"),
        ("attachedPostId", "attached_post_id"),
        ("placeId", "place_id"),
    ):
        value = props.get(src)
        if value:
            meta[dst] = str(value)
    tags = props.get("tags") or []
    if tags:
        meta["tags"] = ",".join(str(t) for t in tags)
    coord = props.get("coordinate")
    if isinstance(coord, dict):
        if "latitude" in coord:
            meta["lat"] = str(coord["latitude"])
        if "longitude" in coord:
            meta["lon"] = str(coord["longitude"])
    return meta
