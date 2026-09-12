"""RabbitMQ consumer for content *.created → typed ingest."""

from __future__ import annotations

import asyncio
import json
import logging
from collections.abc import Awaitable, Callable
from typing import Any

import aio_pika
from aio_pika.abc import AbstractIncomingMessage

from app.config import Settings

logger = logging.getLogger(__name__)

EXCHANGE_EVENTS = "yetbota.events"
QUEUE_CONTENT_INGEST = "ai.content.ingest"

ROUTING_KEYS = (
    "post.created",
    "comment.created",
    "question.created",
    "answer.created",
)

EventHandler = Callable[[str, dict[str, Any]], Awaitable[bool]]


async def _on_message(
    message: AbstractIncomingMessage,
    handle: EventHandler,
) -> None:
    async with message.process(requeue=True):
        try:
            payload = json.loads(message.body.decode("utf-8"))
        except json.JSONDecodeError:
            logger.exception("invalid_content_event_payload")
            return
        event = str(payload.get("event") or "")
        if event not in ROUTING_KEYS:
            logger.warning("ignored_event type=%s", event)
            return
        try:
            await handle(event, payload)
        except (KeyError, TypeError, ValueError):
            logger.exception("invalid_content_event event=%s payload=%s", event, payload)
            return
        except Exception:
            logger.exception("ingest_failed event=%s", event)
            raise


async def run_consumer(
    settings: Settings,
    stop: asyncio.Event,
    handle: EventHandler,
) -> None:
    """Declare topology and consume until stop is set."""
    connection = await aio_pika.connect_robust(settings.rabbitmq_url)
    async with connection:
        channel = await connection.channel()
        await channel.set_qos(prefetch_count=2)
        exchange = await channel.declare_exchange(
            EXCHANGE_EVENTS, aio_pika.ExchangeType.TOPIC, durable=True
        )
        queue = await channel.declare_queue(QUEUE_CONTENT_INGEST, durable=True)
        for key in ROUTING_KEYS:
            await queue.bind(exchange, routing_key=key)
        logger.info(
            "rabbitmq_consumer_ready queue=%s routing_keys=%s",
            QUEUE_CONTENT_INGEST,
            ",".join(ROUTING_KEYS),
        )

        async def handler(message: AbstractIncomingMessage) -> None:
            await _on_message(message, handle)

        await queue.consume(handler)
        await stop.wait()
