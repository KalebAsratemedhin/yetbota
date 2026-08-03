from __future__ import annotations

import asyncio
import json
from datetime import UTC, datetime

from application.errors import (
    AIServiceError,
    EmbeddingFailed,
    IndexingFailed,
    MessageMalformed,
    MetadataError,
    SimilaritySearchFailed,
)
from application.ingest_content import IngestContent
from domain.entities import IncomingMessage, IngestRequest, IngestResult
from infrastructure.config.settings import RabbitMQSettings
from infrastructure.messaging.rabbitmq import RabbitMQConsumer, RabbitMQPublisher
from infrastructure.observability import get_logger

logger = get_logger(__name__)

_TERMINAL_ERRORS: tuple[type[AIServiceError], ...] = (
    MessageMalformed,
    MetadataError,
    IndexingFailed,
)


class IngestWorker:
    def __init__(
        self,
        *,
        consumer: RabbitMQConsumer,
        publisher: RabbitMQPublisher,
        use_case: IngestContent,
        settings: RabbitMQSettings,
    ) -> None:
        self._consumer = consumer
        self._publisher = publisher
        self._use_case = use_case
        self._settings = settings

    async def run(self, stop_event: asyncio.Event) -> None:
        logger.info("ingest_worker.started", queue=self._settings.ingest_queue)
        await self._consumer.consume(self._on_message)
        try:
            await stop_event.wait()
        finally:
            await self.shutdown()
            logger.info("ingest_worker.stopped", queue=self._settings.ingest_queue)

    async def shutdown(self) -> None:
        await self._consumer.shutdown()

    async def _on_message(self, message) -> None:
        incoming = IncomingMessage(
            body=message.body,
            delivery_count=_delivery_count_from_message(message),
            reply_to=message.reply_to,
            correlation_id=message.correlation_id,
        )
        await self._handle(incoming)

    async def _handle(self, message: IncomingMessage) -> None:
        try:
            req = _parse_request(message.body)
        except MessageMalformed as exc:
            await self._publish_error(message, content_id="", kind="post", error=exc)
            return

        try:
            result = await self._use_case.execute(req)
        except _TERMINAL_ERRORS as exc:
            await self._publish_error(message, content_id=req.content_id, kind=req.kind, error=exc)
            return
        except (EmbeddingFailed, SimilaritySearchFailed) as exc:
            if exc.transient and message.delivery_count < self._settings.max_delivery_attempts:
                logger.warning(
                    "ingest.retrying",
                    error_code=exc.code,
                    delivery_count=message.delivery_count,
                    max_delivery_attempts=self._settings.max_delivery_attempts,
                    content_id=req.content_id,
                    kind=req.kind,
                    message=exc.message,
                )
                raise
            logger.warning(
                "ingest.terminal_error",
                error_code=exc.code,
                transient=exc.transient,
                content_id=req.content_id,
                kind=req.kind,
                message=exc.message,
            )
            await self._publish_error(message, content_id=req.content_id, kind=req.kind, error=exc)
            return
        except AIServiceError as exc:
            await self._publish_error(message, content_id=req.content_id, kind=req.kind, error=exc)
            return

        if message.reply_to and message.correlation_id:
            await self._publisher.publish_reply(
                reply_to=message.reply_to,
                correlation_id=message.correlation_id,
                body=_serialize_result(result).encode("utf-8"),
            )

    async def _publish_error(
        self,
        message: IncomingMessage,
        *,
        content_id: str,
        kind: str,
        error: AIServiceError,
    ) -> None:
        if not (message.reply_to and message.correlation_id):
            return
        result = IngestResult(
            content_id=content_id or "unknown",
            kind=kind,  # type: ignore[arg-type]
            verdict="error",
            error_code=error.code,
            processed_at=datetime.now(UTC),
        )
        await self._publisher.publish_reply(
            reply_to=message.reply_to,
            correlation_id=message.correlation_id,
            body=_serialize_result(result).encode("utf-8"),
        )


def _parse_request(body: bytes) -> IngestRequest:
    try:
        data = json.loads(body.decode("utf-8"))
        return IngestRequest.model_validate(data)
    except (json.JSONDecodeError, ValueError) as exc:
        raise MessageMalformed("invalid ingest payload") from exc


def _serialize_result(result: IngestResult) -> str:
    payload: dict[str, object] = {
        "content_id": result.content_id,
        "kind": result.kind,
        "status": result.verdict,
        "duplicate_of": result.duplicate_of,
        "error_code": result.error_code,
        "processed_at": result.processed_at.isoformat(),
    }
    return json.dumps(payload)


def _delivery_count_from_message(message) -> int:
    headers = message.headers or {}
    retry_count = headers.get("x-retry-count")
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
