from __future__ import annotations

import json

import httpx

from infrastructure.observability import get_logger

logger = get_logger(__name__)


class ContentServiceClient:
    def __init__(self, *, base_url: str, api_key: str, timeout_s: float = 15.0) -> None:
        self._base_url = base_url.rstrip("/")
        self._api_key = api_key
        self._timeout_s = timeout_s

    async def patch_place_summary(
        self,
        *,
        place_id: str,
        summary_text: str,
        citations: list[dict[str, str]],
    ) -> None:
        url = f"{self._base_url}/v1/internal/places/{place_id}/summary"
        headers = {"Content-Type": "application/json"}
        if self._api_key:
            headers["X-Internal-API-Key"] = self._api_key
        payload = {
            "summary_text": summary_text,
            "summary_citations": citations,
        }
        async with httpx.AsyncClient(timeout=self._timeout_s) as client:
            resp = await client.patch(url, headers=headers, content=json.dumps(payload))
            if resp.status_code >= 400:
                logger.error(
                    "content_service.patch_summary_failed",
                    place_id=place_id,
                    status=resp.status_code,
                    body=resp.text[:500],
                )
                resp.raise_for_status()
