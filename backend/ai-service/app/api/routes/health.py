from fastapi import APIRouter

router = APIRouter(tags=["Health"])


@router.get("/healthz", summary="Liveness")
async def healthz() -> dict[str, str]:
    return {"status": "ok"}
