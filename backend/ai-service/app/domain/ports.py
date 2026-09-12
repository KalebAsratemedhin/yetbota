"""Ports used by ingest / similarity (implemented in adapters)."""

from __future__ import annotations

from typing import Protocol, Sequence
from uuid import UUID

from app.domain.content_doc import ContentDocument
from app.domain.post_doc import PostDocument
from app.domain.similarity import SimilarHit


class ContentReader(Protocol):
    async def get_post(self, post_id: UUID) -> PostDocument | None: ...

    async def content_for_post(self, post_id: UUID) -> ContentDocument | None: ...

    async def content_for_comment(self, comment_id: UUID) -> ContentDocument | None: ...

    async def content_for_question(self, question_id: UUID) -> ContentDocument | None: ...

    async def content_for_answer(self, answer_id: UUID) -> ContentDocument | None: ...


# Back-compat alias used by older post-only wiring.
PostReader = ContentReader


class EmbeddingProvider(Protocol):
    @property
    def model_name(self) -> str: ...

    @property
    def dimensions(self) -> int: ...

    async def embed(self, text: str) -> list[float]: ...


class PostEmbeddingStore(Protocol):
    async def upsert_embedding(
        self,
        post: PostDocument,
        text: str,
        embedding: Sequence[float],
        model: str,
    ) -> None: ...

    async def upsert_content(
        self,
        doc: ContentDocument,
        embedding: Sequence[float],
        model: str,
    ) -> None: ...

    async def find_similar(
        self,
        embedding: Sequence[float],
        exclude_post_id: UUID,
        *,
        top_k: int,
        min_score: float,
    ) -> list[SimilarHit]: ...

    async def replace_similar_edges(
        self,
        from_post_id: UUID,
        hits: Sequence[SimilarHit],
    ) -> None: ...
