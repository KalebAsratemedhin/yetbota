from fastapi import APIRouter, HTTPException, Request
from pydantic import BaseModel, Field

from application.match_place import MatchPlace
from domain.entities.place_match import PlaceMatchQuery, PlaceMatchResult

router = APIRouter(prefix="/v1/places", tags=["places"])


class MatchPlaceRequest(BaseModel):
    text: str = Field(min_length=1)
    exclude_source_id: str | None = None
    latitude: float | None = None
    longitude: float | None = None
    radius_km: float | None = None
    limit: int = Field(default=40, ge=1, le=100)


@router.post("/match", response_model=PlaceMatchResult)
async def match_place(body: MatchPlaceRequest, request: Request) -> PlaceMatchResult:
    use_case: MatchPlace | None = getattr(request.app.state, "place_matcher", None)
    if use_case is None:
        raise HTTPException(status_code=503, detail="Place matching is unavailable.")
    return await use_case.execute(
        PlaceMatchQuery(
            text=body.text,
            exclude_source_id=body.exclude_source_id,
            latitude=body.latitude,
            longitude=body.longitude,
            radius_km=body.radius_km,
            limit=body.limit,
        )
    )
