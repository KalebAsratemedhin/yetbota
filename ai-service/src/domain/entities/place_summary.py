from pydantic import BaseModel, ConfigDict, Field


class PlaceSummaryCitation(BaseModel):
    model_config = ConfigDict(frozen=True)

    contribution_id: str
    kind: str
    excerpt: str


class PlaceSummaryResult(BaseModel):
    model_config = ConfigDict(frozen=True)

    place_id: str
    summary_text: str
    citations: list[PlaceSummaryCitation] = Field(default_factory=list)
    skipped: bool = False
    skip_reason: str | None = None
