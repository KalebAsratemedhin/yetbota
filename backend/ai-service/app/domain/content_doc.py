"""Typed content units for chatbot / RAG ingest (separate from post similar graph)."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Literal
from uuid import UUID

ContentKind = Literal["post", "comment", "question", "answer"]


@dataclass(frozen=True)
class ContentDocument:
    kind: ContentKind
    entity_id: UUID
    author_id: UUID
    text: str
    parent_id: UUID | None = None
    post_id: UUID | None = None
    title: str = ""
    lat: float | None = None
    lng: float | None = None

    @property
    def record_key(self) -> str:
        """Stable Surreal record id: kind_entityuuid."""
        return f"{self.kind}_{self.entity_id}"
