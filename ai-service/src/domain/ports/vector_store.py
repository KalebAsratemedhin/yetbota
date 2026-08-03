from typing import Protocol, runtime_checkable

from domain.entities import Chunk, ContentKind, Embedding, ScoredChunk, SimilarPost


@runtime_checkable
class VectorStore(Protocol):
    async def upsert(self, chunks: list[Chunk], embeddings: list[Embedding]) -> None: ...

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
    ) -> list[ScoredChunk]: ...

    async def list_by_place(self, place_id: str, limit: int) -> list[Chunk]: ...

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
    ) -> list[ScoredChunk]: ...

    async def find_similar_posts(
        self,
        query_embedding: Embedding,
        *,
        exclude_post_id: str,
        top_n: int,
        oversample: int,
    ) -> list[SimilarPost]: ...

    async def get_post_embedding(self, post_id: str) -> Embedding | None: ...

    async def delete_by_source(self, source_id: str, kind: ContentKind) -> None: ...
