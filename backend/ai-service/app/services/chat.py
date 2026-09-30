"""RAG chat: embed query → retrieve content_embedding → LLM answer with sources."""

from __future__ import annotations

import logging
from typing import Protocol, Sequence

from app.domain.chat import ContentHit
from app.domain.ports import EmbeddingProvider

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = """You are YetBota's local guide assistant for crowdsourced place discovery.
Answer using ONLY the provided context snippets (posts, comments, Q&A about places).
Prefer the closest relevant matches — do not require exact wording or a perfect match.
If several snippets are related, summarize the best ones. If none are on-topic, say so
briefly and suggest what the user could ask instead.
Be concise, friendly, and specific about places (use titles / place names from context).
Never invent venues, hours, or facts that are not in the context.
Never mention internal IDs, UUIDs, database keys, or phrases like "post ID …".
When helpful, say whether you used a post, comment, question, or answer — by name/topic only."""


class ContentSearcher(Protocol):
    async def search_content(
        self,
        embedding: Sequence[float],
        *,
        top_k: int,
        min_score: float,
    ) -> list[ContentHit]: ...


class ChatLLM(Protocol):
    @property
    def configured(self) -> bool: ...

    @property
    def model_name(self) -> str: ...

    async def complete(
        self,
        *,
        system: str,
        messages: list[dict[str, str]],
        json_mode: bool = False,
    ) -> str: ...


class ChatService:
    def __init__(
        self,
        embedder: EmbeddingProvider,
        searcher: ContentSearcher,
        llm: ChatLLM,
        *,
        top_k: int = 8,
        min_score: float = 0.22,
        soft_min_score: float = 0.12,
    ) -> None:
        self._embedder = embedder
        self._searcher = searcher
        self._llm = llm
        self._top_k = top_k
        self._min_score = min_score
        self._soft_min_score = soft_min_score

    async def chat(
        self,
        message: str,
        *,
        history: list[dict[str, str]] | None = None,
    ) -> dict:
        text = (message or "").strip()
        if not text:
            raise ValueError("message is required")
        if not self._llm.configured:
            raise RuntimeError(
                "chat LLM is not configured (set GEMINI_API_KEY or OPENAI_API_KEY)"
            )

        embedding = await self._embedder.embed(text)
        hits = await self._retrieve(embedding)
        context = _format_context(hits)
        user_block = (
            f"Context:\n{context}\n\nUser question:\n{text}"
            if context
            else f"Context:\n(none)\n\nUser question:\n{text}"
        )

        messages: list[dict[str, str]] = []
        for turn in history or []:
            role = (turn.get("role") or "").strip()
            content = (turn.get("content") or "").strip()
            if role in ("user", "assistant") and content:
                messages.append({"role": role, "content": content})
        messages.append({"role": "user", "content": user_block})

        reply = await self._llm.complete(system=SYSTEM_PROMPT, messages=messages)
        logger.info(
            "chat_ok model=%s hits=%s reply_chars=%s",
            self._llm.model_name,
            len(hits),
            len(reply),
        )
        return {
            "reply": reply,
            "model": self._llm.model_name,
            "sources": [_source_dict(h) for h in hits],
        }

    async def _retrieve(self, embedding: Sequence[float]) -> list[ContentHit]:
        """Prefer hits above min_score; else closest above soft_min_score."""
        floor = min(self._soft_min_score, self._min_score)
        candidates = await self._searcher.search_content(
            embedding, top_k=self._top_k, min_score=floor
        )
        if not candidates:
            return []
        strict = [h for h in candidates if h.score >= self._min_score]
        if strict:
            return strict
        logger.info(
            "chat_soft_retrieve hits=%s top_score=%.3f floor=%.3f",
            len(candidates),
            candidates[0].score,
            floor,
        )
        return candidates


def _format_context(hits: list[ContentHit]) -> str:
    if not hits:
        return ""
    parts: list[str] = []
    for i, h in enumerate(hits, start=1):
        title = f" — {h.title}" if h.title else ""
        parts.append(f"[{i}] kind={h.kind}{title}\n{h.text}")
    return "\n\n".join(parts)


def _source_dict(h: ContentHit) -> dict:
    out: dict = {
        "kind": h.kind,
        "entity_id": str(h.entity_id),
        "score": round(h.score, 4),
        "title": h.title,
        "text": h.text[:500],
    }
    if h.parent_id is not None:
        out["parent_id"] = str(h.parent_id)
    if h.post_id is not None:
        out["post_id"] = str(h.post_id)
    return out
