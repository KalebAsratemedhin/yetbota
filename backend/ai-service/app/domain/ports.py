"""Domain ports — implement in adapters when ingest/similarity land.

Example shape (not enforced yet):

    class EmbeddingProvider(Protocol):
        async def embed(self, texts: list[str]) -> list[list[float]]: ...

    class SimilarityGraph(Protocol):
        async def upsert_similar(self, post_id: str, similar_ids: list[str], scores: list[float]) -> None: ...
"""

from typing import Protocol


class HealthPort(Protocol):
    """Placeholder so the ports module is importable."""

    async def ping(self) -> bool: ...
