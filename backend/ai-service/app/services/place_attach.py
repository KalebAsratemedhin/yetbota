"""Attach posts to Places (same-place match) + AI tip/change/important + summary."""

from __future__ import annotations

import asyncio
import json
import logging
import math
import re
from difflib import SequenceMatcher
from typing import Any
from uuid import UUID

from app.adapters.content.places import ContentPlacesClient
from app.domain.post_doc import PostDocument
from app.services.chat import ChatLLM

logger = logging.getLogger(__name__)

CLASSIFY_SYSTEM = """You extract place-directory facts from one crowdsourced post.
Return ONLY valid JSON: {"entries":[{"category":"tip"|"change"|"important","body":"..."}]}
- tip: practical advice (what to try, how to visit)
- change: something that changed (hours, closed, moved, menu)
- important: safety/access/must-know
Use only facts present in the post. No UUIDs or IDs. Empty entries array if nothing fits.
Max 5 entries. Each body at most 240 characters."""

SUMMARY_SYSTEM = """You write a short place-directory summary for YetBota.
Use ONLY the provided post snippets and directory entries.
Be concise (2–4 sentences). Name the place. No UUIDs or internal IDs.
Do not invent facts. If little info, say what is known briefly."""


class PlaceAttachService:
    def __init__(
        self,
        places: ContentPlacesClient,
        llm: ChatLLM,
        *,
        radius_m: float = 150.0,
        embed_min: float = 0.78,
        name_min: float = 0.55,
        ambiguity_margin: float = 0.05,
        summary_debounce_sec: float = 10.0,
    ) -> None:
        self._places = places
        self._llm = llm
        self._radius_m = radius_m
        self._embed_min = embed_min
        self._name_min = name_min
        self._ambiguity_margin = ambiguity_margin
        self._debounce_sec = summary_debounce_sec
        self._summary_tasks: dict[str, asyncio.Task] = {}
        self._lock = asyncio.Lock()

    async def after_post_ingest(
        self,
        post: PostDocument,
        embedding: list[float],
    ) -> None:
        """Match/create place, then classify + summary independently."""
        place: dict[str, Any] | None = None
        try:
            existing = await self._places.get_by_post(post.id)
            if existing:
                place = existing
                logger.info(
                    "place_already_attached post_id=%s place_id=%s",
                    post.id,
                    place.get("id"),
                )
            else:
                place = await self._match_or_create(post, embedding)
        except Exception:
            logger.exception("place_attach_failed post_id=%s", post.id)
            return

        place_id = UUID(str(place["id"]))
        needs_immediate = not str(place.get("summary") or "").strip()

        try:
            await self._classify_and_store(place_id, post)
        except Exception:
            logger.exception(
                "place_classify_failed post_id=%s place_id=%s", post.id, place_id
            )

        try:
            self._schedule_summary(place_id, post, immediate=needs_immediate)
        except Exception:
            logger.exception(
                "place_summary_schedule_failed post_id=%s place_id=%s",
                post.id,
                place_id,
            )

    async def _match_or_create(
        self,
        post: PostDocument,
        embedding: list[float],
    ) -> dict[str, Any]:
        candidates = await self._places.nearby(
            lat=post.lat,
            lng=post.lng,
            radius_m=self._radius_m,
            limit=25,
        )
        scored: list[tuple[float, dict[str, Any]]] = []
        for c in candidates:
            name_s = _name_score(post.title, str(c.get("name") or ""))
            emb = c.get("embedding") or []
            emb_s = _cosine(embedding, emb) if emb else 0.0
            if name_s >= self._name_min or emb_s >= self._embed_min:
                score = max(name_s, emb_s)
                scored.append((score, c))
        scored.sort(key=lambda x: x[0], reverse=True)

        if scored:
            best_score, best = scored[0]
            if len(scored) > 1:
                second = scored[1][0]
                if best_score - second < self._ambiguity_margin:
                    logger.info(
                        "place_match_ambiguous post_id=%s best=%.3f second=%.3f → create",
                        post.id,
                        best_score,
                        second,
                    )
                    return await self._places.attach(
                        post_id=post.id,
                        name=post.title,
                        lat=post.lat,
                        lng=post.lng,
                        attach_score=0.0,
                        attach_method="created_ambiguous",
                        embedding=embedding,
                    )
            centroid = _blend_embedding(best.get("embedding") or [], embedding)
            return await self._places.attach(
                post_id=post.id,
                place_id=UUID(str(best["id"])),
                attach_score=best_score,
                attach_method="geo_name_embed",
                embedding=centroid or embedding,
            )

        return await self._places.attach(
            post_id=post.id,
            name=post.title,
            lat=post.lat,
            lng=post.lng,
            attach_score=0.0,
            attach_method="created_seed",
            embedding=embedding,
        )

    async def _classify_and_store(self, place_id: UUID, post: PostDocument) -> None:
        if not self._llm.configured:
            logger.warning("place_classify_skipped llm_not_configured")
            return
        text = post.embed_text()
        raw = await self._llm.complete(
            system=CLASSIFY_SYSTEM,
            messages=[
                {"role": "user", "content": f"Post title: {post.title}\n\n{text}"}
            ],
            json_mode=True,
        )
        entries = _parse_entries_json(raw, post.id)
        if not entries:
            logger.info(
                "place_classify_empty place_id=%s post_id=%s", place_id, post.id
            )
            return
        await self._places.add_entries(place_id, entries)
        logger.info(
            "place_entries_ok place_id=%s post_id=%s count=%s",
            place_id,
            post.id,
            len(entries),
        )

    def _schedule_summary(
        self,
        place_id: UUID,
        post: PostDocument,
        *,
        immediate: bool = False,
    ) -> None:
        key = str(place_id)
        snippet = post.embed_text()[:1500]
        delay = 0.0 if immediate else self._debounce_sec

        async def _run() -> None:
            try:
                if delay > 0:
                    await asyncio.sleep(delay)
                await self._refresh_summary(place_id, post.title, snippet)
            except asyncio.CancelledError:
                raise
            except Exception:
                logger.exception("place_summary_failed place_id=%s", place_id)
            finally:
                async with self._lock:
                    cur = self._summary_tasks.get(key)
                    if cur is asyncio.current_task():
                        self._summary_tasks.pop(key, None)

        async def _replace() -> None:
            async with self._lock:
                prev = self._summary_tasks.get(key)
                if prev and not prev.done():
                    prev.cancel()
                self._summary_tasks[key] = asyncio.create_task(
                    _run(), name=f"place-summary-{key}"
                )

        asyncio.create_task(_replace())

    async def _refresh_summary(
        self, place_id: UUID, place_name: str, latest_snippet: str
    ) -> None:
        if not self._llm.configured:
            return
        user = (
            f"Place name: {place_name}\n\n"
            f"Latest post context:\n{latest_snippet}\n\n"
            f"Write an updated directory summary for this place."
        )
        summary = await self._llm.complete(
            system=SUMMARY_SYSTEM,
            messages=[{"role": "user", "content": user}],
        )
        summary = (summary or "").strip()
        if not summary:
            return
        await self._places.update_summary(
            place_id, summary=summary, model=self._llm.model_name
        )
        logger.info("place_summary_ok place_id=%s chars=%s", place_id, len(summary))


