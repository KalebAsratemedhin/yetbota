# YetBota AI Service

Ingests posts on `post.created` (RabbitMQ): load from Postgres → embed → upsert
into SurrealDB table `post_embedding` (vector + metadata).

## Stack

| Piece | Choice |
| --- | --- |
| HTTP | FastAPI + Uvicorn |
| Settings | pydantic-settings |
| Postgres | SQLAlchemy 2 + asyncpg (read posts) |
| Embeddings | fastembed `all-MiniLM-L6-v2` (384-d, ONNX) |
| Vectors | SurrealDB `post_embedding` + HNSW cosine index |
| Messaging | RabbitMQ `yetbota.events` / `post.created` → `ai.post.ingest` |

## Flow

1. content-service creates a post and publishes `{ event, post_id, occurred_at }`
2. ai-service consumes the message
3. Loads `title`, `caption`, hashtags from Postgres
4. Embeds joined text
5. `UPSERT post_embedding:<post_id>` in Surreal with `embedding`, `text`, geo, model

## Local setup

```bash
cd backend
docker compose up -d postgres surrealdb rabbitmq

cd ai-service
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt   # first run downloads MiniLM ONNX weights
cp .env.example .env
uvicorn app.main:app --reload --port 8083
```

Health: `GET http://localhost:8083/healthz`
