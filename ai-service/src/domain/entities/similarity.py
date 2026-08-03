from pydantic import BaseModel, ConfigDict


class SimilarPost(BaseModel):
    model_config = ConfigDict(frozen=True)

    post_id: str
    score: float
