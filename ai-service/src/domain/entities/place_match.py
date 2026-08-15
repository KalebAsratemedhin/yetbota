from pydantic import BaseModel, ConfigDict, Field


class PlaceMatchQuery(BaseModel):
    model_config = ConfigDict(frozen=True)

    text: str = Field(min_length=1)
    exclude_source_id: str | None = None
    latitude: float | None = None
    longitude: float | None = None
    radius_km: float | None = None
    limit: int = Field(default=40, ge=1, le=100)


class PlaceMatchResult(BaseModel):
    model_config = ConfigDict(frozen=True)

    place_id: str | None = None
    score: float = 0.0
