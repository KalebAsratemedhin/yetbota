"""Google Gemini embedContent client (HTTP)."""

from __future__ import annotations

import asyncio
import logging
from typing import Any

import httpx

logger = logging.getLogger(__name__)

_RETRYABLE = {429, 503}
_MAX_ATTEMPTS = 4

_TASK_MAP = {
    "document": "RETRIEVAL_DOCUMENT",
    "query": "RETRIEVAL_QUERY",
    None: "RETRIEVAL_DOCUMENT",
}


class GeminiEmbeddingClient:
    def __init__(
        self,
        api_key: str,
        *,
        model: str = "gemini-embedding-001",
        dimensions: int = 768,
        base_url: str = "https://generativelanguage.googleapis.com/v1beta",
        timeout_s: float = 60.0,
    ) -> None:
        self._api_key = api_key.strip()
        # ListModels returns "models/…"; callers may pass either form.
        self._model = model.strip().removeprefix("models/")
        self._dimensions = dimensions
        self._base_url = base_url.rstrip("/")
        self._timeout = timeout_s

    @property
    def configured(self) -> bool:
        return bool(self._api_key)

    @property
    def model_name(self) -> str:
        return self._model

    @property
    def dimensions(self) -> int:
        return self._dimensions

    async def warmup(self) -> None:
        if not self._api_key:
            logger.warning("gemini_embedding_not_configured")
            return
        logger.info(
            "gemini_embedding_ready model=%s dim=%s",
            self._model,
            self._dimensions,
        )

    async def embed(self, text: str, *, task: str | None = None) -> list[float]:
        if not self._api_key:
            raise RuntimeError("GEMINI_API_KEY is not configured")

        body_text = (text or "").strip() or " "
        task_type = _TASK_MAP.get(task, _TASK_MAP[None])
        payload: dict[str, Any] = {
            "content": {"parts": [{"text": body_text}]},
            "taskType": task_type,
        }
        # Default native size is 3072; pin so Surreal HNSW dim stays stable.
        if self._dimensions > 0:
            payload["outputDimensionality"] = self._dimensions

        url = f"{self._base_url}/models/{self._model}:embedContent"
        data: dict[str, Any] | None = None
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            for attempt in range(_MAX_ATTEMPTS):
                try:
                    resp = await client.post(
                        url,
                        params={"key": self._api_key},
                        headers={"Content-Type": "application/json"},
                        json=payload,
                    )
                except (
                    httpx.TimeoutException,
                    httpx.NetworkError,
                    httpx.RemoteProtocolError,
                ) as e:
                    if attempt >= _MAX_ATTEMPTS - 1:
                        raise
                    delay = 0.75 * (2**attempt)
                    logger.warning(
                        "gemini_embed_transport_retry err=%s attempt=%s sleep=%.2fs",
                        type(e).__name__,
                        attempt + 1,
                        delay,
                    )
                    await asyncio.sleep(delay)
                    continue
                if resp.status_code in _RETRYABLE and attempt < _MAX_ATTEMPTS - 1:
                    delay = 0.75 * (2**attempt)
                    logger.warning(
                        "gemini_embed_retry status=%s attempt=%s sleep=%.2fs",
                        resp.status_code,
                        attempt + 1,
                        delay,
                    )
                    await asyncio.sleep(delay)
                    continue
                if resp.status_code >= 400:
                    logger.error(
                        "gemini_embed_failed status=%s body=%s",
                        resp.status_code,
                        resp.text[:500],
                    )
                    resp.raise_for_status()
                data = resp.json()
                break

        if data is None:
            raise RuntimeError("Gemini embed request failed with no response")

        try:
            values = data["embedding"]["values"]
            out = [float(x) for x in values]
        except (KeyError, TypeError) as e:
            raise RuntimeError(f"unexpected Gemini embed response shape: {e}") from e

        if len(out) != self._dimensions:
            raise ValueError(
                f"embedding dim {len(out)} != expected {self._dimensions} "
                f"for {self._model} (check EMBEDDING_DIMENSIONS)"
            )
        return out
