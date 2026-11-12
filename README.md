# YetBota

Crowdsourced **location discovery** — people pin real places with photos, short video, or audio, follow each other, ask questions, and browse a place directory that AI keeps organized over time.

Guests can explore feeds, profiles, places, and Q&A. Signing in unlocks compose, follow, reactions, bookmarks, and reputation.

---

## Product idea

YetBota treats every post as a **place-first** report: a title (place name), coordinates, caption/hashtags, and media. The product is not a generic social feed with optional location — location *is* the content model.

Around that core:

- **Social graph** — follow people whose places you care about; home feed blends in-network and nearby posts, then falls through to suggested ranking.
- **Engagement & reputation** — likes, comments, shares, and bookmarks feed both post ranking and a user reputation ledger with levels and badges.
- **Community Q&A** — Stack Overflow–style questions (optionally tied to a post or map pin) with answers and votes.
- **Place directory** — one Place per real-world location; posts attach via geo + name + embedding match; AI summaries and tip/change/important entries stay current as new reports land.
- **Assistant** — RAG chat grounded in ingested posts, comments, questions, and answers.

The product target is practical local knowledge (hotels, cafés, civic landmarks, “what changed”) more than viral short-form for its own sake.

---

## Repository layout

```
yetbota/
├── frontend/                 # Next.js App Router client
├── backend/
│   ├── identity-service/     # Auth, profiles, follows, reputation
│   ├── content-service/      # Posts, feed, places, Q&A, engagement
│   ├── ai-service/           # Embeddings, place attach, RAG chat
│   └── docker-compose.yml    # Postgres, SurrealDB, RabbitMQ, Adminer
└── practices/                # Design / process references
```

| Surface | Default local URL |
| --- | --- |
| Frontend | `http://localhost:3000` |
| Identity API + Swagger | `http://localhost:8081` · `/swagger/index.html` |
| Content API + Swagger | `http://localhost:8082` · `/swagger/index.html` |
| AI API | `http://localhost:8083` |
| Postgres | `localhost:5433` (`yetbota` / `postgres` / `postgres`) |
| SurrealDB | `ws://localhost:8000` |
| RabbitMQ UI | `http://localhost:15672` (`yetbota` / `yetbota`) |
| Adminer | `http://localhost:8080` |

---

## Architecture

### Services

```text
                    ┌─────────────────┐
                    │    Frontend     │
                    │    (Next.js)    │
                    └────────┬────────┘
           Bearer JWT        │
     ┌───────────────────────┼───────────────────────┐
     ▼                       ▼                       ▼
┌─────────────┐       ┌─────────────┐         ┌─────────────┐
│  identity   │◄──────│   content   │────────►│     ai      │
│   (Go)      │introspect (Go)      │ events  │  (Python)   │
└──────┬──────┘       └──────┬──────┘ RabbitMQ└──────┬──────┘
       │                     │                       │
       ▼                     ▼                       ▼
   Postgres              Postgres                 Postgres
   SurrealDB ◄────────── SurrealDB ◄──────────── SurrealDB
   (follow writes)       (read follows,           (embeddings,
                          similar edges)           similar edges)
       │                     │
       └──────── Cloudinary ─┘
            (avatars + post media via signed upload)
```

| Service | Owns | Notable deps |
| --- | --- | --- |
| **identity-service** | Signup/signin, JWT access + opaque refresh, profiles/avatars, follow graph writes, reputation score/levels/badges, public leaderboard | Postgres, SurrealDB (follows), Cloudinary (avatars) |
| **content-service** | Posts & media metadata, reactions/comments/bookmarks/share, search/hashtags/related, home + trending + top feeds, Q&A, place directory reads, interaction event log → reputation sync | Postgres, SurrealDB (read follows / similar), RabbitMQ publish, Cloudinary sign |
| **ai-service** | Content embeddings, post `similar` edges, place attach/classify/summary, RAG `/chat` | Postgres, SurrealDB, RabbitMQ consume, Gemini (default) |

Cross-service auth: content and AI validate access tokens by calling identity **introspect** (not local-only JWT verify).

### Data stores

| Store | Role |
| --- | --- |
| **PostgreSQL** | Shared cluster, service-owned tables — users/auth/reputation (identity); posts, media, engagement, Q&A, places (content); AI-side relational needs |
| **SurrealDB** | Shared graph — `user→follows→user`; post embeddings and `similar` edges; content embeddings for chat RAG |
| **RabbitMQ** | `*.created` / `*.updated` (post, comment, question, answer) → `ai.content.ingest` |
| **Cloudinary** | Binary media — client uploads with signatures from identity (avatars) or content (`/media/sign`) |

### Go service shape

Identity and content follow a go-kit layered layout: transport → endpoint → usecase → repository, with drivers for Postgres (sqlc/pgx), Surreal, JWT, Cloudinary, and config. Migrations are goose SQL applied on startup.

### Frontend

Next.js App Router, TypeScript, Tailwind, RTK Query talking **directly** to identity and content base URLs (CORS on the services). Session: access + refresh in Redux/`localStorage`, refresh-on-401. Compose flow: `/media/sign` → Cloudinary FormData upload → `POST /posts`.

---

## Features overview

### Accounts & social

