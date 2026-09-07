# YetBota AI Service

Python FastAPI service for post ingestion (RAG), embedding similarity, and writing
similar-post edges into SurrealDB. **Scaffold only** — no ingest/similarity APIs yet.

## Stack

| Piece | Choice |
| --- | --- |
| HTTP | FastAPI + Uvicorn |
| Settings | pydantic-settings |
| Postgres | SQLAlchemy 2 + asyncpg (same DB as Go services) |
| Migrations | Alembic (`alembic_version_ai` — does not clash with goose) |
| Graph | SurrealDB (shared; follows + future similarity edges) |

## Layout

```text
app/
  main.py                 # app factory
  config.py               # env / settings
  api/routes/             # HTTP routers (health only for now)
  domain/                 # entities + ports (future)
  services/               # use cases: ingest, similarity (future)
  adapters/
    postgres/             # engine, session, Base
    surreal/              # Surreal client stub
  models/                 # SQLAlchemy models (ai-owned tables only)
alembic/                  # migrations
```

## Local setup

```bash
cd backend/ai-service
python -m venv .venv
source .venv/bin/activate   # Windows: .venv\Scripts\activate
pip install -r requirements.txt
cp .env.example .env
```

Postgres + Surreal from `backend/docker-compose.yml` (Postgres on `localhost:5433`).

```bash
# Apply AI migrations (none yet beyond empty baseline when added)
alembic upgrade head

# Run API
uvicorn app.main:app --reload --port 8083
```

Health: `GET http://localhost:8083/healthz`

## Env

See `.env.example`. Defaults match content-service local Postgres/Surreal.
