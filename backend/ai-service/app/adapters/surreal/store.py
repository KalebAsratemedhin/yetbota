"""SurrealDB store for post embeddings and similar-to graph edges."""

from __future__ import annotations

import logging
import math
from datetime import datetime, timezone
from typing import Any, Sequence
from uuid import UUID

from surrealdb import AsyncSurreal, RecordID

from app.adapters.surreal.client import SurrealConfig
from app.domain.post_doc import PostDocument
from app.domain.similarity import SimilarHit

logger = logging.getLogger(__name__)

TABLE = "post_embedding"
EDGE = "similar"


def _schema_surql(dimensions: int) -> str:
    return f"""
DEFINE TABLE IF NOT EXISTS {TABLE} SCHEMALESS;
DEFINE FIELD IF NOT EXISTS post_id ON {TABLE} TYPE string;
DEFINE FIELD IF NOT EXISTS author_id ON {TABLE} TYPE string;
DEFINE FIELD IF NOT EXISTS title ON {TABLE} TYPE string;
DEFINE FIELD IF NOT EXISTS text ON {TABLE} TYPE string;
DEFINE FIELD IF NOT EXISTS lat ON {TABLE} TYPE float;
DEFINE FIELD IF NOT EXISTS lng ON {TABLE} TYPE float;
DEFINE FIELD IF NOT EXISTS embedding ON {TABLE} TYPE array<float>;
DEFINE FIELD IF NOT EXISTS model ON {TABLE} TYPE string;
DEFINE FIELD IF NOT EXISTS updated_at ON {TABLE} TYPE datetime;
DEFINE INDEX IF NOT EXISTS {TABLE}_hnsw ON {TABLE}
    FIELDS embedding HNSW DIMENSION {dimensions} DIST COSINE TYPE F32;
DEFINE TABLE IF NOT EXISTS {EDGE} TYPE RELATION IN {TABLE} OUT {TABLE} SCHEMALESS;
DEFINE FIELD IF NOT EXISTS score ON {EDGE} TYPE float;
DEFINE FIELD IF NOT EXISTS updated_at ON {EDGE} TYPE option<datetime>;
"""


def _cosine(a: Sequence[float], b: Sequence[float]) -> float:
    if len(a) != len(b) or not a:
        return 0.0
    dot = 0.0
    na = 0.0
    nb = 0.0
    for x, y in zip(a, b, strict=True):
        dot += x * y
        na += x * x
        nb += y * y
    if na <= 0 or nb <= 0:
        return 0.0
    return float(dot / (math.sqrt(na) * math.sqrt(nb)))


def _unwrap_query_rows(result: Any) -> list[dict[str, Any]]:
    """Normalize Surreal Python query payloads into a list of row dicts."""
    if result is None:
        return []
    if isinstance(result, list):
        if not result:
            return []
        first = result[0]
        # Newer SDK: list of result envelopes with .get("result")
        if isinstance(first, dict) and "result" in first:
            rows = first.get("result") or []
            return [r for r in rows if isinstance(r, dict)]
        if isinstance(first, dict):
            return [r for r in result if isinstance(r, dict)]
    if isinstance(result, dict):
        if "result" in result:
            rows = result.get("result") or []
            return [r for r in rows if isinstance(r, dict)]
        return [result]
    return []


