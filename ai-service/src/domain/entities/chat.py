from pydantic import BaseModel, ConfigDict, Field

from domain.entities.citation import Citation
from domain.entities.conversation import Message


class ChatQuery(BaseModel):
    model_config = ConfigDict(frozen=True)

    text: str
    user_id: str | None = None
    history: list[Message] = Field(default_factory=list)
    place_id: str | None = None
    latitude: float | None = None
    longitude: float | None = None
    radius_km: float | None = None


class ChatResponse(BaseModel):
    model_config = ConfigDict(frozen=True)

    answer: str
    citations: list[Citation] = Field(default_factory=list)
