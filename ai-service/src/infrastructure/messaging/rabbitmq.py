from __future__ import annotations

import asyncio
from collections.abc import Awaitable, Callable

import aio_pika
from aio_pika.abc import AbstractChannel, AbstractIncomingMessage, AbstractRobustConnection

from application.errors import AIServiceError
from infrastructure.config.settings import RabbitMQSettings
from infrastructure.messaging.queues import (
    DEAD_LETTER_EXCHANGE,
    EXCHANGE_NAME,
    QUEUE_AI_INGEST,
    QUEUE_AI_INGEST_DLQ,
    QUEUE_PLACE_SUMMARIZE,
    QUEUE_PLACE_SUMMARIZE_DLQ,
)
from infrastructure.observability import get_logger

logger = get_logger(__name__)

Handler = Callable[[AbstractIncomingMessage], Awaitable[None]]

_RETRY_HEADER = "x-retry-count"
_MAX_RETRY_BACKOFF_S = 30.0


def _queue_args(dlq_routing_key: str) -> dict[str, str]:
    return {
        "x-dead-letter-exchange": DEAD_LETTER_EXCHANGE,
        "x-dead-letter-routing-key": dlq_routing_key,
    }


class RabbitMQConnection:
    def __init__(self, settings: RabbitMQSettings) -> None:
        self._settings = settings
        self._connection: AbstractRobustConnection | None = None
        self._channel: AbstractChannel | None = None

    async def connect(self) -> AbstractChannel:
        self._connection = await aio_pika.connect_robust(self._settings.url)
        self._channel = await self._connection.channel()
        await self._channel.set_qos(prefetch_count=self._settings.prefetch_count)
        await self._declare_topology(self._channel)
        return self._channel

    async def close(self) -> None:
        if self._channel is not None:
            await self._channel.close()
            self._channel = None
        if self._connection is not None:
            await self._connection.close()
            self._connection = None

    @staticmethod
    async def _declare_topology(channel: AbstractChannel) -> None:
        exchange = await channel.declare_exchange(EXCHANGE_NAME, aio_pika.ExchangeType.DIRECT, durable=True)
        dlx = await channel.declare_exchange(
            DEAD_LETTER_EXCHANGE, aio_pika.ExchangeType.DIRECT, durable=True
        )

        for queue_name, dlq_name in (
            (QUEUE_AI_INGEST, QUEUE_AI_INGEST_DLQ),
            (QUEUE_PLACE_SUMMARIZE, QUEUE_PLACE_SUMMARIZE_DLQ),
        ):
            dlq = await channel.declare_queue(dlq_name, durable=True)
            await dlq.bind(dlx, routing_key=dlq_name)
            queue = await channel.declare_queue(
                queue_name,
                durable=True,
                arguments=_queue_args(dlq_name),
            )
            await queue.bind(exchange, routing_key=queue_name)

    @property
    def channel(self) -> AbstractChannel:
        if self._channel is None:
            raise RuntimeError("RabbitMQ connection is not established")
        return self._channel


class RabbitMQConsumer:
    def __init__(
        self,
        connection: RabbitMQConnection,
        settings: RabbitMQSettings,
        *,
        queue_name: str = QUEUE_AI_INGEST,
        dlq_name: str = QUEUE_AI_INGEST_DLQ,
    ) -> None:
        self._connection = connection
        self._settings = settings
        self._consume_tag: str | None = None
        self._queue_name = queue_name
        self._dlq_name = dlq_name

    async def consume(self, handler: Handler) -> None:
        channel = self._connection.channel
        queue = await channel.declare_queue(
            self._queue_name,
            durable=True,
            arguments=_queue_args(self._dlq_name),
        )
        self._consume_tag = await queue.consume(self._wrap(handler, channel))

    async def shutdown(self) -> None:
        if self._consume_tag is not None:
            channel = self._connection.channel
            await channel.cancel(self._consume_tag)
            self._consume_tag = None

    def _wrap(self, handler: Handler, channel: AbstractChannel) -> Handler:
        async def _inner(message: AbstractIncomingMessage) -> None:
            try:
                await handler(message)
                await message.ack()
            except Exception as exc:
                transient = isinstance(exc, AIServiceError) and exc.transient
                count = _delivery_count(message)
                if transient and count < self._settings.max_delivery_attempts:
                    delay_s = min(_MAX_RETRY_BACKOFF_S, 2 ** max(count - 1, 0))
                    logger.warning(
                        "rabbitmq.retry_scheduled",
                        queue=self._queue_name,
                        delivery_count=count,
                        max_delivery_attempts=self._settings.max_delivery_attempts,
                        delay_s=delay_s,
                        error_type=type(exc).__name__,
                        error_code=getattr(exc, "code", None),
                        transient=transient,
                    )
                    await asyncio.sleep(delay_s)
                    headers = dict(message.headers or {})
                    headers[_RETRY_HEADER] = count
                    exchange = await channel.declare_exchange(
                        EXCHANGE_NAME, aio_pika.ExchangeType.DIRECT, durable=True
                    )
                    await exchange.publish(
                        aio_pika.Message(
                            body=message.body,
                            headers=headers,
                            content_type=message.content_type or "application/json",
                            correlation_id=message.correlation_id,
                            reply_to=message.reply_to,
                            delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
                        ),
                        routing_key=self._queue_name,
                    )
                    await message.ack()
                    return
                logger.warning(
                    "rabbitmq.dead_lettered",
                    queue=self._queue_name,
                    delivery_count=count,
                    max_delivery_attempts=self._settings.max_delivery_attempts,
                    error_type=type(exc).__name__,
                    error_code=getattr(exc, "code", None),
                    transient=transient,
                )
                await message.nack(requeue=False)

        return _inner


class RabbitMQPublisher:
    def __init__(self, connection: RabbitMQConnection) -> None:
        self._connection = connection

    async def publish_reply(
        self,
        *,
        reply_to: str,
        correlation_id: str,
        body: bytes,
    ) -> None:
        channel = self._connection.channel
        await channel.default_exchange.publish(
            aio_pika.Message(
                body=body,
                content_type="application/json",
                correlation_id=correlation_id,
                delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
            ),
            routing_key=reply_to,
        )


def _delivery_count(message: AbstractIncomingMessage) -> int:
    headers = message.headers or {}
    retry_count = headers.get(_RETRY_HEADER)
    if isinstance(retry_count, int) and retry_count >= 1:
        return retry_count + 1
    deaths = headers.get("x-death")
    if isinstance(deaths, list) and deaths:
        total = sum(
            death.get("count", 0)
            for death in deaths
            if isinstance(death, dict) and isinstance(death.get("count"), int)
        )
        if total > 0:
            return total + 1
    return 1
