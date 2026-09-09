from uuid import uuid4

from app.domain.post_doc import PostDocument
from app.services.ingest import IngestService


class FakeReader:
    def __init__(self, doc: PostDocument | None) -> None:
        self.doc = doc
        self.calls = 0

    async def get_post(self, post_id):
        self.calls += 1
        return self.doc


class FakeEmbedder:
    model_name = "fake"
    dimensions = 3

    async def embed(self, text: str) -> list[float]:
        return [0.1, 0.2, 0.3]


class FakeStore:
    def __init__(self) -> None:
        self.items = []

    async def upsert_embedding(self, post, text, embedding, model):
        self.items.append((post.id, text, list(embedding), model))


import pytest


@pytest.mark.asyncio
async def test_ingest_post_ok():
    pid = uuid4()
    doc = PostDocument(
        id=pid,
        author_id=uuid4(),
        title="Tomoca",
        caption="Great coffee",
        lat=9.0,
        lng=38.0,
        hashtags="cafe addis",
    )
    store = FakeStore()
    svc = IngestService(FakeReader(doc), FakeEmbedder(), store)
    assert await svc.ingest_post(pid) is True
    assert len(store.items) == 1
    assert store.items[0][0] == pid
    assert "Tomoca" in store.items[0][1]


@pytest.mark.asyncio
async def test_ingest_post_missing():
    store = FakeStore()
    svc = IngestService(FakeReader(None), FakeEmbedder(), store)
    assert await svc.ingest_post(uuid4()) is False
    assert store.items == []


def test_embed_text_joins_fields():
    doc = PostDocument(
        id=uuid4(),
        author_id=uuid4(),
        title=" A ",
        caption=" B ",
        lat=0,
        lng=0,
        hashtags=" c ",
    )
    assert doc.embed_text() == "A\nB\nc"
