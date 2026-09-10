# YetBota AI Service

Ingests posts on `post.created` (RabbitMQ): load from Postgres → embed → upsert
Surreal `post_embedding`, then write outbound `similar` relation edges to nearest neighbors.

## Stack

| Piece | Choice |
| --- | --- |
| HTTP | FastAPI + Uvicorn |
| Postgres | Read posts (title, caption, hashtags) |
| Embeddings | fastembed MiniLM (384-d) |
| Vectors | Surreal `post_embedding` (+ HNSW index) |
| Graph | Surreal `similar` relation (`post_embedding` → `post_embedding`, `score`) |
| Messaging | RabbitMQ `post.created` → `ai.post.ingest` |

## Flow

1. content-service creates a post → publishes `{ event, post_id, occurred_at }`
2. ai-service embeds title + caption + hashtags
3. Upserts `post_embedding:<post_id>`
4. Cosine-ranks other embeddings (`SIMILARITY_TOP_K`, `SIMILARITY_MIN_SCORE`)
5. Replaces `post_embedding:A->similar->…` edges with scores

## Local setup

```bash
cd backend
docker compose up -d postgres surrealdb rabbitmq

cd ai-service
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
uvicorn app.main:app --reload --port 8083
```

Health: `GET http://localhost:8083/healthz`