def _normalize_name(name: str) -> str:
    s = name.lower().strip()
    s = re.sub(r"[^\w\s]", " ", s, flags=re.UNICODE)
    return re.sub(r"\s+", " ", s).strip()


def _name_score(a: str, b: str) -> float:
    na, nb = _normalize_name(a), _normalize_name(b)
    if not na or not nb:
        return 0.0
    if na == nb:
        return 1.0
    return SequenceMatcher(None, na, nb).ratio()


def _cosine(a: list[float], b: list[float]) -> float:
    if not a or not b or len(a) != len(b):
        return 0.0
    dot = sum(x * y for x, y in zip(a, b))
    na = math.sqrt(sum(x * x for x in a))
    nb = math.sqrt(sum(y * y for y in b))
    if na == 0 or nb == 0:
        return 0.0
    return dot / (na * nb)


def _blend_embedding(old: list[float], new: list[float]) -> list[float]:
    if not old:
        return list(new)
    if len(old) != len(new):
        return list(new)
    return [0.7 * o + 0.3 * n for o, n in zip(old, new)]


def _parse_entries_json(raw: str, post_id: UUID) -> list[dict[str, Any]]:
    text = raw.strip()
    if text.startswith("```"):
        text = re.sub(r"^```(?:json)?\s*", "", text)
        text = re.sub(r"\s*```$", "", text)
    try:
        data = json.loads(text)
    except json.JSONDecodeError:
        logger.warning("place_classify_bad_json raw=%s", raw[:200])
        return []
    items = data.get("entries") if isinstance(data, dict) else None
    if not isinstance(items, list):
        return []
    out: list[dict[str, Any]] = []
    for item in items[:5]:
        if not isinstance(item, dict):
            continue
        cat = str(item.get("category") or "").strip().lower()
        body = str(item.get("body") or "").strip()
        if cat not in ("tip", "change", "important") or not body:
            continue
        out.append(
            {
                "category": cat,
                "body": body[:2000],
                "source": "ai",
                "source_post_id": str(post_id),
            }
        )
    return out
