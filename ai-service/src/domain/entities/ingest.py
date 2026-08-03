from datetime import datetime
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

from domain.entities.chunk import ContentKind

Verdict = Literal["unique", "duplicate", "indexed", "error"]


class IngestRequest(BaseModel):
    model_config = ConfigDict(frozen=True)

    content_id: str
    kind: ContentKind
    user_id: str
    text: str
    parent_id: str | None = None
    attached_post_id: str | None = None
    place_id: str | None = None
    tags: list[str] = Field(default_factory=list)
    category: str | None = None
    latitude: float | None = None
    longitude: float | None = None
    context: str | None = None


class IngestResult(BaseModel):
    model_config = ConfigDict(frozen=True)

    content_id: str
    kind: ContentKind
    verdict: Verdict
    duplicate_of: str | None = None
    error_code: str | None = None
    processed_at: datetime


class DeleteRequest(BaseModel):
    model_config = ConfigDict(frozen=True)

    content_id: str
    kind: ContentKind


class DeleteResult(BaseModel):
    model_config = ConfigDict(frozen=True)

    content_id: str
    kind: ContentKind
    deleted: bool
    error_code: str | None = None
    processed_at: datetime
