from typing import Protocol, runtime_checkable

from domain.entities import Chunk, ContentKind


@runtime_checkable
class Chunker(Protocol):
    def chunk(
        self,
        *,
        source_id: str,
        kind: ContentKind,
        text: str,
        metadata: dict[str, str] | None = None,
    ) -> list[Chunk]: ...
