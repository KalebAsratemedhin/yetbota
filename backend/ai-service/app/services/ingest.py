"""Content ingest: Postgres → embed → Surreal (post similar + typed content index)."""

from __future__ import annotations

import logging
from typing import Protocol
from uuid import UUID

from app.domain.content_doc import ContentDocument
from app.domain.ports import ContentReader, EmbeddingProvider, PostEmbeddingStore
from app.domain.post_doc import PostDocument

logger = logging.getLogger(__name__)


class PlaceAttachHook(Protocol):
    async def after_post_ingest(
        self, post: PostDocument, embedding: list[float]
    ) -> None: ...


class IngestService:
    def __init__(
        self,
        reader: ContentReader,
        embedder: EmbeddingProvider,
        store: PostEmbeddingStore,
        *,
        similarity_top_k: int = 10,
        similarity_min_score: float = 0.55,
        place_attach: PlaceAttachHook | None = None,
    ) -> None:
        self._reader = reader
        self._embedder = embedder
        self._store = store
        self._top_k = similarity_top_k
        self._min_score = similarity_min_score
        self._place_attach = place_attach

    async def handle_event(self, event: str, payload: dict) -> bool:
        """Dispatch thin domain events to the right ingest path."""
        if event == "post.created":
            return await self.ingest_post(UUID(str(payload["post_id"])))
        if event == "post.updated":
            return await self.ingest_post(UUID(str(payload["post_id"])))
        if event == "comment.created":
            return await self.ingest_comment(UUID(str(payload["comment_id"])))
        if event == "question.created":
            return await self.ingest_question(UUID(str(payload["question_id"])))
        if event == "question.updated":
            return await self.ingest_question(UUID(str(payload["question_id"])))
        if event == "answer.created":
            return await self.ingest_answer(UUID(str(payload["answer_id"])))
        if event == "answer.updated":
            return await self.ingest_answer(UUID(str(payload["answer_id"])))
        logger.warning("ignored_event type=%s", event)
        return False

    async def ingest_post(self, post_id: UUID) -> bool:
        post = await self._reader.get_post(post_id)
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
        content = await self._reader.content_for_post(post_id)
        if content is not None and content.text.strip():
            await self._store.upsert_content(
                content, embedding, self._embedder.model_name
            )
        logger.info(
            "ingest_post_ok post_id=%s model=%s similar=%s",
            post_id,
            self._embedder.model_name,
            len(hits),
        )
        if self._place_attach is not None:
            await self._place_attach.after_post_ingest(post, embedding)
        return True

    async def ingest_comment(self, comment_id: UUID) -> bool:
        return await self._ingest_content_only(
            await self._reader.content_for_comment(comment_id),
            label="comment",
            entity_id=comment_id,
        )

    async def ingest_question(self, question_id: UUID) -> bool:
        return await self._ingest_content_only(
            await self._reader.content_for_question(question_id),
            label="question",
            entity_id=question_id,
        )

    async def ingest_answer(self, answer_id: UUID) -> bool:
        return await self._ingest_content_only(
            await self._reader.content_for_answer(answer_id),
            label="answer",
            entity_id=answer_id,
        )

    async def _ingest_content_only(
        self,
        doc: ContentDocument | None,
        *,
        label: str,
        entity_id: UUID,
    ) -> bool:
        if doc is None:
            logger.warning("ingest_%s_not_found id=%s", label, entity_id)
            return False
        if not doc.text.strip():
            logger.warning("ingest_%s_empty_text id=%s", label, entity_id)
            return False
        embedding = await self._embedder.embed(doc.text)
        await self._store.upsert_content(doc, embedding, self._embedder.model_name)
        logger.info(
            "ingest_%s_ok id=%s model=%s",
            label,
            entity_id,
            self._embedder.model_name,
        )
        return True
