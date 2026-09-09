"""Ports used by ingest (implemented in adapters)."""

from __future__ import annotations

from typing import Protocol, Sequence
from uuid import UUID

from app.domain.post_doc import PostDocument


class PostReader(Protocol):
    async def get_post(self, post_id: UUID) -> PostDocument | None: ...


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
