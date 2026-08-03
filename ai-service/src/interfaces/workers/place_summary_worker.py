from __future__ import annotations

import asyncio
import json

from application.errors import AIServiceError, EmbeddingFailed, LLMUnavailable, SimilaritySearchFailed
from application.summarize_place import SummarizePlace
from infrastructure.config.settings import RabbitMQSettings
from infrastructure.content_service import ContentServiceClient
from infrastructure.messaging.queues import QUEUE_PLACE_SUMMARIZE
from infrastructure.messaging.rabbitmq import RabbitMQConsumer
from infrastructure.observability import get_logger

logger = get_logger(__name__)


class PlaceSummaryWorker:
    def __init__(
        self,
        *,
        consumer: RabbitMQConsumer,
        use_case: SummarizePlace,
        content_client: ContentServiceClient,
        settings: RabbitMQSettings,
    ) -> None:
        self._consumer = consumer
        self._use_case = use_case
        self._content_client = content_client
        self._settings = settings

    async def run(self, stop_event: asyncio.Event) -> None:
        logger.info("place_summary_worker.started", queue=QUEUE_PLACE_SUMMARIZE)
        await self._consumer.consume(self._on_message)
        try:
            await stop_event.wait()
        finally:
            await self._consumer.shutdown()
            logger.info("place_summary_worker.stopped", queue=QUEUE_PLACE_SUMMARIZE)

    async def _on_message(self, message) -> None:
        try:
            data = json.loads(message.body.decode("utf-8"))
            place_id = str(data.get("place_id") or "").strip()
        except (json.JSONDecodeError, ValueError, AttributeError) as exc:
            logger.warning("place_summary.malformed", error=str(exc))
            return
        if not place_id:
            logger.warning("place_summary.missing_place_id")
            return

        try:
            result = await self._use_case.execute(place_id)
        except (EmbeddingFailed, SimilaritySearchFailed, LLMUnavailable) as exc:
            if exc.transient:
                raise
            logger.warning(
                "place_summary.failed",
                place_id=place_id,
                error_code=exc.code,
                message=exc.message,
            )
            return
        except AIServiceError as exc:
            logger.warning(
                "place_summary.failed",
                place_id=place_id,
                error_code=exc.code,
                message=exc.message,
            )
            return

        if result.skipped:
            logger.info(
                "place_summary.skipped",
                place_id=place_id,
                reason=result.skip_reason,
            )
            return

        await self._content_client.patch_place_summary(
            place_id=place_id,
            summary_text=result.summary_text,
            citations=[
                {
                    "contribution_id": c.contribution_id,
                    "kind": c.kind,
                    "excerpt": c.excerpt,
                }
                for c in result.citations
            ],
        )
        logger.info("place_summary.updated", place_id=place_id, citations=len(result.citations))
