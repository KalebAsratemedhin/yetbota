"""Local embedding provider via fastembed (ONNX — no PyTorch required)."""

from __future__ import annotations

import asyncio
import logging
from pathlib import Path

logger = logging.getLogger(__name__)

# sentence-transformers/all-MiniLM-L6-v2 via fastembed → 384 dims
DEFAULT_DIMENSIONS = 384
# fastembed's default is tempfile.gettempdir()/fastembed_cache — wiped on reboot / macOS cleanup.
DEFAULT_CACHE_DIR = Path.home() / ".cache" / "yetbota" / "fastembed"


class SentenceTransformerEmbedder:
    """Name kept for config compatibility; backed by fastembed TextEmbedding."""

    def __init__(
        self,
        model_name: str,
        dimensions: int = DEFAULT_DIMENSIONS,
        cache_dir: str | Path | None = None,
    ) -> None:
        self._model_name = model_name
        self._dimensions = dimensions
        self._cache_dir = Path(cache_dir) if cache_dir else DEFAULT_CACHE_DIR
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

        self._cache_dir.mkdir(parents=True, exist_ok=True)
        logger.info(
            "loading_embedding_model model=%s cache_dir=%s",
            self._model_name,
            self._cache_dir,
        )
        self._model = TextEmbedding(
            model_name=self._model_name,
            cache_dir=str(self._cache_dir),
        )
        return self._model

    async def warmup(self) -> None:
        """Download (once into persistent cache) / load the ONNX model at process start."""
        await asyncio.to_thread(self._load)
        logger.info(
            "embedding_model_ready model=%s dim=%s cache_dir=%s",
            self._model_name,
            self._dimensions,
            self._cache_dir,
        )

    async def embed(self, text: str, *, task: str | None = None) -> list[float]:
        del task  # local model has no query/document split
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
