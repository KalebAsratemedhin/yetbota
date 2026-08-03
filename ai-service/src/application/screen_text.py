import asyncio
import hashlib
import time
from collections.abc import Callable
from typing import Any, get_args

from application.errors import LLMUnavailable
from domain.entities import ScreeningCategory, ScreeningRequest, ScreeningResult
from domain.entities.chunk import ContentKind
from domain.ports import LLM

ScreeningPromptBuilder = Callable[[str, ContentKind], str]
_VALID_CATEGORIES: tuple[ScreeningCategory, ...] = get_args(ScreeningCategory)


def _cache_key(text: str, kind: ContentKind) -> str:
    h = hashlib.sha256()
    h.update(kind.encode("utf-8"))
    h.update(b"\x00")
    h.update(text.encode("utf-8"))
    return h.hexdigest()


def _coerce_categories(parsed: dict[str, Any]) -> dict[ScreeningCategory, float]:
    out: dict[ScreeningCategory, float] = {}
    for category in _VALID_CATEGORIES:
        raw = parsed.get(category, 0.0)
        try:
            out[category] = float(raw)
        except (TypeError, ValueError):
            out[category] = 0.0
    return out


class _TTLCache:
    def __init__(self, maxsize: int, ttl_s: int) -> None:
        self._maxsize = maxsize
        self._ttl_s = ttl_s
        self._store: dict[str, tuple[float, ScreeningResult]] = {}

    def get(self, key: str) -> ScreeningResult | None:
        entry = self._store.get(key)
        if entry is None:
            return None
        expires_at, value = entry
        if expires_at < time.monotonic():
            self._store.pop(key, None)
            return None
        return value

    def set(self, key: str, value: ScreeningResult) -> None:
        if len(self._store) >= self._maxsize and key not in self._store:
            self._store.pop(next(iter(self._store)), None)
        self._store[key] = (time.monotonic() + self._ttl_s, value)


class ScreenText:
    def __init__(
        self,
        *,
        llm: LLM,
        prompt_builder: ScreeningPromptBuilder,
        response_schema: dict[str, Any],
        block_threshold: float,
        timeout_s: float,
        cache_size: int,
        cache_ttl_s: int,
    ) -> None:
        self._llm = llm
        self._prompt_builder = prompt_builder
        self._schema = response_schema
        self._block_threshold = block_threshold
        self._timeout_s = timeout_s
        self._cache = _TTLCache(cache_size, cache_ttl_s)

    async def execute(self, req: ScreeningRequest) -> ScreeningResult:
        key = _cache_key(req.text, req.kind)
        cached = self._cache.get(key)
        if cached is not None:
            return cached

        try:
            parsed = await asyncio.wait_for(
                self._llm.classify(
                    self._prompt_builder(req.text, req.kind),
                    schema=self._schema,
                ),
                timeout=self._timeout_s,
            )
        except (TimeoutError, LLMUnavailable):
            return ScreeningResult(
                ok=False,
                reason="screening_unavailable",
                categories={},
            )

        categories = _coerce_categories(parsed)
        worst_score = max(categories.values()) if categories else 0.0
        ok = worst_score < self._block_threshold
        reason = None if ok else max(categories, key=lambda c: categories[c])
        result = ScreeningResult(ok=ok, reason=reason, categories=categories)
        self._cache.set(key, result)
        return result
