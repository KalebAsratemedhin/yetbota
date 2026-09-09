"""RabbitMQ consumer for post.created → ingest."""

from __future__ import annotations

import asyncio
import json
import logging
from collections.abc import Awaitable, Callable
from typing import Any
from uuid import UUID

import aio_pika
from aio_pika.abc import AbstractIncomingMessage

from app.config import Settings

logger = logging.getLogger(__name__)

EXCHANGE_EVENTS = "yetbota.events"
ROUTING_POST_CREATED = "post.created"
QUEUE_POST_INGEST = "ai.post.ingest"

IngestHandler = Callable[[UUID], Awaitable[bool]]


async def _on_message(
    message: AbstractIncomingMessage,
    ingest: IngestHandler,
) -> None:
    async with message.process(requeue=True):
        try:
            payload = json.loads(message.body.decode("utf-8"))
        except json.JSONDecodeError:
            logger.exception("invalid_post_created_payload")
            return
        if payload.get("event") != ROUTING_POST_CREATED:
            logger.warning("ignored_event type=%s", payload.get("event"))
            return
        raw_id = payload.get("post_id")
        try:
            post_id = UUID(str(raw_id))
        except (TypeError, ValueError):
            logger.error("invalid_post_id value=%s", raw_id)
            return
        try:
            await ingest(post_id)
        except Exception:
            logger.exception("ingest_failed post_id=%s", post_id)
            raise  # requeue via process(requeue=True) on exception? 
            # aio_pika process() nacks/requeues if exception escapes before exit
            # Actually `async with message.process(requeue=True)` requeues on exception.


async def run_consumer(
    settings: Settings,
    stop: asyncio.Event,
    ingest: IngestHandler,
) -> None:
    """Declare topology and consume until stop is set."""
    connection = await aio_pika.connect_robust(settings.rabbitmq_url)
    async with connection:
        channel = await connection.channel()
        await channel.set_qos(prefetch_count=2)
        exchange = await channel.declare_exchange(
            EXCHANGE_EVENTS, aio_pika.ExchangeType.TOPIC, durable=True
        )
        queue = await channel.declare_queue(QUEUE_POST_INGEST, durable=True)
        await queue.bind(exchange, routing_key=ROUTING_POST_CREATED)
        logger.info(
            "rabbitmq_consumer_ready queue=%s routing_key=%s",
            QUEUE_POST_INGEST,
            ROUTING_POST_CREATED,
        )

        async def handler(message: AbstractIncomingMessage) -> None:
            await _on_message(message, ingest)

        await queue.consume(handler)
        await stop.wait()
