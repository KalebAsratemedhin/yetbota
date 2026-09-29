from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.adapters.embeddings.sentence_transformers import SentenceTransformerEmbedder
from app.adapters.llm.factory import build_chat_llm
from app.adapters.postgres.content import PostgresContentReader
from app.adapters.postgres.session import create_engine, create_session_factory
from app.adapters.rabbitmq.consumer import run_consumer
from app.adapters.surreal.client import SurrealConfig
from app.adapters.surreal.store import SurrealPostEmbeddingStore
from app.api.routes import chat, health
from app.config import get_settings
from app.services.chat import ChatService
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
    await embedder.warmup()
    store = SurrealPostEmbeddingStore(
        SurrealConfig.from_settings(settings),
        embedding_dimensions=settings.embedding_dimensions,
    )
    await store.connect()
    ingest = IngestService(
        PostgresContentReader(sessions),
        embedder,
        store,
        similarity_top_k=settings.similarity_top_k,
        similarity_min_score=settings.similarity_min_score,
    )
    llm = build_chat_llm(settings)
    app.state.chat = ChatService(
        embedder,
        store,
        llm,
        top_k=settings.chat_top_k,
        min_score=settings.chat_min_score,
        soft_min_score=settings.chat_soft_min_score,
    )

    stop = asyncio.Event()
    task = asyncio.create_task(
        run_consumer(settings, stop, ingest.handle_event),
        name="rabbitmq-consumer",
    )
    logger.info(
        "ai_service_started rabbitmq=%s model=%s dim=%s chat_provider=%s chat_llm=%s chat_configured=%s",
        settings.rabbitmq_url,
        settings.embedding_model,
        settings.embedding_dimensions,
        settings.chat_llm_provider,
        llm.model_name,
        llm.configured,
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
            "Ingests content events into Surreal embeddings and exposes RAG chat "
            "over posts, comments, questions, and answers."
        ),
        version="0.2.0",
        lifespan=lifespan,
    )
    app.include_router(health.router)
    app.include_router(chat.router)
    return app


app = create_app()
