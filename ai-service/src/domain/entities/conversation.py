from datetime import datetime
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

from domain.entities.citation import Citation

MessageRole = Literal["user", "assistant"]


class Chat(BaseModel):
    model_config = ConfigDict(frozen=True)

    id: str
    user_id: str
    title: str
    created_at: datetime
    updated_at: datetime


class Message(BaseModel):
    model_config = ConfigDict(frozen=True)

    id: str
    chat_id: str
    role: MessageRole
    content: str
    citations: list[Citation] = Field(default_factory=list)
    created_at: datetime
