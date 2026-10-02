"""Unit tests for embedding provider factory."""

from __future__ import annotations

import pytest

from app.adapters.embeddings.factory import build_embedding_provider
from app.adapters.embeddings.gemini import GeminiEmbeddingClient
from app.adapters.embeddings.sentence_transformers import SentenceTransformerEmbedder
from app.config import Settings


def test_build_embedding_gemini_default():
    settings = Settings(
        embedding_provider="gemini",
        gemini_api_key="g-key",
        gemini_embedding_model="text-embedding-004",
        embedding_dimensions=768,
    )
    emb = build_embedding_provider(settings)
    assert isinstance(emb, GeminiEmbeddingClient)
    assert emb.configured
    assert emb.model_name == "text-embedding-004"
    assert emb.dimensions == 768


def test_build_embedding_local():
    settings = Settings(
        embedding_provider="local",
        embedding_model="sentence-transformers/all-MiniLM-L6-v2",
        embedding_dimensions=384,
    )
    emb = build_embedding_provider(settings)
    assert isinstance(emb, SentenceTransformerEmbedder)
    assert emb.dimensions == 384


def test_build_embedding_rejects_unknown():
    with pytest.raises(ValueError, match="unsupported EMBEDDING_PROVIDER"):
        build_embedding_provider(Settings(embedding_provider="openai"))
