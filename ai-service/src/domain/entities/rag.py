from pydantic import BaseModel, ConfigDict

from domain.entities.embedding import ScoredChunk


class RagAnswerInput(BaseModel):
    model_config = ConfigDict(frozen=True)

    query: str
    hits: list[ScoredChunk]
