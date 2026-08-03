import time
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from prometheus_client import (
    CONTENT_TYPE_LATEST,
    CollectorRegistry,
    Counter,
    Histogram,
    generate_latest,
)

REGISTRY = CollectorRegistry(auto_describe=True)

INGEST_TOTAL = Counter(
    "ai_service_ingest_total",
    "Ingest pipeline outcomes by content kind and status.",
    ["kind", "status"],
    registry=REGISTRY,
)

INGEST_DURATION = Histogram(
    "ai_service_ingest_duration_seconds",
    "Ingest pipeline duration by content kind.",
    ["kind"],
    registry=REGISTRY,
)

EMBED_CALLS = Counter(
    "ai_service_embed_calls_total",
    "Embedding adapter calls.",
    ["task_type", "outcome"],
    registry=REGISTRY,
)

LLM_CALLS = Counter(
    "ai_service_llm_calls_total",
    "LLM adapter calls by op.",
    ["op", "outcome"],
    registry=REGISTRY,
)

WEAVIATE_CALLS = Counter(
    "ai_service_weaviate_calls_total",
    "Weaviate adapter calls by op.",
    ["op", "outcome"],
    registry=REGISTRY,
)

RAG_CHAT_DURATION = Histogram(
    "ai_service_rag_chat_duration_seconds",
    "Assistant chat use case duration.",
    registry=REGISTRY,
)

SCREENING_DURATION = Histogram(
    "ai_service_screening_duration_seconds",
    "Screening use case duration.",
    registry=REGISTRY,
)

SCREENING_BLOCKS = Counter(
    "ai_service_screening_blocks_total",
    "Screening blocks by reason.",
    ["reason"],
    registry=REGISTRY,
)


@asynccontextmanager
async def observe(counter: Counter, **labels: str) -> AsyncIterator[None]:
    try:
        yield
    except BaseException:
        counter.labels(outcome="error", **labels).inc()
        raise
    counter.labels(outcome="success", **labels).inc()


@asynccontextmanager
async def time_histogram(histogram: Histogram, **labels: str) -> AsyncIterator[None]:
    start = time.monotonic()
    try:
        yield
    finally:
        elapsed = time.monotonic() - start
        if labels:
            histogram.labels(**labels).observe(elapsed)
        else:
            histogram.observe(elapsed)


def render_latest() -> tuple[bytes, str]:
    return generate_latest(REGISTRY), CONTENT_TYPE_LATEST
