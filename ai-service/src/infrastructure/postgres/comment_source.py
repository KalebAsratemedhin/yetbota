import asyncpg

from application.errors import CommentNotFound, CommentSourceUnavailable
from domain.entities import Comment
from infrastructure.config.settings import PostgresSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)

_READ_COMMENT_SQL = """
SELECT id::text AS id, content AS body, user_id::text AS user_id,
       post_id::text AS post_id, is_answer, comment_id::text AS comment_id
FROM comments
WHERE id = $1::uuid
"""


class PostgresCommentSource:
    def __init__(self, settings: PostgresSettings) -> None:
        self._settings = settings
        self._pool: asyncpg.Pool | None = None

    async def connect(self) -> None:
        try:
            self._pool = await asyncpg.create_pool(
                dsn=self._settings.dsn,
                min_size=self._settings.min_pool_size,
                max_size=self._settings.max_pool_size,
            )
        except (asyncpg.PostgresError, OSError) as exc:
            raise CommentSourceUnavailable(
                f"failed to create postgres pool: {exc}", cause=exc
            ) from exc
        logger.info("postgres.comment_source.connected")

    async def close(self) -> None:
        if self._pool is not None:
            await self._pool.close()
            self._pool = None

    async def read_comment(self, comment_id: str) -> Comment | None:
        assert self._pool is not None
        try:
            row = await self._pool.fetchrow(_READ_COMMENT_SQL, comment_id)
        except (asyncpg.PostgresError, OSError) as exc:
            raise CommentSourceUnavailable(
                f"postgres read_comment failed: {exc}", cause=exc
            ) from exc
        if row is None:
            raise CommentNotFound(f"comment {comment_id} not found")
        return Comment(
            id=row["id"],
            body=row["body"] or "",
            user_id=row["user_id"],
            post_id=row["post_id"],
            is_answer=bool(row["is_answer"]),
            comment_id=row["comment_id"],
        )
