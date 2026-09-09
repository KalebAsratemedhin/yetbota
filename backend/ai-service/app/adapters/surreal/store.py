"""SurrealDB store for post embeddings (vectors live here)."""

from __future__ import annotations

import logging
from datetime import datetime, timezone
from typing import Any, Sequence

from surrealdb import AsyncSurreal, RecordID

from app.adapters.surreal.client import SurrealConfig
from app.domain.post_doc import PostDocument

logger = logging.getLogger(__name__)

TABLE = "post_embedding"


def _schema_surql(dimensions: int) -> str:
    # SCHEMALESS + HNSW for cosine ANN (dimension must match embedding model).
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
"""


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
            "surreal_connected ns=%s db=%s table=%s dim=%s",
            self._config.namespace,
            self._config.database,
            TABLE,
            self._dimensions,
        )

    async def ensure_schema(self) -> None:
        if self._db is None:
            raise RuntimeError("surreal not connected")
        # Prefer IF NOT EXISTS (Surreal 2+); fall back to bare DEFINE and ignore conflicts.
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
