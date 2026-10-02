from uuid import uuid4

import pytest

from app.domain.chat import ContentHit
from app.services.chat import SYSTEM_PROMPT, ChatService, _format_context


class FakeEmbedder:
    model_name = "fake"
    dimensions = 3

    async def embed(self, text: str, *, task: str | None = None) -> list[float]:
        return [0.1, 0.2, 0.3]


class FakeSearcher:
    def __init__(self, hits: list[ContentHit] | None = None) -> None:
        self.hits = hits or []
        self.last_min_score: float | None = None

    async def search_content(self, embedding, *, top_k, min_score):
        self.last_min_score = min_score
        return [h for h in self.hits if h.score >= min_score][:top_k]


class FakeLLM:
    configured = True
    model_name = "fake-llm"
    last_system: str | None = None
    last_messages: list[dict[str, str]] | None = None

    async def complete(self, *, system: str, messages: list[dict[str, str]], json_mode: bool = False) -> str:
        self.last_system = system
        self.last_messages = messages
        assert system
        assert messages
        return "Try Mama's kitchen nearby."


@pytest.mark.asyncio
async def test_chat_returns_reply_and_sources():
    hit = ContentHit(
        kind="post",
        entity_id=uuid4(),
        score=0.9,
        text="Great injera",
        title="Mama's kitchen",
    )
    svc = ChatService(FakeEmbedder(), FakeSearcher([hit]), FakeLLM(), min_score=0.5)
    out = await svc.chat("Where should I eat?")
    assert out["reply"] == "Try Mama's kitchen nearby."
    assert out["model"] == "fake-llm"
    assert len(out["sources"]) == 1
    assert out["sources"][0]["kind"] == "post"


@pytest.mark.asyncio
async def test_chat_rejects_empty_message():
    svc = ChatService(FakeEmbedder(), FakeSearcher(), FakeLLM())
    with pytest.raises(ValueError):
        await svc.chat("   ")


@pytest.mark.asyncio
async def test_chat_requires_llm_config():
    class NoKey(FakeLLM):
        configured = False

    svc = ChatService(FakeEmbedder(), FakeSearcher(), NoKey())
    with pytest.raises(RuntimeError, match="not configured"):
        await svc.chat("hello")


@pytest.mark.asyncio
async def test_chat_soft_fallback_when_below_min_score():
    hit = ContentHit(
        kind="post",
        entity_id=uuid4(),
        score=0.15,
        text="Brisket bowl",
        title="Chanoly Carnivore",
    )
    llm = FakeLLM()
    searcher = FakeSearcher([hit])
    svc = ChatService(
        FakeEmbedder(),
        searcher,
        llm,
        min_score=0.22,
        soft_min_score=0.12,
    )
    out = await svc.chat("carnivore?")
    assert len(out["sources"]) == 1
    assert searcher.last_min_score == 0.12
    assert "Brisket bowl" in (llm.last_messages or [])[-1]["content"]


@pytest.mark.asyncio
async def test_chat_prefers_strict_hits_when_available():
    low = ContentHit(
        kind="post",
        entity_id=uuid4(),
        score=0.15,
        text="unrelated",
        title="Noise",
    )
    high = ContentHit(
        kind="post",
        entity_id=uuid4(),
        score=0.4,
        text="Great injera",
        title="Mama's kitchen",
    )
    llm = FakeLLM()
    svc = ChatService(
        FakeEmbedder(),
        FakeSearcher([low, high]),
        llm,
        min_score=0.22,
        soft_min_score=0.12,
    )
    out = await svc.chat("food?")
    assert len(out["sources"]) == 1
    assert out["sources"][0]["title"] == "Mama's kitchen"


def test_format_context_omits_ids():
    eid = uuid4()
    text = _format_context(
        [
            ContentHit(
                kind="post",
                entity_id=eid,
                score=0.9,
                text="Hello",
                title="Place",
            )
        ]
    )
    assert str(eid) not in text
    assert "id=" not in text
    assert "kind=post" in text
    assert "Place" in text


def test_system_prompt_forbids_ids():
    assert "UUID" in SYSTEM_PROMPT or "uuid" in SYSTEM_PROMPT.lower()
    assert "Never mention internal IDs" in SYSTEM_PROMPT
