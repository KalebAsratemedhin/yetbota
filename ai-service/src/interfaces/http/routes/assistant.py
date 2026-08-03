from fastapi import APIRouter, HTTPException, Query, Request, Response
from pydantic import BaseModel, Field

from application.consult_assistant import ConsultAssistant
from application.errors import ChatNotFound, LLMUnavailable
from application.manage_conversation import ManageConversation
from domain.entities import Chat, ChatQuery, ChatResponse, Message
from infrastructure.observability import RAG_CHAT_DURATION, time_histogram

router = APIRouter(prefix="/v1/assistant", tags=["assistant"])


class CreateChatRequest(BaseModel):
    user_id: str
    title: str = ""


class SendMessageRequest(BaseModel):
    text: str = Field(min_length=1)


class ChatListResponse(BaseModel):
    chats: list[Chat]


class MessageListResponse(BaseModel):
    messages: list[Message]


def _conversations(request: Request) -> ManageConversation:
    return request.app.state.conversations


@router.post("/chat", response_model=ChatResponse)
async def chat(query: ChatQuery, request: Request) -> ChatResponse:
    use_case: ConsultAssistant = request.app.state.assistant
    try:
        async with time_histogram(RAG_CHAT_DURATION):
            return await use_case.execute(query)
    except LLMUnavailable as exc:
        raise HTTPException(
            status_code=503,
            detail="AI Assistant is currently unavailable.",
        ) from exc


@router.post("/chats", response_model=Chat, status_code=201)
async def create_chat(body: CreateChatRequest, request: Request) -> Chat:
    return await _conversations(request).create_chat(body.user_id, body.title)


@router.get("/chats", response_model=ChatListResponse)
async def list_chats(
    request: Request,
    user_id: str = Query(...),
    limit: int = Query(20, ge=1, le=100),
) -> ChatListResponse:
    chats = await _conversations(request).list_chats(user_id, limit)
    return ChatListResponse(chats=chats)


@router.get("/chats/{chat_id}/messages", response_model=MessageListResponse)
async def list_messages(
    chat_id: str,
    request: Request,
    limit: int = Query(50, ge=1, le=200),
) -> MessageListResponse:
    try:
        messages = await _conversations(request).list_messages(chat_id, limit)
    except ChatNotFound as exc:
        raise HTTPException(status_code=404, detail="Chat not found.") from exc
    return MessageListResponse(messages=messages)


@router.post("/chats/{chat_id}/messages", response_model=Message, status_code=201)
async def send_message(chat_id: str, body: SendMessageRequest, request: Request) -> Message:
    try:
        async with time_histogram(RAG_CHAT_DURATION):
            return await _conversations(request).send_message(chat_id, body.text)
    except ChatNotFound as exc:
        raise HTTPException(status_code=404, detail="Chat not found.") from exc
    except LLMUnavailable as exc:
        raise HTTPException(
            status_code=503,
            detail="AI Assistant is currently unavailable.",
        ) from exc


@router.delete("/chats/{chat_id}", status_code=204)
async def delete_chat(chat_id: str, request: Request) -> Response:
    try:
        await _conversations(request).delete_chat(chat_id)
    except ChatNotFound as exc:
        raise HTTPException(status_code=404, detail="Chat not found.") from exc
    return Response(status_code=204)
