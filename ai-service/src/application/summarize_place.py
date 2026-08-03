import re
from collections.abc import Callable

from domain.entities import Chunk
from domain.entities.place_summary import PlaceSummaryCitation, PlaceSummaryResult
from domain.ports import LLM, VectorStore

_CITATION_PATTERN = re.compile(r"\[(\d+)\]")

PromptBuilder = Callable[[list[Chunk]], str]


class SummarizePlace:
    def __init__(
        self,
        *,
        vector_store: VectorStore,
        llm: LLM,
        prompt_builder: PromptBuilder,
        min_contributions: int,
        max_source_contributions: int,
        max_tokens: int,
        temperature: float,
    ) -> None:
        self._vector_store = vector_store
        self._llm = llm
        self._prompt_builder = prompt_builder
        self._min_contributions = min_contributions
        self._max_source_contributions = max_source_contributions
        self._max_tokens = max_tokens
        self._temperature = temperature

    async def execute(self, place_id: str) -> PlaceSummaryResult:
        chunks = await self._vector_store.list_by_place(
            place_id, limit=self._max_source_contributions
        )
        # Prefer contribution-like kinds for the min threshold; place description alone is not enough.
        contrib_chunks = [c for c in chunks if c.kind != "place"]
        if len(contrib_chunks) < self._min_contributions:
            return PlaceSummaryResult(
                place_id=place_id,
                summary_text="",
                skipped=True,
                skip_reason="below_min_contributions",
            )
        if not chunks:
            return PlaceSummaryResult(
                place_id=place_id,
                summary_text="",
                skipped=True,
                skip_reason="no_chunks",
            )

        prompt = self._prompt_builder(chunks)
        answer = await self._llm.generate(
            prompt,
            max_tokens=self._max_tokens,
            temperature=self._temperature,
        )
        answer = answer.strip()
        citations = _citations_from(answer, chunks)
        return PlaceSummaryResult(
            place_id=place_id,
            summary_text=answer,
            citations=citations,
        )


def _citations_from(answer: str, chunks: list[Chunk]) -> list[PlaceSummaryCitation]:
    seen: set[int] = set()
    out: list[PlaceSummaryCitation] = []
    for match in _CITATION_PATTERN.finditer(answer):
        idx = int(match.group(1)) - 1
        if 0 <= idx < len(chunks) and idx not in seen:
            seen.add(idx)
            chunk = chunks[idx]
            excerpt = chunk.text.strip()
            if len(excerpt) > 240:
                excerpt = excerpt[:237] + "..."
            out.append(
                PlaceSummaryCitation(
                    contribution_id=chunk.source_id,
                    kind=chunk.kind,
                    excerpt=excerpt,
                )
            )
    return out
