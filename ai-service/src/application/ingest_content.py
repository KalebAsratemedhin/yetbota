from datetime import UTC, datetime

from application.errors import AIServiceError, MetadataError
from domain.entities import (
    Embedding,
    IngestRequest,
    IngestResult,
    ScoredChunk,
    Verdict,
)
from domain.ports import Chunker, Embedder, SimilarityGraph, VectorStore


def _mean(vectors: list[Embedding]) -> Embedding:
    if not vectors:
        return []
    width = len(vectors[0])
    summed = [0.0] * width
    for vec in vectors:
        if len(vec) != width:
            raise ValueError("embeddings have inconsistent dimensions")
        for i, value in enumerate(vec):
            summed[i] += value
    n = float(len(vectors))
    return [v / n for v in summed]


def _build_metadata(req: IngestRequest) -> dict[str, str]:
    meta: dict[str, str] = {}
    if req.category:
        meta["category"] = req.category
    if req.parent_id:
        meta["parent_id"] = req.parent_id
    if req.attached_post_id:
        meta["attached_post_id"] = req.attached_post_id
    if req.place_id:
        meta["place_id"] = req.place_id
    if req.tags:
        meta["tags"] = ",".join(req.tags)
    if req.latitude is not None and req.longitude is not None:
        meta["lat"] = str(req.latitude)
        meta["lon"] = str(req.longitude)
    if req.context:
        meta["parent_text"] = req.context
    return meta


def _validate(req: IngestRequest) -> None:
    if not req.text or not req.text.strip():
        raise MetadataError("text is empty")
    if not req.content_id:
        raise MetadataError("content_id is empty")
    if not req.user_id:
        raise MetadataError("user_id is empty")
    if req.kind == "answer" and not req.parent_id:
        raise MetadataError("answer requires parent_id")


class IngestContent:
    def __init__(
        self,
        *,
        chunker: Chunker,
        embedder: Embedder,
        vector_store: VectorStore,
        distance_threshold: float,
        spatial_radius_m: float | None = None,
        similarity_graph: SimilarityGraph | None = None,
        similarity_top_n: int = 10,
        similarity_oversample: int = 5,
        logger: object | None = None,
    ) -> None:
        self._chunker = chunker
        self._embedder = embedder
        self._vector_store = vector_store
        self._distance_threshold = distance_threshold
        self._spatial_radius_m = spatial_radius_m
        self._similarity_graph = similarity_graph
        self._similarity_top_n = similarity_top_n
        self._similarity_oversample = similarity_oversample
        self._logger = logger

    async def execute(self, req: IngestRequest) -> IngestResult:
        _validate(req)

        chunks = self._chunker.chunk(
            source_id=req.content_id,
            kind=req.kind,
            text=req.text,
            metadata=_build_metadata(req),
        )
        if not chunks:
            raise MetadataError("text produced no chunks after normalization")

        # For answers, prepend the question (req.context) to each chunk for the
        # embedding only, so the answer's vector carries its question's context.
        # The stored chunk text stays the answer itself.
        if req.kind == "answer" and req.context:
            embed_texts = [f"{req.context}\n\n{c.text}" for c in chunks]
        else:
            embed_texts = [c.text for c in chunks]

        embeddings = await self._embedder.embed_batch(
            embed_texts,
            task_type="RETRIEVAL_DOCUMENT",
        )

        if req.kind == "post":
            mean_vec = _mean(embeddings)
            duplicate = await self._dedup_check(
                mean_vec,
                exclude_source_id=req.content_id,
                lat=req.latitude,
                lon=req.longitude,
            )
            if duplicate is not None:
                original_id = duplicate.chunk.source_id
                await self._record_duplicate_edge(req.content_id, original_id, duplicate.score)
                return _result(
                    req,
                    verdict="duplicate",
                    duplicate_of=original_id,
                )
        else:
            mean_vec = []

        await self._vector_store.upsert(chunks, embeddings)

        if req.kind == "post":
            await self._maintain_similarity_graph(req.content_id, mean_vec)

        verdict: Verdict = "unique" if req.kind == "post" else "indexed"
        return _result(req, verdict=verdict)

    async def _dedup_check(
        self,
        mean_vec: Embedding,
        *,
        exclude_source_id: str,
        lat: float | None = None,
        lon: float | None = None,
    ) -> ScoredChunk | None:
        hits = await self._vector_store.search_dedup(
            mean_vec,
            distance_threshold=self._distance_threshold,
            limit=1,
            exclude_source_id=exclude_source_id,
            lat=lat,
            lon=lon,
            radius_m=self._spatial_radius_m,
        )
        return hits[0] if hits else None

    async def _record_duplicate_edge(
        self, duplicate_id: str, original_id: str, score: float
    ) -> None:
        if self._similarity_graph is None:
            return
        try:
            await self._similarity_graph.mark_duplicate(duplicate_id, original_id, score)
        except AIServiceError as exc:
            if self._logger is not None:
                self._logger.warning(  # type: ignore[attr-defined]
                    "ingest.duplicate_edge_failed",
                    duplicate_id=duplicate_id,
                    original_id=original_id,
                    error_code=exc.code,
                )

    async def _maintain_similarity_graph(self, post_id: str, mean_vec: Embedding) -> None:
        if self._similarity_graph is None or self._similarity_top_n <= 0:
            return
        try:
            similar = await self._vector_store.find_similar_posts(
                mean_vec,
                exclude_post_id=post_id,
                top_n=self._similarity_top_n,
                oversample=self._similarity_oversample,
            )
            await self._similarity_graph.update_similar(post_id, similar)
            for neighbor in similar:
                neighbor_vec = await self._vector_store.get_post_embedding(neighbor.post_id)
                if neighbor_vec is None:
                    continue
                neighbor_similar = await self._vector_store.find_similar_posts(
                    neighbor_vec,
                    exclude_post_id=neighbor.post_id,
                    top_n=self._similarity_top_n,
                    oversample=self._similarity_oversample,
                )
                await self._similarity_graph.update_similar(neighbor.post_id, neighbor_similar)
        except AIServiceError as exc:
            if self._logger is not None:
                self._logger.warning(  # type: ignore[attr-defined]
                    "ingest.similarity_update_failed",
                    post_id=post_id,
                    error_code=exc.code,
                )


def _result(
    req: IngestRequest,
    *,
    verdict: Verdict,
    duplicate_of: str | None = None,
) -> IngestResult:
    return IngestResult(
        content_id=req.content_id,
        kind=req.kind,
        verdict=verdict,
        duplicate_of=duplicate_of,
        error_code=None,
        processed_at=datetime.now(UTC),
    )


__all__ = ["IngestContent"]
