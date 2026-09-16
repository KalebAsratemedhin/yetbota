# YetBota AI Service

Ingests content on RabbitMQ `*.created` events: load from Postgres → embed →
upsert Surreal. Posts also maintain the Related-posts `similar` graph.

## Stack

| Piece | Choice |
| --- | --- |
| HTTP | FastAPI + Uvicorn |
| Postgres | Read posts, comments, questions, answers |
| Embeddings | fastembed MiniLM (384-d) |
| Chatbot index | Surreal `content_embedding` (`kind`, `entity_id`, `parent_id`, …) |
| Related graph | Surreal `post_embedding` + `similar` (posts only) |
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
