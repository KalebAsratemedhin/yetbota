"""Similarity hit between two posts."""

from __future__ import annotations

from dataclasses import dataclass
from uuid import UUID


@dataclass(frozen=True)
class SimilarHit:
    post_id: UUID
    score: float
