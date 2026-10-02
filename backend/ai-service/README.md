# YetBota AI Service

Ingests content on RabbitMQ `*.created` / `*.updated` events: load from Postgres →
embed → upsert Surreal. Posts also maintain the Related-posts `similar` graph.
Exposes **RAG chat** over the typed `content_embedding` index.

## Stack

| Piece | Choice |
| --- | --- |
| HTTP | FastAPI + Uvicorn (`:8083`) |
| Postgres | Read posts, comments, questions, answers |
| Embeddings | Gemini `text-embedding-004` (768-d); optional local MiniLM via `EMBEDDING_PROVIDER=local` |
| Chatbot index | Surreal `content_embedding` (`kind`, `entity_id`, `parent_id`, …) |
| Related graph | Surreal `post_embedding` + `similar` (posts only) |
| Chat LLM | Pluggable (`CHAT_LLM_PROVIDER`): default **Gemini 2.5 Flash**; OpenAI kept as alternate |
| Messaging | `yetbota.events` → queue `ai.content.ingest` |

## Events

| Routing key | Payload IDs | Surreal writes |
| --- | --- | --- |
| `post.created` | `post_id` | `post_embedding` + `similar` **and** `content_embedding` (`kind=post`) |
| `post.updated` | `post_id` | Same as create (re-embed) |
| `comment.created` | `comment_id`, `post_id` | `content_embedding` (`kind=comment`, parent=post) |
| `question.created` | `question_id`, optional `post_id` | `content_embedding` (`kind=question`) |
| `question.updated` | `question_id`, optional `post_id` | Same as create (re-embed) |
| `answer.created` | `answer_id`, `question_id` | `content_embedding` (`kind=answer`, text=`Q:`+`A:`) |
| `answer.updated` | `answer_id`, `question_id` | Same as create (re-embed) |

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

Set `GEMINI_API_KEY` in `.env` before using chat (default provider). To use OpenAI
instead: `CHAT_LLM_PROVIDER=openai` and `OPENAI_API_KEY`.

## Chat (RAG)

```bash
curl -s http://localhost:8083/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Where should I eat injera?"}'
```

Flow: embed message → cosine search on `content_embedding` → LLM (Gemini by
default) with retrieved snippets → `{ reply, model, sources }`. Optional `history`
(up to 20 user/assistant turns) is forwarded to the LLM. Auth is deferred for this
slice.

OpenAPI: `api/openapi.yaml` · Swagger: `http://localhost:8083/docs`
