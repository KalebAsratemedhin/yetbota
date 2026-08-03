import asyncpg

from application.errors import PostNotFound, PostSourceUnavailable
from domain.entities import Post
from infrastructure.config.settings import PostgresSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)

_READ_POST_SQL = """
SELECT id::text AS id, title, description, user_id::text AS user_id, is_question, tags,
       ST_Y(location::geometry) AS latitude, ST_X(location::geometry) AS longitude
FROM posts
WHERE id = $1::uuid
"""


class PostgresPostSource:
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
            raise PostSourceUnavailable(
                f"failed to create postgres pool: {exc}", cause=exc
            ) from exc
        logger.info("postgres.connected")

    async def close(self) -> None:
        if self._pool is not None:
            await self._pool.close()
            self._pool = None

    async def read_post(self, post_id: str) -> Post | None:
        assert self._pool is not None
        try:
            row = await self._pool.fetchrow(_READ_POST_SQL, post_id)
        except (asyncpg.PostgresError, OSError) as exc:
            raise PostSourceUnavailable(
                f"postgres read_post failed: {exc}", cause=exc
            ) from exc
        if row is None:
            raise PostNotFound(f"post {post_id} not found")
        return Post(
            id=row["id"],
            title=row["title"] or "",
            description=row["description"] or "",
            user_id=row["user_id"],
            is_question=bool(row["is_question"]),
            tags=list(row["tags"] or []),
            latitude=row["latitude"],
            longitude=row["longitude"],
        )
