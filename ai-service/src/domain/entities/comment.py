from pydantic import BaseModel, ConfigDict


class Comment(BaseModel):
    model_config = ConfigDict(frozen=True)

    id: str
    body: str
    user_id: str
    post_id: str
    is_answer: bool
    comment_id: str | None = None
