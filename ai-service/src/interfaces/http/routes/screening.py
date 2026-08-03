from fastapi import APIRouter, Request

from application.screen_text import ScreenText
from domain.entities import ScreeningRequest, ScreeningResult
from infrastructure.observability import (
    SCREENING_BLOCKS,
    SCREENING_DURATION,
    time_histogram,
)

router = APIRouter(prefix="/v1/screening", tags=["screening"])


@router.post("/check", response_model=ScreeningResult)
async def check(req: ScreeningRequest, request: Request) -> ScreeningResult:
    use_case: ScreenText = request.app.state.screening
    async with time_histogram(SCREENING_DURATION):
        result = await use_case.execute(req)
    if not result.ok:
        SCREENING_BLOCKS.labels(reason=result.reason or "unknown").inc()
    return result
