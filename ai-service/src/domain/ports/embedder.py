from typing import Literal, Protocol, runtime_checkable

from domain.entities import Embedding

EmbedTaskType = Literal["RETRIEVAL_DOCUMENT", "RETRIEVAL_QUERY"]

@runtime_checkable
class Embedder(Protocol):
    async def embed(
        self, text: str, *, task_type: EmbedTaskType = "RETRIEVAL_DOCUMENT"
    ) -> Embedding: ...

    async def embed_batch(
        self, texts: list[str], *, task_type: EmbedTaskType = "RETRIEVAL_DOCUMENT"
    ) -> list[Embedding]: ...
