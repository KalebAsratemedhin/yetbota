import json

import asyncpg

from application.errors import PostSourceUnavailable
from domain.entities import Chat, Citation, Message, MessageRole
from infrastructure.config.settings import PostgresSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)

_DDL = """
CREATE TABLE IF NOT EXISTS chats (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL,
    title       text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS chats_user_idx ON chats (user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS chat_messages (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id     uuid NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    role        text NOT NULL,
    content     text NOT NULL,
    citations   jsonb NOT NULL DEFAULT '[]',
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS chat_messages_chat_idx ON chat_messages (chat_id, created_at);
"""

_INSERT_CHAT = """
INSERT INTO chats (user_id, title)
VALUES ($1::uuid, $2)
RETURNING id::text, user_id::text, title, created_at, updated_at
"""

_LIST_CHATS = """
SELECT id::text, user_id::text, title, created_at, updated_at
FROM chats
WHERE user_id = $1::uuid
ORDER BY updated_at DESC
LIMIT $2
"""

_GET_CHAT = """
SELECT id::text, user_id::text, title, created_at, updated_at
FROM chats
WHERE id = $1::uuid
"""

_DELETE_CHAT = "DELETE FROM chats WHERE id = $1::uuid RETURNING id"

_SET_TITLE = "UPDATE chats SET title = $2, updated_at = now() WHERE id = $1::uuid"

_INSERT_MESSAGE = """
INSERT INTO chat_messages (chat_id, role, content, citations)
VALUES ($1::uuid, $2, $3, $4::jsonb)
RETURNING id::text, chat_id::text, role, content, citations, created_at
"""

_TOUCH_CHAT = "UPDATE chats SET updated_at = now() WHERE id = $1::uuid"

_LIST_MESSAGES = """
SELECT id::text, chat_id::text, role, content, citations, created_at
FROM chat_messages
WHERE chat_id = $1::uuid
ORDER BY created_at ASC
LIMIT $2
"""

_RECENT_MESSAGES = """
SELECT id::text, chat_id::text, role, content, citations, created_at
FROM (
    SELECT id, chat_id, role, content, citations, created_at
    FROM chat_messages
    WHERE chat_id = $1::uuid
    ORDER BY created_at DESC
    LIMIT $2
) t
ORDER BY created_at ASC
"""


def _chat(row: asyncpg.Record) -> Chat:
    return Chat(
        id=row["id"],
        user_id=row["user_id"],
        title=row["title"],
        created_at=row["created_at"],
        updated_at=row["updated_at"],
    )


def _message(row: asyncpg.Record) -> Message:
    raw = row["citations"]
    items = json.loads(raw) if isinstance(raw, str) else (raw or [])
    return Message(
        id=row["id"],
        chat_id=row["chat_id"],
        role=row["role"],
        content=row["content"],
        citations=[Citation(**item) for item in items],
        created_at=row["created_at"],
    )


class PostgresConversationRepository:
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
            async with self._pool.acquire() as conn:
                await conn.execute(_DDL)
        except (asyncpg.PostgresError, OSError) as exc:
            raise PostSourceUnavailable(
                f"failed to init conversation repository: {exc}", cause=exc
            ) from exc
        logger.info("conversation_repository.connected")

    async def close(self) -> None:
        if self._pool is not None:
            await self._pool.close()
            self._pool = None

    async def create_chat(self, user_id: str, title: str) -> Chat:
        assert self._pool is not None
        row = await self._pool.fetchrow(_INSERT_CHAT, user_id, title)
        return _chat(row)

    async def list_chats(self, user_id: str, limit: int) -> list[Chat]:
        assert self._pool is not None
        rows = await self._pool.fetch(_LIST_CHATS, user_id, limit)
        return [_chat(row) for row in rows]

    async def get_chat(self, chat_id: str) -> Chat | None:
        assert self._pool is not None
        row = await self._pool.fetchrow(_GET_CHAT, chat_id)
        return _chat(row) if row is not None else None

    async def delete_chat(self, chat_id: str) -> bool:
        assert self._pool is not None
        row = await self._pool.fetchrow(_DELETE_CHAT, chat_id)
        return row is not None

    async def set_title(self, chat_id: str, title: str) -> None:
        assert self._pool is not None
        await self._pool.execute(_SET_TITLE, chat_id, title)

    async def add_message(
        self,
        chat_id: str,
        role: MessageRole,
        content: str,
        citations: list[Citation],
    ) -> Message:
        assert self._pool is not None
        payload = json.dumps([c.model_dump() for c in citations])
        async with self._pool.acquire() as conn:
            async with conn.transaction():
                row = await conn.fetchrow(_INSERT_MESSAGE, chat_id, role, content, payload)
                await conn.execute(_TOUCH_CHAT, chat_id)
        return _message(row)

    async def list_messages(self, chat_id: str, limit: int) -> list[Message]:
        assert self._pool is not None
        rows = await self._pool.fetch(_LIST_MESSAGES, chat_id, limit)
        return [_message(row) for row in rows]

    async def recent_messages(self, chat_id: str, limit: int) -> list[Message]:
        assert self._pool is not None
        rows = await self._pool.fetch(_RECENT_MESSAGES, chat_id, limit)
        return [_message(row) for row in rows]
