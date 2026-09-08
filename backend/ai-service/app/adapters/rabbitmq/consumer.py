"""RabbitMQ consumer for post.created → future ingest."""

from __future__ import annotations

import asyncio
import json
import logging
from typing import Any

import aio_pika
from aio_pika.abc import AbstractIncomingMessage

from app.config import Settings

logger = logging.getLogger(__name__)

EXCHANGE_EVENTS = "yetbota.events"
ROUTING_POST_CREATED = "post.created"
QUEUE_POST_INGEST = "ai.post.ingest"


async def handle_post_created(payload: dict[str, Any]) -> None:
    """Communication stub — real Surreal embedding ingest comes later."""
    post_id = payload.get("post_id")
    logger.info("post_created_received post_id=%s event=%s", post_id, payload.get("event"))


async def _on_message(message: AbstractIncomingMessage) -> None:
    async with message.process(requeue=True):
        try:
            payload = json.loads(message.body.decode("utf-8"))
        except json.JSONDecodeError:
            logger.exception("invalid_post_created_payload")
            return
        if payload.get("event") != ROUTING_POST_CREATED:
            logger.warning("ignored_event type=%s", payload.get("event"))
            return
        await handle_post_created(payload)


async def run_consumer(settings: Settings, stop: asyncio.Event) -> None:
    """Declare topology and consume until stop is set."""
    connection = await aio_pika.connect_robust(settings.rabbitmq_url)
    async with connection:
        channel = await connection.channel()
        await channel.set_qos(prefetch_count=10)
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
        await queue.consume(_on_message)
        await stop.wait()
