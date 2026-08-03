from google import genai
from google.genai import errors, types
from tenacity import (
    AsyncRetrying,
    retry_if_exception,
    stop_after_attempt,
    wait_exponential_jitter,
)

from application.errors import EmbeddingFailed
from domain.entities import Embedding
from domain.ports import EmbedTaskType
from infrastructure.config.settings import GeminiSettings
from infrastructure.observability import EMBED_CALLS, observe

_BATCH_LIMIT = 96


def _is_retryable(exc: BaseException) -> bool:
    if isinstance(exc, errors.ServerError):
        return True
    if isinstance(exc, errors.ClientError) and getattr(exc, "code", None) == 429:
        return True
    return False


def _api_error_transient(exc: errors.APIError) -> bool:
    return _is_retryable(exc)


class GeminiEmbedder:
    def __init__(self, settings: GeminiSettings) -> None:
        self._settings = settings
        self._client = genai.Client(api_key=settings.api_key)

    async def embed(
        self, text: str, *, task_type: EmbedTaskType = "RETRIEVAL_DOCUMENT"
    ) -> Embedding:
        result = await self._call([text], task_type=task_type)
        return result[0]

    async def embed_batch(
        self, texts: list[str], *, task_type: EmbedTaskType = "RETRIEVAL_DOCUMENT"
    ) -> list[Embedding]:
        if not texts:
            return []
        out: list[Embedding] = []
        for start in range(0, len(texts), _BATCH_LIMIT):
            shard = texts[start : start + _BATCH_LIMIT]
            out.extend(await self._call(shard, task_type=task_type))
        return out

    async def _call(self, texts: list[str], *, task_type: EmbedTaskType) -> list[Embedding]:
        config = types.EmbedContentConfig(
            task_type=task_type,
            output_dimensionality=self._settings.embedding_dimensions,
        )
        async with observe(EMBED_CALLS, task_type=task_type):
            try:
                async for attempt in AsyncRetrying(
                    stop=stop_after_attempt(3),
                    wait=wait_exponential_jitter(initial=0.5, max=4.0),
                    retry=retry_if_exception(_is_retryable),
                    reraise=True,
                ):
                    with attempt:
                        response = await self._client.aio.models.embed_content(
                            model=self._settings.embedding_model,
                            contents=list(texts),
                            config=config,
                        )
            except errors.APIError as exc:
                raise EmbeddingFailed(
                    f"gemini embedding call failed: {exc}",
                    cause=exc,
                    transient=_api_error_transient(exc),
                ) from exc

        embeddings = response.embeddings or []
        if len(embeddings) != len(texts):
            raise EmbeddingFailed(
                f"gemini returned {len(embeddings)} embeddings for {len(texts)} inputs",
                transient=False,
            )

        out: list[Embedding] = []
        for item in embeddings:
            values = item.values or []
            if len(values) != self._settings.embedding_dimensions:
                raise EmbeddingFailed(
                    f"gemini returned {len(values)}-d vector, "
                    f"expected {self._settings.embedding_dimensions}",
                    transient=False,
                )
            out.append(list(values))
        return out
