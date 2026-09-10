"""Post ingest: Postgres → embed → Surreal vector + similar edges."""

from __future__ import annotations

import logging
from uuid import UUID

from app.domain.ports import EmbeddingProvider, PostEmbeddingStore, PostReader

logger = logging.getLogger(__name__)


class IngestService:
    def __init__(
        self,
        posts: PostReader,
        embedder: EmbeddingProvider,
        store: PostEmbeddingStore,
        *,
        similarity_top_k: int = 10,
        similarity_min_score: float = 0.55,
    ) -> None:
        self._posts = posts
        self._embedder = embedder
        self._store = store
        self._top_k = similarity_top_k
        self._min_score = similarity_min_score

    async def ingest_post(self, post_id: UUID) -> bool:
        post = await self._posts.get_post(post_id)
        if post is None:
            logger.warning("ingest_post_not_found post_id=%s", post_id)
            return False
        text = post.embed_text()
        if not text.strip():
            logger.warning("ingest_post_empty_text post_id=%s", post_id)
            return False
        embedding = await self._embedder.embed(text)
        await self._store.upsert_embedding(
            post, text, embedding, self._embedder.model_name
        )
        hits = await self._store.find_similar(
            embedding,
            post.id,
            top_k=self._top_k,
            min_score=self._min_score,
        )
        await self._store.replace_similar_edges(post.id, hits)
        logger.info(
            "ingest_post_ok post_id=%s model=%s similar=%s",
            post_id,
            self._embedder.model_name,
            len(hits),
        )
        return True
