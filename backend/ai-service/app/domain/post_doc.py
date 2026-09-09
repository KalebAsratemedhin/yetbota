"""Domain types for post ingest."""

from __future__ import annotations

from dataclasses import dataclass
from uuid import UUID


@dataclass(frozen=True)
class PostDocument:
    id: UUID
    author_id: UUID
    title: str
    caption: str
    lat: float
    lng: float
    hashtags: str

    def embed_text(self) -> str:
        """Text fed to the embedding model (place title + caption + tags)."""
        parts = [self.title.strip(), self.caption.strip(), self.hashtags.strip()]
        return "\n".join(p for p in parts if p)
