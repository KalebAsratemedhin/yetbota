import re
import unicodedata

from domain.entities import Chunk, ContentKind
from infrastructure.config.settings import ChunkerSettings

_WHITESPACE = re.compile(r"\s+")


def _normalize(text: str) -> str:
    cleaned = "".join(ch for ch in text if ch == "\n" or unicodedata.category(ch)[0] != "C")
    return _WHITESPACE.sub(" ", cleaned).strip()


class RuneWindowChunker:
    def __init__(self, settings: ChunkerSettings) -> None:
        if settings.size <= 0:
            raise ValueError("chunker size must be positive")
        if settings.overlap < 0 or settings.overlap >= settings.size:
            raise ValueError("chunker overlap must satisfy 0 <= overlap < size")
        self._size = settings.size
        self._overlap = settings.overlap

    def chunk(
        self,
        *,
        source_id: str,
        kind: ContentKind,
        text: str,
        metadata: dict[str, str] | None = None,
    ) -> list[Chunk]:
        normalized = _normalize(text)
        if not normalized:
            return []

        meta = dict(metadata or {})
        if len(normalized) <= self._size:
            return [
                Chunk(
                    source_id=source_id,
                    kind=kind,
                    text=normalized,
                    metadata=meta,
                    segment_idx=0,
                )
            ]

        step = self._size - self._overlap
        chunks: list[Chunk] = []
        idx = 0
        start = 0
        while start < len(normalized):
            end = start + self._size
            window = normalized[start:end]
            chunks.append(
                Chunk(
                    source_id=source_id,
                    kind=kind,
                    text=window,
                    metadata=meta,
                    segment_idx=idx,
                )
            )
            if end >= len(normalized):
                break
            start += step
            idx += 1
        return chunks
