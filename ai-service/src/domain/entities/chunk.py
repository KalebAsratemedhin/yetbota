from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

ContentKind = Literal[
    "post",
    "question",
    "answer",
    "place",
    "tip",
    "change",
    "important",
]


class Chunk(BaseModel):
    model_config = ConfigDict(frozen=True)

    source_id: str
    kind: ContentKind
    text: str
    metadata: dict[str, str] = Field(default_factory=dict)
    segment_idx: int = 0
