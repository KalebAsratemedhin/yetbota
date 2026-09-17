"""Chat / RAG domain types."""

from __future__ import annotations

from dataclasses import dataclass
from uuid import UUID


@dataclass(frozen=True)
class ContentHit:
    kind: str
    entity_id: UUID
    score: float
    text: str
    title: str = ""
    parent_id: UUID | None = None
    post_id: UUID | None = None
