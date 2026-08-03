from typing import Protocol, runtime_checkable

from domain.entities import Post


@runtime_checkable
class PostSource(Protocol):
    async def read_post(self, post_id: str) -> Post | None: ...
