from typing import Protocol, runtime_checkable

from domain.entities import Chat, Citation, Message, MessageRole


@runtime_checkable
class ConversationRepository(Protocol):
    async def create_chat(self, user_id: str, title: str) -> Chat: ...

    async def list_chats(self, user_id: str, limit: int) -> list[Chat]: ...

    async def get_chat(self, chat_id: str) -> Chat | None: ...

    async def delete_chat(self, chat_id: str) -> bool: ...

    async def set_title(self, chat_id: str, title: str) -> None: ...

    async def add_message(
        self,
        chat_id: str,
        role: MessageRole,
        content: str,
        citations: list[Citation],
    ) -> Message: ...

    async def list_messages(self, chat_id: str, limit: int) -> list[Message]: ...

    async def recent_messages(self, chat_id: str, limit: int) -> list[Message]: ...
