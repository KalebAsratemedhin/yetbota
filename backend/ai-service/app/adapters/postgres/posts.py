"""Read posts from shared Postgres (content-service tables)."""

from __future__ import annotations

from uuid import UUID

from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker

from app.domain.post_doc import PostDocument

_GET_POST = text(
    """
    SELECT
        p.id,
        p.author_id,
        p.title,
        p.caption,
        p.lat,
        p.lng,
        coalesce(string_agg(h.tag, ' ' ORDER BY h.tag), '') AS hashtags
    FROM posts p
    LEFT JOIN post_hashtags ph ON ph.post_id = p.id
    LEFT JOIN hashtags h ON h.id = ph.hashtag_id
    WHERE p.id = CAST(:post_id AS uuid)
    GROUP BY p.id, p.author_id, p.title, p.caption, p.lat, p.lng
    """
)


class PostgresPostReader:
    def __init__(self, sessions: async_sessionmaker[AsyncSession]) -> None:
        self._sessions = sessions

    async def get_post(self, post_id: UUID) -> PostDocument | None:
        async with self._sessions() as session:
            row = (
                await session.execute(_GET_POST, {"post_id": str(post_id)})
            ).mappings().first()
        if row is None:
            return None
        return PostDocument(
            id=UUID(str(row["id"])),
            author_id=UUID(str(row["author_id"])),
            title=row["title"] or "",
            caption=row["caption"] or "",
            lat=float(row["lat"]),
            lng=float(row["lng"]),
            hashtags=row["hashtags"] or "",
        )
