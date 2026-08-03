from pydantic import BaseModel, ConfigDict

from domain.entities.chunk import Chunk

Embedding = list[float]


class ScoredChunk(BaseModel):
    model_config = ConfigDict(frozen=True)

    chunk: Chunk
    score: float
