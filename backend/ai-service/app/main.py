from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.adapters.rabbitmq.consumer import run_consumer
from app.api.routes import health
from app.config import get_settings

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s %(message)s")
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    settings = get_settings()
    stop = asyncio.Event()
    task = asyncio.create_task(run_consumer(settings, stop), name="rabbitmq-consumer")
    logger.info("ai_service_started rabbitmq=%s", settings.rabbitmq_url)
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


def create_app() -> FastAPI:
    app = FastAPI(
        title="YetBota AI Service",
        description=(
            "Post ingestion (RAG), similarity, and Surreal similarity graph. "
            "Consumes post.created from RabbitMQ; ingest/embeddings TBD."
        ),
        version="0.1.0",
        lifespan=lifespan,
    )
    app.include_router(health.router)
    return app


app = create_app()