- Signup / signin, refresh, logout, password change, account delete
- Public profiles (guests OK); optional-auth follower & following lists
- Follow / unfollow; user search (auth required)
- Avatar upload/delete; reputation score, level, badges; leaderboard

### Places & posts

- Create/edit place-titled posts with lat/lng, hashtags, and media modes: images · images+audio · video
- Signed Cloudinary uploads with size limits advertised on `/media/sign`
- Public post detail, user posts, hashtag and text search, related posts (search + similar graph)
- Reactions, comments (+ one-level replies), bookmarks, share

### Discovery

- Home feed (home → suggested phases) with optional geolocation
- Trending posts & creators; top posts by absolute engagement
- Place directory: list/search/nearby, place detail, attached posts, tip/change/important entries, top places by engagement

### Community & AI

- Questions (optional post / place pin), answers, votes, author close/edit
- AI ingest on content events; place attach + debounced summaries
- RAG assistant chat over ingested community content

---

## Deep dive: feed ranking

Home discovery is a **two-phase, non-ML** ranker on content-service (`GET /feed`). No personalized model — configurable heuristics only.

**Phase `home`.** Candidate pools are merged and de-duplicated:

1. Posts from people you follow (Surreal follow graph; empty graph → recent global), and  
2. Optional **nearby** posts within a configurable radius when the client sends lat/lng.

Candidates are limited by a lookback window (default ~14 days) and a pool-size cap, then scored with Hacker News–style gravity plus YetBota boosts:

```text
engagement = likes + w_c·comments + w_s·shares (+ bookmark weight where configured)
rank = (engagement + 1)^0.8 / (age_hours + 2)^gravity
     × (1 + follow_boost if author is followed)
     × geo_factor          # closer places rank higher when lat/lng present
```

**Phase `suggested`.** When the home pool is empty on first paint, or keyset pagination exhausts it, the feed continues without an age floor: trending-style HN gravity **without** follow/geo boosts, excluding posts already shown in home. The cursor encodes the phase; the response includes `phase` (`home` | `suggested`) so the UI can label the discovery continuation.

Related surfaces reuse pieces of the same idea:

| Endpoint | Difference |
| --- | --- |
| `GET /posts/trending` | Recent window (default 7d), HN gravity, no follow/geo |
| `GET /posts/top` | Absolute engagement (likes+comments+shares+bookmarks), not age-decayed |
| `GET /creators/trending` | Authors ranked from recent interaction events |

Guests get the same feed path with an empty follow graph (recent + suggested).

---

## Deep dive: reputation

Reputation is **earned from how others engage with your posts**, not from self-actions (except a small weight on post create).

1. **content-service** appends rows to `post_interaction_events` on create / like / unlike / dislike / comment / share / bookmark (and inverse kinds where needed). Self-engagement is skipped except create.  
2. Score is recomputed from a time-decayed weighted sum of those events (lookback ~90 days), then **synced to identity-service**, which owns the denormalized `reputation_score`, display **level** bands, and **badges**.  
3. Public profiles and `GET /users/leaderboard` expose the denormalized score; badges are one-shot score milestones (not activity counters).

Intuition: fresher, higher-signal actions (comment / share / bookmark weighted above a like) move the needle more than old likes. Feed ranking still uses denormalized post counters; reputation is a separate author-level ledger.

Levels are display bands over the synced score (e.g. Contributor / Local Guide / Pathfinder thresholds). Identity never re-derives the event log — content remains the source of interaction truth.

---

## Deep dive: AI post processing

AI work is **event-driven**. content-service publishes `*.created` / `*.updated` on RabbitMQ; ai-service consumes `ai.content.ingest`.

**Ingest pipeline (per event)**

1. Load the entity from Postgres (post, comment, question, or answer).  
2. Embed text (default Gemini embeddings; optional local MiniLM).  
3. Upsert into Surreal `content_embedding` (typed by `kind` / `entity_id` / parent) for **RAG chat**.  
4. For **posts only**, also upsert `post_embedding`, find cosine neighbors above a min score, and replace outbound `similar` edges — this powers **Related posts**, separate from place identity.

**Place directory (posts)**

After embed ingest, ai-service tries to **attach** the post to an existing Place or create one: hybrid geo + name overlap + embedding cutoff. Ambiguous matches do not merge. A post belongs to at most one Place in v1. Membership is *same-place identity*, not the soft `similar` graph.

On attach (and when membership/entries change), AI refreshes the Place **summary** asynchronously (debounced) and may write directory entries classified as `tip`, `change`, or `important`. content-service owns Place reads (`GET /places…`); attach/summary writes go through internal service APIs.

**Chat**

`POST /chat` embeds the user message, retrieves nearby `content_embedding` chunks, and asks the chat LLM (default Gemini) for a grounded reply plus sources. Optional short history is supported; auth on chat is deferred in the current slice.

---

## Quick start (local)

```bash
# Infra
cd backend && make db-up

# APIs (identity :8081, content :8082, ai :8083) — hot reload
make up
# or: make run   # if deps already up

# Frontend
cd ../frontend && npm install && npm run dev
```

Configure Cloudinary and JWT secrets via each service’s `.env` / config (see `.env.example` under each service). Seed demo posts against local or deployed APIs with `backend/scripts/seed-demo.sh`.

HTTP contracts live in `backend/*/api/openapi.yaml` and are served via each service’s Swagger UI.
