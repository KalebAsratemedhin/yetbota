from typing import Protocol, runtime_checkable

from domain.entities import SimilarPost


@runtime_checkable
class SimilarityGraph(Protocol):
    async def update_similar(self, post_id: str, similar: list[SimilarPost]) -> None: ...

    async def mark_duplicate(
        self, duplicate_post_id: str, original_post_id: str, score: float
    ) -> None: ...

    async def delete_post(self, post_id: str) -> None: ...

    async def close(self) -> None: ...
