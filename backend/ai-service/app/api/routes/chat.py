from __future__ import annotations

from typing import Annotated, Literal

from fastapi import APIRouter, Depends, HTTPException, Request
from pydantic import BaseModel, Field

from app.services.chat import ChatService

router = APIRouter(tags=["Chat"])


class ChatTurn(BaseModel):
    role: Literal["user", "assistant"]
    content: str = Field(min_length=1, max_length=8000)


class ChatRequest(BaseModel):
    message: str = Field(min_length=1, max_length=4000)
    history: list[ChatTurn] = Field(default_factory=list, max_length=20)


class ChatSource(BaseModel):
    kind: str
    entity_id: str
    score: float
    title: str = ""
    text: str = ""
    parent_id: str | None = None
    post_id: str | None = None


class ChatResponse(BaseModel):
    reply: str
    model: str
    sources: list[ChatSource]


def get_chat_service(request: Request) -> ChatService:
    svc = getattr(request.app.state, "chat", None)
    if svc is None:
        raise HTTPException(status_code=503, detail="chat service unavailable")
    return svc


@router.post("/chat", summary="RAG chat over posts, comments, and Q&A")
async def chat(
    body: ChatRequest,
    svc: Annotated[ChatService, Depends(get_chat_service)],
) -> ChatResponse:
    history = [{"role": t.role, "content": t.content} for t in body.history]
    try:
        result = await svc.chat(body.message, history=history)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e)) from e
    except RuntimeError as e:
        msg = str(e)
        if "not configured" in msg or "API_KEY" in msg:
            raise HTTPException(status_code=503, detail=msg) from e
        raise HTTPException(status_code=502, detail=msg) from e
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"chat failed: {e}") from e
    return ChatResponse(
        reply=result["reply"],
        model=result["model"],
        sources=[ChatSource(**s) for s in result["sources"]],
    )
