from application.consult_assistant import ConsultAssistant
from application.errors import ChatNotFound
from domain.entities import Chat, ChatQuery, Message
from domain.ports import ConversationRepository

_TITLE_MAX_LEN = 60
_HISTORY_LIMIT = 8


def _derive_title(text: str) -> str:
    collapsed = " ".join(text.split())
    if len(collapsed) <= _TITLE_MAX_LEN:
        return collapsed
    return collapsed[:_TITLE_MAX_LEN].rstrip() + "..."


class ManageConversation:
    def __init__(
        self,
        *,
        repository: ConversationRepository,
        assistant: ConsultAssistant,
    ) -> None:
        self._repository = repository
        self._assistant = assistant

    async def create_chat(self, user_id: str, title: str = "") -> Chat:
        return await self._repository.create_chat(user_id, title.strip())

    async def list_chats(self, user_id: str, limit: int) -> list[Chat]:
        return await self._repository.list_chats(user_id, limit)

    async def list_messages(self, chat_id: str, limit: int) -> list[Message]:
        chat = await self._repository.get_chat(chat_id)
        if chat is None:
            raise ChatNotFound(f"chat {chat_id} not found")
        return await self._repository.list_messages(chat_id, limit)

    async def delete_chat(self, chat_id: str) -> None:
        deleted = await self._repository.delete_chat(chat_id)
        if not deleted:
            raise ChatNotFound(f"chat {chat_id} not found")

    async def send_message(self, chat_id: str, text: str) -> Message:
        chat = await self._repository.get_chat(chat_id)
        if chat is None:
            raise ChatNotFound(f"chat {chat_id} not found")

        history = await self._repository.recent_messages(chat_id, _HISTORY_LIMIT)
        await self._repository.add_message(chat_id, "user", text, [])
        if not chat.title:
            await self._repository.set_title(chat_id, _derive_title(text))

        answer = await self._assistant.execute(
            ChatQuery(text=text, user_id=chat.user_id, history=history)
        )
        return await self._repository.add_message(
            chat_id,
            "assistant",
            answer.answer,
            answer.citations,
        )


__all__ = ["ManageConversation"]
