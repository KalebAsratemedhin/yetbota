import re
from collections.abc import Callable

from domain.entities import (
    ChatQuery,
    ChatResponse,
    Citation,
    Message,
    ScoredChunk,
)
from domain.ports import LLM, Embedder, VectorStore

_NO_INFO_ANSWER = "I do not have verified information on that topic in YetBota."

_CITATION_PATTERN = re.compile(r"\[(\d+)\]")

_RETRIEVAL_CONTEXT_TURNS = 2

PromptBuilder = Callable[[list[ScoredChunk], str, list[Message]], str]


def _retrieval_text(query: str, history: list[Message]) -> str:
    prior_user = [m.content for m in history if m.role == "user"]
    parts = [*prior_user[-_RETRIEVAL_CONTEXT_TURNS:], query]
    return "\n".join(p for p in parts if p.strip())


def _cited_hits(answer: str, hits: list[ScoredChunk]) -> list[ScoredChunk]:
    seen: set[int] = set()
    cited: list[ScoredChunk] = []
    for match in _CITATION_PATTERN.finditer(answer):
        idx = int(match.group(1)) - 1
        if 0 <= idx < len(hits) and idx not in seen:
            seen.add(idx)
            cited.append(hits[idx])
    return cited


def _citations_from(hits: list[ScoredChunk]) -> list[Citation]:
    return [
        Citation(
            source_id=h.chunk.source_id,
            kind=h.chunk.kind,
            text=h.chunk.text,
            score=h.score,
        )
        for h in hits
    ]


class ConsultAssistant:
    def __init__(
        self,
        *,
        embedder: Embedder,
        vector_store: VectorStore,
        llm: LLM,
        prompt_builder: PromptBuilder,
        top_k: int,
        min_similarity: float,
        max_tokens: int,
        temperature: float,
    ) -> None:
        self._embedder = embedder
        self._vector_store = vector_store
        self._llm = llm
        self._prompt_builder = prompt_builder
        self._top_k = top_k
        self._min_similarity = min_similarity
        self._max_tokens = max_tokens
        self._temperature = temperature

    async def execute(self, query: ChatQuery) -> ChatResponse:
        retrieval_text = _retrieval_text(query.text, query.history)
        query_vec = await self._embedder.embed(retrieval_text, task_type="RETRIEVAL_QUERY")
        radius_m = None
        if query.radius_km is not None:
            radius_m = query.radius_km * 1000.0
        hits = await self._vector_store.search(
            query_vec,
            limit=self._top_k,
            kinds=None,
            place_id=query.place_id,
            lat=query.latitude,
            lon=query.longitude,
            radius_m=radius_m,
        )

        if not hits or max(h.score for h in hits) < self._min_similarity:
            return ChatResponse(answer=_NO_INFO_ANSWER, citations=[])

        prompt = self._prompt_builder(hits, query.text, query.history)
        answer = await self._llm.generate(
            prompt,
            max_tokens=self._max_tokens,
            temperature=self._temperature,
        )
        answer = answer.strip()
        return ChatResponse(answer=answer, citations=_citations_from(_cited_hits(answer, hits)))
