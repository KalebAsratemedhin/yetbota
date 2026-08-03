from pydantic import BaseModel, ConfigDict

from domain.entities.chunk import ContentKind


class Citation(BaseModel):
    model_config = ConfigDict(frozen=True)

    source_id: str
    kind: ContentKind
    text: str
    score: float
