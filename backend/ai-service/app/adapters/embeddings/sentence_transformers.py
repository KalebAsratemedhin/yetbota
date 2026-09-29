"""Local embedding provider via fastembed (ONNX — no PyTorch required)."""

from __future__ import annotations

import asyncio
import logging

logger = logging.getLogger(__name__)

# sentence-transformers/all-MiniLM-L6-v2 via fastembed → 384 dims
DEFAULT_DIMENSIONS = 384


class SentenceTransformerEmbedder:
    """Name kept for config compatibility; backed by fastembed TextEmbedding."""

    def __init__(self, model_name: str, dimensions: int = DEFAULT_DIMENSIONS) -> None:
        self._model_name = model_name
        self._dimensions = dimensions
        self._model = None

    @property
    def model_name(self) -> str:
        return self._model_name

    @property
    def dimensions(self) -> int:
        return self._dimensions

    def _load(self):
        if self._model is not None:
            return self._model
        from fastembed import TextEmbedding

        logger.info("loading_embedding_model model=%s", self._model_name)
        self._model = TextEmbedding(model_name=self._model_name)
        return self._model

    async def warmup(self) -> None:
        """Download/load the ONNX model at process start (not on first request)."""
        await asyncio.to_thread(self._load)
        logger.info(
            "embedding_model_ready model=%s dim=%s",
            self._model_name,
            self._dimensions,
        )

    async def embed(self, text: str) -> list[float]:
        text = (text or "").strip() or " "
        model = await asyncio.to_thread(self._load)

        def _encode() -> list[float]:
            vec = next(model.embed([text]))
            return [float(x) for x in vec]

        out = await asyncio.to_thread(_encode)
        if len(out) != self._dimensions:
            raise ValueError(
                f"embedding dim {len(out)} != expected {self._dimensions} for {self._model_name}"
            )
        return out
