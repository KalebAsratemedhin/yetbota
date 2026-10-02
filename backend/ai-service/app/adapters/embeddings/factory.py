"""Build an EmbeddingProvider from settings."""

from __future__ import annotations

from app.adapters.embeddings.gemini import GeminiEmbeddingClient
from app.adapters.embeddings.sentence_transformers import SentenceTransformerEmbedder
from app.config import Settings
from app.domain.ports import EmbeddingProvider


def build_embedding_provider(settings: Settings) -> EmbeddingProvider:
    provider = (settings.embedding_provider or "").strip().lower()
    if provider == "gemini":
        return GeminiEmbeddingClient(
            settings.gemini_api_key,
            model=settings.gemini_embedding_model,
            dimensions=settings.embedding_dimensions,
            base_url=settings.gemini_base_url,
        )
    if provider in ("local", "fastembed", "minilm"):
        return SentenceTransformerEmbedder(
            settings.embedding_model,
            dimensions=settings.embedding_dimensions,
            cache_dir=settings.embedding_cache_dir or None,
        )
    raise ValueError(
        f"unsupported EMBEDDING_PROVIDER={settings.embedding_provider!r} "
        "(expected gemini or local)"
    )
