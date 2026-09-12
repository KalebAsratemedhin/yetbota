"""Load posts, comments, questions, answers for typed content ingest."""

from __future__ import annotations

from uuid import UUID

from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker

from app.domain.content_doc import ContentDocument
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

_GET_COMMENT = text(
    """
    SELECT
        c.id,
        c.post_id,
        c.author_id,
        c.body,
        c.parent_id,
        p.title AS post_title
    FROM post_comments c
    JOIN posts p ON p.id = c.post_id
    WHERE c.id = CAST(:comment_id AS uuid)
    """
)

_GET_QUESTION = text(
    """
    SELECT
        q.id,
        q.author_id,
        q.body,
        q.post_id,
        q.lat,
        q.lng,
        p.title AS post_title
    FROM questions q
    LEFT JOIN posts p ON p.id = q.post_id
    WHERE q.id = CAST(:question_id AS uuid)
    """
)

_GET_ANSWER = text(
    """
    SELECT
        a.id,
        a.question_id,
        a.author_id,
        a.body,
        q.body AS question_body,
        q.post_id
    FROM answers a
    JOIN questions q ON q.id = a.question_id
    WHERE a.id = CAST(:answer_id AS uuid)
    """
)


def _join_parts(*parts: str) -> str:
    return "\n".join(p.strip() for p in parts if p and p.strip())


class PostgresContentReader:
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

    async def content_for_post(self, post_id: UUID) -> ContentDocument | None:
        post = await self.get_post(post_id)
        if post is None:
            return None
        return ContentDocument(
            kind="post",
            entity_id=post.id,
            author_id=post.author_id,
            text=post.embed_text(),
            post_id=post.id,
            title=post.title,
            lat=post.lat,
            lng=post.lng,
        )

    async def content_for_comment(self, comment_id: UUID) -> ContentDocument | None:
        async with self._sessions() as session:
            row = (
                await session.execute(_GET_COMMENT, {"comment_id": str(comment_id)})
            ).mappings().first()
        if row is None:
            return None
        post_id = UUID(str(row["post_id"]))
        parent_raw = row["parent_id"]
        parent_id = UUID(str(parent_raw)) if parent_raw is not None else post_id
        body = row["body"] or ""
        title = row["post_title"] or ""
        text_body = _join_parts(f"Post: {title}" if title else "", f"Comment: {body}")
        return ContentDocument(
            kind="comment",
            entity_id=UUID(str(row["id"])),
            author_id=UUID(str(row["author_id"])),
            text=text_body,
            parent_id=parent_id,
            post_id=post_id,
            title=title,
        )

    async def content_for_question(self, question_id: UUID) -> ContentDocument | None:
        async with self._sessions() as session:
            row = (
                await session.execute(_GET_QUESTION, {"question_id": str(question_id)})
            ).mappings().first()
        if row is None:
            return None
        post_id = UUID(str(row["post_id"])) if row["post_id"] is not None else None
        place = row["post_title"] or ""
        body = row["body"] or ""
        text_body = _join_parts(f"Place: {place}" if place else "", body)
        lat = float(row["lat"]) if row["lat"] is not None else None
        lng = float(row["lng"]) if row["lng"] is not None else None
        return ContentDocument(
            kind="question",
            entity_id=UUID(str(row["id"])),
            author_id=UUID(str(row["author_id"])),
            text=text_body,
            parent_id=post_id,
            post_id=post_id,
            title=place,
            lat=lat,
            lng=lng,
        )

    async def content_for_answer(self, answer_id: UUID) -> ContentDocument | None:
        async with self._sessions() as session:
            row = (
                await session.execute(_GET_ANSWER, {"answer_id": str(answer_id)})
            ).mappings().first()
        if row is None:
            return None
        qid = UUID(str(row["question_id"]))
        post_id = UUID(str(row["post_id"])) if row["post_id"] is not None else None
        q_body = row["question_body"] or ""
        a_body = row["body"] or ""
        text_body = _join_parts(f"Q:\n{q_body}", f"A:\n{a_body}")
        return ContentDocument(
            kind="answer",
            entity_id=UUID(str(row["id"])),
            author_id=UUID(str(row["author_id"])),
            text=text_body,
            parent_id=qid,
            post_id=post_id,
        )
