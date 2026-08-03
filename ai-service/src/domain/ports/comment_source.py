from typing import Protocol, runtime_checkable

from domain.entities import Comment


@runtime_checkable
class CommentSource(Protocol):
    async def read_comment(self, comment_id: str) -> Comment | None: ...
