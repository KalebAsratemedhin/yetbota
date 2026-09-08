# YetBota AI Service

Python FastAPI service for post ingestion (RAG), embedding similarity, and writing
similar-post data into SurrealDB.

## Stack

| Piece | Choice |
| --- | --- |
| HTTP | FastAPI + Uvicorn |
| Settings | pydantic-settings |
| Postgres | SQLAlchemy 2 + asyncpg (same DB as Go services) |
| Migrations | Alembic (`alembic_version_ai` — does not clash with goose) |
| Graph | SurrealDB (embeddings + similarity later) |
| Messaging | RabbitMQ — consumes `post.created` from content-service |

## Layout

```text
app/
  main.py                 # app factory + RabbitMQ consumer lifespan
  config.py
  api/routes/             # health
  domain/ / services/     # future use cases
  adapters/
    postgres/
    surreal/
    rabbitmq/             # post.created consumer (log stub for now)
  models/
alembic/
```

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

Payload from content-service after `POST /posts`:

```json
{ "event": "post.created", "post_id": "<uuid>", "occurred_at": "<rfc3339>" }
```

Exchange `yetbota.events` (topic), routing key `post.created`, queue `ai.post.ingest`.
Consumer currently logs only — Surreal ingest TBD.

Health: `GET http://localhost:8083/healthz`  
RabbitMQ UI: `http://localhost:15672` (yetbota / yetbota)
