"""HTTP client for content-service place internal APIs."""

from __future__ import annotations

import logging
from typing import Any
from uuid import UUID

import httpx

logger = logging.getLogger(__name__)


class ContentPlacesClient:
    def __init__(
        self,
        base_url: str,
        service_token: str,
        *,
        timeout_s: float = 30.0,
    ) -> None:
        self._base = base_url.rstrip("/")
        self._token = service_token.strip()
        self._timeout = timeout_s

    def _headers(self) -> dict[str, str]:
        return {
            "Content-Type": "application/json",
            "X-Service-Token": self._token,
        }

    async def nearby(
        self,
        *,
        lat: float,
        lng: float,
        radius_m: float,
        limit: int = 20,
    ) -> list[dict[str, Any]]:
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.get(
                f"{self._base}/internal/places/nearby",
                params={
                    "lat": lat,
                    "lng": lng,
                    "radius_m": radius_m,
                    "limit": limit,
                },
                headers=self._headers(),
            )
            if resp.status_code >= 400:
                logger.error(
                    "places_nearby_failed status=%s body=%s",
                    resp.status_code,
                    resp.text[:300],
                )
                resp.raise_for_status()
            data = resp.json()
        return list(data.get("places") or [])

    async def get_by_post(self, post_id: UUID) -> dict[str, Any] | None:
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.get(
                f"{self._base}/internal/places/by-post/{post_id}",
                headers=self._headers(),
            )
            if resp.status_code == 404:
                return None
            if resp.status_code >= 400:
                resp.raise_for_status()
            return resp.json()

    async def attach(
        self,
        *,
        post_id: UUID,
        place_id: UUID | None = None,
        name: str = "",
        lat: float = 0.0,
        lng: float = 0.0,
        attach_score: float = 0.0,
        attach_method: str = "",
        embedding: list[float] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "post_id": str(post_id),
            "attach_score": attach_score,
            "attach_method": attach_method,
        }
        if place_id is not None:
            body["place_id"] = str(place_id)
        else:
            body["name"] = name
            body["lat"] = lat
            body["lng"] = lng
        if embedding:
            body["embedding"] = embedding
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.post(
                f"{self._base}/internal/places/attach",
                headers=self._headers(),
                json=body,
            )
            if resp.status_code >= 400:
                logger.error(
                    "places_attach_failed status=%s body=%s",
                    resp.status_code,
                    resp.text[:300],
                )
                resp.raise_for_status()
            return resp.json()

    async def add_entries(
        self,
        place_id: UUID,
        entries: list[dict[str, Any]],
    ) -> None:
        if not entries:
            return
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.post(
                f"{self._base}/internal/places/{place_id}/entries",
                headers=self._headers(),
                json={"entries": entries},
            )
            if resp.status_code >= 400:
                logger.error(
                    "places_entries_failed status=%s body=%s",
                    resp.status_code,
                    resp.text[:300],
                )
                resp.raise_for_status()

    async def update_summary(
        self,
        place_id: UUID,
        *,
        summary: str,
        model: str,
    ) -> None:
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.patch(
                f"{self._base}/internal/places/{place_id}/summary",
                headers=self._headers(),
                json={"summary": summary, "model": model},
            )
            if resp.status_code >= 400:
                logger.error(
                    "places_summary_failed status=%s body=%s",
                    resp.status_code,
                    resp.text[:300],
                )
                resp.raise_for_status()
