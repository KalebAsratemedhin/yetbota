from datetime import UTC, datetime

from application.errors import AIServiceError, MetadataError
from domain.entities import DeleteRequest, DeleteResult
from domain.ports import SimilarityGraph, VectorStore


def _validate(req: DeleteRequest) -> None:
    if not req.content_id:
        raise MetadataError("content_id is empty")


class DeleteContent:
    def __init__(
        self,
        *,
        vector_store: VectorStore,
        similarity_graph: SimilarityGraph | None = None,
        logger: object | None = None,
    ) -> None:
        self._vector_store = vector_store
        self._similarity_graph = similarity_graph
        self._logger = logger

    async def execute(self, req: DeleteRequest) -> DeleteResult:
        _validate(req)

        await self._vector_store.delete_by_source(req.content_id, req.kind)

        if req.kind == "post":
            await self._delete_graph_post(req.content_id)

        return DeleteResult(
            content_id=req.content_id,
            kind=req.kind,
            deleted=True,
            error_code=None,
            processed_at=datetime.now(UTC),
        )

    async def _delete_graph_post(self, post_id: str) -> None:
        if self._similarity_graph is None:
            return
        try:
            await self._similarity_graph.delete_post(post_id)
        except AIServiceError as exc:
            if self._logger is not None:
                self._logger.warning(  # type: ignore[attr-defined]
                    "delete.graph_cleanup_failed",
                    post_id=post_id,
                    error_code=exc.code,
                )


__all__ = ["DeleteContent"]