class SurrealPostEmbeddingStore:
    def __init__(self, config: SurrealConfig, embedding_dimensions: int) -> None:
        self._config = config
        self._dimensions = embedding_dimensions
        self._db: Any | None = None

    async def connect(self) -> None:
        db = AsyncSurreal(self._config.url)
        await db.connect()
        await db.signin(
            {
                "username": self._config.username,
                "password": self._config.password,
            }
        )
        await db.use(self._config.namespace, self._config.database)
        self._db = db
        await self.ensure_schema()
        logger.info(
            "surreal_connected ns=%s db=%s table=%s edge=%s dim=%s",
            self._config.namespace,
            self._config.database,
            TABLE,
            EDGE,
            self._dimensions,
        )

    async def ensure_schema(self) -> None:
        if self._db is None:
            raise RuntimeError("surreal not connected")
        try:
            await self._db.query(_schema_surql(self._dimensions))
        except Exception:
            logger.warning("surreal_schema_if_not_exists_failed; retrying without IF NOT EXISTS")
            bare = f"""
DEFINE TABLE {TABLE} SCHEMALESS;
DEFINE FIELD post_id ON {TABLE} TYPE string;
DEFINE FIELD author_id ON {TABLE} TYPE string;
DEFINE FIELD title ON {TABLE} TYPE string;
DEFINE FIELD text ON {TABLE} TYPE string;
DEFINE FIELD lat ON {TABLE} TYPE float;
DEFINE FIELD lng ON {TABLE} TYPE float;
DEFINE FIELD embedding ON {TABLE} TYPE array<float>;
DEFINE FIELD model ON {TABLE} TYPE string;
DEFINE FIELD updated_at ON {TABLE} TYPE datetime;
DEFINE INDEX {TABLE}_hnsw ON {TABLE}
    FIELDS embedding HNSW DIMENSION {self._dimensions} DIST COSINE TYPE F32;
DEFINE TABLE {EDGE} TYPE RELATION IN {TABLE} OUT {TABLE} SCHEMALESS;
DEFINE FIELD score ON {EDGE} TYPE float;
DEFINE FIELD updated_at ON {EDGE} TYPE option<datetime>;
"""
            try:
                await self._db.query(bare)
            except Exception:
                logger.exception("surreal_schema_define_failed")
                raise

    async def close(self) -> None:
        if self._db is not None:
            await self._db.close()
            self._db = None

    async def upsert_embedding(
        self,
        post: PostDocument,
        text: str,
        embedding: Sequence[float],
        model: str,
    ) -> None:
        if self._db is None:
            raise RuntimeError("surreal not connected")
        if len(embedding) != self._dimensions:
            raise ValueError(
                f"embedding length {len(embedding)} != schema dim {self._dimensions}"
            )
        rid = RecordID(TABLE, str(post.id))
        payload = {
            "post_id": str(post.id),
            "author_id": str(post.author_id),
            "title": post.title,
            "text": text,
            "lat": post.lat,
            "lng": post.lng,
            "embedding": list(embedding),
            "model": model,
            "updated_at": datetime.now(timezone.utc).isoformat(),
        }
        await self._db.upsert(rid, payload)
        logger.info("post_embedding_upserted post_id=%s dim=%s", post.id, len(embedding))

    async def find_similar(
        self,
        embedding: Sequence[float],
        exclude_post_id: UUID,
        *,
        top_k: int,
        min_score: float,
    ) -> list[SimilarHit]:
        """Rank other post_embedding rows by cosine similarity (in-process).

        Vectors stay in Surreal; for v1 we score candidates in Python so results
        are exact and work even when HNSW KNN syntax differs across Surreal versions.
        """
        if self._db is None:
            raise RuntimeError("surreal not connected")
        if top_k <= 0:
            return []
        raw = await self._db.query(
            f"""
            SELECT post_id, embedding FROM {TABLE}
            WHERE post_id != $exclude
            """,
            {"exclude": str(exclude_post_id)},
        )
        rows = _unwrap_query_rows(raw)
        scored: list[SimilarHit] = []
        for row in rows:
            pid_raw = row.get("post_id")
            emb = row.get("embedding")
            if pid_raw is None or not isinstance(emb, (list, tuple)):
                continue
            try:
                pid = UUID(str(pid_raw))
            except ValueError:
                continue
            score = _cosine(embedding, emb)
            if score < min_score:
                continue
            scored.append(SimilarHit(post_id=pid, score=score))
        scored.sort(key=lambda h: (-h.score, str(h.post_id)))
        return scored[:top_k]

    async def replace_similar_edges(
        self,
        from_post_id: UUID,
        hits: Sequence[SimilarHit],
    ) -> None:
        """Replace outbound `similar` edges from this post's embedding record."""
        if self._db is None:
            raise RuntimeError("surreal not connected")
        frm = RecordID(TABLE, str(from_post_id))
        # Drop previous outbound edges, then relate top hits.
        await self._db.query(
            "DELETE $from->similar;",
            {"from": frm},
        )
        for hit in hits:
            to = RecordID(TABLE, str(hit.post_id))
            await self._db.query(
                """
                RELATE $from->similar->$to SET
                    score = $score,
                    updated_at = time::now();
                """,
                {"from": frm, "to": to, "score": float(hit.score)},
            )
        logger.info(
            "similar_edges_replaced post_id=%s count=%s",
            from_post_id,
            len(hits),
        )
