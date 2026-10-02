from uuid import uuid4

import pytest

from app.adapters.surreal.store import _cosine
from app.domain.content_doc import ContentDocument
from app.domain.post_doc import PostDocument
from app.domain.similarity import SimilarHit
from app.services.ingest import IngestService


class FakeReader:
    def __init__(
        self,
        post: PostDocument | None = None,
        *,
        comment: ContentDocument | None = None,
        question: ContentDocument | None = None,
        answer: ContentDocument | None = None,
    ) -> None:
        self.post = post
        self.comment = comment
        self.question = question
        self.answer = answer

    async def get_post(self, post_id):
        return self.post

    async def content_for_post(self, post_id):
        if self.post is None:
            return None
        return ContentDocument(
            kind="post",
            entity_id=self.post.id,
            author_id=self.post.author_id,
            text=self.post.embed_text(),
            post_id=self.post.id,
            title=self.post.title,
            lat=self.post.lat,
            lng=self.post.lng,
        )

    async def content_for_comment(self, comment_id):
        return self.comment

    async def content_for_question(self, question_id):
        return self.question

    async def content_for_answer(self, answer_id):
        return self.answer


class FakeEmbedder:
    model_name = "fake"
    dimensions = 3

    async def embed(self, text: str, *, task: str | None = None) -> list[float]:
        return [0.1, 0.2, 0.3]


class FakeStore:
    def __init__(self, hits: list[SimilarHit] | None = None) -> None:
        self.items = []
        self.content = []
        self.hits = hits or []
        self.edges: list[tuple] = []

    async def upsert_embedding(self, post, text, embedding, model):
        self.items.append((post.id, text, list(embedding), model))

    async def upsert_content(self, doc, embedding, model):
        self.content.append((doc.kind, doc.entity_id, doc.text, list(embedding), model))

    async def find_similar(self, embedding, exclude_post_id, *, top_k, min_score):
        return [h for h in self.hits if h.score >= min_score][:top_k]

    async def replace_similar_edges(self, from_post_id, hits):
        self.edges.append((from_post_id, list(hits)))


@pytest.mark.asyncio
async def test_ingest_post_ok_writes_edges_and_content():
    pid = uuid4()
    other = uuid4()
    doc = PostDocument(
        id=pid,
        author_id=uuid4(),
        title="Tomoca",
        caption="Great coffee",
        lat=9.0,
        lng=38.0,
        hashtags="cafe addis",
    )
    store = FakeStore(hits=[SimilarHit(post_id=other, score=0.9)])
    svc = IngestService(FakeReader(doc), FakeEmbedder(), store, similarity_min_score=0.5)
    assert await svc.ingest_post(pid) is True
    assert len(store.items) == 1
    assert len(store.content) == 1
    assert store.content[0][0] == "post"
    assert len(store.edges) == 1
    assert store.edges[0][0] == pid
    assert store.edges[0][1][0].post_id == other


@pytest.mark.asyncio
async def test_ingest_post_missing():
    store = FakeStore()
    svc = IngestService(FakeReader(None), FakeEmbedder(), store)
    assert await svc.ingest_post(uuid4()) is False
    assert store.items == []
    assert store.content == []
    assert store.edges == []


@pytest.mark.asyncio
async def test_ingest_comment_content_only():
    cid = uuid4()
    pid = uuid4()
    comment = ContentDocument(
        kind="comment",
        entity_id=cid,
        author_id=uuid4(),
        text="Post: Tomoca\nComment: Try the macchiato",
        parent_id=pid,
        post_id=pid,
    )
    store = FakeStore()
    svc = IngestService(FakeReader(comment=comment), FakeEmbedder(), store)
    assert await svc.ingest_comment(cid) is True
    assert store.items == []
    assert store.edges == []
    assert len(store.content) == 1
    assert store.content[0][0] == "comment"


@pytest.mark.asyncio
async def test_ingest_answer_prepends_via_reader_text():
    aid = uuid4()
    qid = uuid4()
    answer = ContentDocument(
        kind="answer",
        entity_id=aid,
        author_id=uuid4(),
        text="Q:\nIs it open?\nA:\nYes until 10",
        parent_id=qid,
    )
    store = FakeStore()
    svc = IngestService(FakeReader(answer=answer), FakeEmbedder(), store)
    assert await svc.handle_event("answer.created", {"answer_id": str(aid)}) is True
    assert store.content[0][2].startswith("Q:\n")


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


def test_cosine_identical():
    v = [1.0, 0.0, 0.0]
    assert abs(_cosine(v, v) - 1.0) < 1e-9


def test_cosine_orthogonal():
    assert abs(_cosine([1.0, 0.0], [0.0, 1.0])) < 1e-9
