import json
from typing import Any

from google import genai
from google.genai import errors, types
from tenacity import (
    AsyncRetrying,
    retry_if_exception,
    stop_after_attempt,
    wait_exponential_jitter,
)

from application.errors import LLMUnavailable
from infrastructure.config.settings import GeminiSettings
from infrastructure.observability import LLM_CALLS, observe


def _is_retryable(exc: BaseException) -> bool:
    if isinstance(exc, errors.ServerError):
        return True
    if isinstance(exc, errors.ClientError) and getattr(exc, "code", None) == 429:
        return True
    return False


class GeminiLLM:
    def __init__(self, settings: GeminiSettings) -> None:
        self._settings = settings
        self._client = genai.Client(api_key=settings.api_key)

    async def generate(
        self,
        prompt: str,
        *,
        max_tokens: int,
        temperature: float,
    ) -> str:
        config = types.GenerateContentConfig(
            temperature=temperature,
            max_output_tokens=max_tokens,
        )
        response = await self._call(prompt, config=config, op="generate")
        return response.text or ""

    async def classify(
        self,
        prompt: str,
        *,
        schema: dict[str, Any],
        temperature: float = 0.0,
    ) -> dict[str, Any]:
        config = types.GenerateContentConfig(
            temperature=temperature,
            response_mime_type="application/json",
            response_schema=schema,
        )
        response = await self._call(prompt, config=config, op="classify")
        text = response.text or "{}"
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError as exc:
            raise LLMUnavailable(
                f"gemini returned non-json classification: {exc}", cause=exc
            ) from exc
        if not isinstance(parsed, dict):
            raise LLMUnavailable(
                f"gemini classification was not an object: {type(parsed).__name__}"
            )
        return parsed

    async def _call(
        self,
        prompt: str,
        *,
        config: types.GenerateContentConfig,
        op: str,
    ) -> types.GenerateContentResponse:
        async with observe(LLM_CALLS, op=op):
            try:
                async for attempt in AsyncRetrying(
                    stop=stop_after_attempt(3),
                    wait=wait_exponential_jitter(initial=0.5, max=4.0),
                    retry=retry_if_exception(_is_retryable),
                    reraise=True,
                ):
                    with attempt:
                        response = await self._client.aio.models.generate_content(
                            model=self._settings.llm_model,
                            contents=prompt,
                            config=config,
                        )
            except errors.APIError as exc:
                raise LLMUnavailable(f"gemini call failed: {exc}", cause=exc) from exc
        return response
