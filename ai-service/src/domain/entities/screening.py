from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

from domain.entities.chunk import ContentKind

ScreeningCategory = Literal["profanity", "hate", "sexual", "violence", "harassment"]


class ScreeningRequest(BaseModel):
    model_config = ConfigDict(frozen=True)

    text: str
    kind: ContentKind


class ScreeningResult(BaseModel):
    model_config = ConfigDict(frozen=True)

    ok: bool
    reason: ScreeningCategory | Literal["screening_unavailable"] | None = None
    categories: dict[ScreeningCategory, float] = Field(default_factory=dict)
