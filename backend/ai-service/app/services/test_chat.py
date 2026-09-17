from uuid import uuid4

import pytest

from app.domain.chat import ContentHit
from app.services.chat import ChatService


class FakeEmbedder:
    model_name = "fake"
    dimensions = 3

    async def embed(self, text: str) -> list[float]:
        return [0.1, 0.2, 0.3]


class FakeSearcher:
    def __init__(self, hits: list[ContentHit] | None = None) -> None:
        self.hits = hits or []

    async def search_content(self, embedding, *, top_k, min_score):
        return [h for h in self.hits if h.score >= min_score][:top_k]


class FakeLLM:
    configured = True
    model_name = "fake-llm"

    async def complete(self, *, system: str, messages: list[dict[str, str]]) -> str:
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
