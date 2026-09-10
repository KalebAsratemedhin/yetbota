from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.adapters.embeddings.sentence_transformers import SentenceTransformerEmbedder
from app.adapters.postgres.posts import PostgresPostReader
from app.adapters.postgres.session import create_engine, create_session_factory
from app.adapters.rabbitmq.consumer import run_consumer
from app.adapters.surreal.client import SurrealConfig
from app.adapters.surreal.store import SurrealPostEmbeddingStore
from app.api.routes import health
from app.config import get_settings
from app.services.ingest import IngestService

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s %(message)s")
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    settings = get_settings()
    engine = create_engine(settings)
    sessions = create_session_factory(engine)
    embedder = SentenceTransformerEmbedder(
        settings.embedding_model, dimensions=settings.embedding_dimensions
    )
    store = SurrealPostEmbeddingStore(
        SurrealConfig.from_settings(settings),
        embedding_dimensions=settings.embedding_dimensions,
    )
    await store.connect()
    ingest = IngestService(
        PostgresPostReader(sessions),
        embedder,
        store,
        similarity_top_k=settings.similarity_top_k,
        similarity_min_score=settings.similarity_min_score,
    )

    stop = asyncio.Event()
    task = asyncio.create_task(
        run_consumer(settings, stop, ingest.ingest_post),
        name="rabbitmq-consumer",
    )
    logger.info(
        "ai_service_started rabbitmq=%s model=%s dim=%s top_k=%s min_score=%s",
        settings.rabbitmq_url,
        settings.embedding_model,
        settings.embedding_dimensions,
        settings.similarity_top_k,
        settings.similarity_min_score,
    )
    try:
        yield
    finally:
        stop.set()
        try:
            await asyncio.wait_for(task, timeout=5)
        except asyncio.TimeoutError:
            task.cancel()
            try:
                await task
            except asyncio.CancelledError:
                pass
        except Exception:
            logger.exception("rabbitmq_consumer_shutdown_error")
        await store.close()
        await engine.dispose()


def create_app() -> FastAPI:
    app = FastAPI(
        title="YetBota AI Service",
        description=(
            "Ingests posts from RabbitMQ post.created: embed title/caption/tags, "
            "upsert Surreal post_embedding, and write similar relation edges."
        ),
        version="0.1.0",
        lifespan=lifespan,
    )
    app.include_router(health.router)
    return app


app = create_app()
