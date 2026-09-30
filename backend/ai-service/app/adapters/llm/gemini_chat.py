"""Google Gemini generateContent client (HTTP)."""

from __future__ import annotations

import asyncio
import logging
from typing import Any

import httpx


logger = logging.getLogger(__name__)

_RETRYABLE = {429, 503}
_MAX_ATTEMPTS = 4


class GeminiChatClient:
    def __init__(
        self,
        api_key: str,
        *,
        model: str = "gemini-2.5-flash",
        base_url: str = "https://generativelanguage.googleapis.com/v1beta",
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
        json_mode: bool = False,
    ) -> str:
        if not self._api_key:
            raise RuntimeError("GEMINI_API_KEY is not configured")

        contents = _to_gemini_contents(messages)
        generation: dict[str, Any] = {"temperature": 0.4}
        if json_mode:
            generation["responseMimeType"] = "application/json"
        payload: dict[str, Any] = {
            "contents": contents,
            "generationConfig": generation,
        }
        if system.strip():
            payload["system_instruction"] = {"parts": [{"text": system}]}

        url = f"{self._base_url}/models/{self._model}:generateContent"
        data: dict[str, Any] | None = None
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            for attempt in range(_MAX_ATTEMPTS):
                resp = await client.post(
                    url,
                    params={"key": self._api_key},
                    headers={"Content-Type": "application/json"},
                    json=payload,
                )
                if resp.status_code in _RETRYABLE and attempt < _MAX_ATTEMPTS - 1:
                    delay = 0.75 * (2**attempt)
                    logger.warning(
                        "gemini_chat_retry status=%s attempt=%s sleep=%.2fs",
                        resp.status_code,
                        attempt + 1,
                        delay,
                    )
                    await asyncio.sleep(delay)
                    continue
                if resp.status_code >= 400:
                    logger.error(
                        "gemini_chat_failed status=%s body=%s",
                        resp.status_code,
                        resp.text[:500],
                    )
                    resp.raise_for_status()
                data = resp.json()
                break

        if data is None:
            raise RuntimeError("Gemini request failed with no response")

        try:
            parts = data["candidates"][0]["content"]["parts"]
            texts = [str(p.get("text", "")) for p in parts if p.get("text")]
            reply = "".join(texts).strip()
        except (KeyError, IndexError, TypeError) as e:
            raise RuntimeError(f"unexpected Gemini response shape: {e}") from e
        if not reply:
            raise RuntimeError("Gemini returned an empty reply")
        return reply


def _to_gemini_contents(messages: list[dict[str, str]]) -> list[dict[str, Any]]:
    """Map OpenAI-style roles to Gemini contents (assistant → model)."""
    contents: list[dict[str, Any]] = []
    for msg in messages:
        role = (msg.get("role") or "").strip()
        text = (msg.get("content") or "").strip()
        if not text:
            continue
        if role == "assistant":
            gemini_role = "model"
        elif role == "user":
            gemini_role = "user"
        else:
            continue
        if contents and contents[-1]["role"] == gemini_role:
            prev = contents[-1]["parts"][0]["text"]
            contents[-1]["parts"][0]["text"] = f"{prev}\n\n{text}"
        else:
            contents.append({"role": gemini_role, "parts": [{"text": text}]})
    return contents
