from pydantic import BaseModel, ConfigDict, Field


class Post(BaseModel):
    model_config = ConfigDict(frozen=True)

    id: str
    title: str
    description: str
    user_id: str
    is_question: bool
    tags: list[str] = Field(default_factory=list)
    latitude: float | None = None
    longitude: float | None = None
