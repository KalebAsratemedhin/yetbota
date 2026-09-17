"""OpenAI Chat Completions client (HTTP)."""

from __future__ import annotations

import logging
from typing import Any

import httpx

logger = logging.getLogger(__name__)


class OpenAIChatClient:
    def __init__(
        self,
        api_key: str,
        *,
        model: str = "gpt-4o-mini",
        base_url: str = "https://api.openai.com/v1",
        timeout_s: float = 60.0,
    ) -> None:
        self._api_key = api_key.strip()
        self._model = model
        self._base_url = base_url.rstrip("/")
        self._timeout = timeout_s

    @property
    def configured(self) -> bool:
        return bool(self._api_key)

    @property
    def model_name(self) -> str:
        return self._model

    async def complete(
        self,
        *,
        system: str,
        messages: list[dict[str, str]],
    ) -> str:
        if not self._api_key:
            raise RuntimeError("OPENAI_API_KEY is not configured")
        payload: dict[str, Any] = {
            "model": self._model,
            "messages": [{"role": "system", "content": system}, *messages],
            "temperature": 0.4,
        }
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            resp = await client.post(
                f"{self._base_url}/chat/completions",
                headers={
                    "Authorization": f"Bearer {self._api_key}",
                    "Content-Type": "application/json",
                },
                json=payload,
            )
            if resp.status_code >= 400:
                logger.error(
                    "openai_chat_failed status=%s body=%s",
                    resp.status_code,
                    resp.text[:500],
                )
                resp.raise_for_status()
            data = resp.json()
        try:
            return str(data["choices"][0]["message"]["content"]).strip()
        except (KeyError, IndexError, TypeError) as e:
            raise RuntimeError(f"unexpected OpenAI response shape: {e}") from e
