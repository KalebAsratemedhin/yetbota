# YetBota — SRS

## Product

YetBota is a crowdsourced location discovery system.

Backend shape: three microservices — `content-service` (Go), `identity-service` (Go), `ai-service` (Python). Auth belongs to `identity-service`.

Other services validate **access** tokens by calling identity **introspect** (not local-only JWT verify).

Follow graph is stored in **SurrealDB** (shared). Identity writes follows; other services may **read** the same SurrealDB.

Posts and post media belong to **content-service** (shared Postgres). Media files are stored on **Cloudinary** via **client-side signed uploads**; content-service issues signatures and persists asset metadata.

## Tools and stack

| Layer | Choice | Notes |
| --- | --- | --- |
| Frontend | Next.js | |
| content-service | Go | Owns posts, media metadata, reactions, comments, bookmarks, share, post search, home feed, trending, **Q&A** |
| identity-service | Go | Owns auth + profile + follow writes + **reputation ledger, levels, badges** |
| ai-service | Python | FastAPI; future RAG ingest + post similarity → Surreal edges |
| Database | PostgreSQL | Shared; users/auth + posts/content tables (service-owned) |
| Graph DB | SurrealDB | Shared; follow relationships + future post embeddings/similarity. Local: Docker in `backend/docker-compose.yml` |
| Messaging | RabbitMQ | `post.created` content → ai (`yetbota.events` / `ai.post.ingest`) |
| Object / media | Cloudinary | Avatars (identity); post images/video/audio (content, signed upload) |
| Migrations (Go) | goose-format SQL | Applied on startup |
| Migrations (Python) | Alembic | ai-service; version table `alembic_version_ai` |
| Data access (Go) | sqlc + pgx | Postgres |
| Data access (Python) | SQLAlchemy 2 + asyncpg | ai-service |
| Password hashing | bcrypt | |
| Access token | JWT (HS256), short-lived | |
| Refresh token | Opaque, stored hashed | |
| Token validation (cross-service) | Introspect | |
| API docs | OpenAPI 3 + Swagger UI | |
| Logging (Go) | log/slog JSON | |
| Other | TBD | |

## Functional requirements

| ID | Requirement | Notes |
| --- | --- | --- |
| FR-001 | A user shall be able to sign up with a username, password, and full name. | Returns access + refresh tokens; `full_name` required |
| FR-002 | A user shall be able to sign in with username and password. | Returns access + refresh tokens |
| FR-003 | Sign-up shall fail if the username is already taken. | |
| FR-004 | Sign-in shall fail if the username or password is wrong. | |
| FR-005 | An authenticated user shall be able to fetch their profile (`/auth/me`). | |
| FR-006 | Other services shall validate an access token via identity introspect. | |
| FR-007 | An authenticated user shall be able to log out. | |
| FR-008 | An authenticated user shall be able to change their password. | |
| FR-009 | A client shall refresh tokens using a valid refresh token. | |
| FR-010 | An authenticated user shall be able to delete their account. | |
| FR-011 | An authenticated user shall be able to update profile fields: full name, phone, address, bio. | Phone: if non-empty, length 7–15 |
| FR-012 | An authenticated user shall be able to view another user’s public profile by id. | |
| FR-013 | An authenticated user shall be able to follow another user. | Stored in SurrealDB graph |
| FR-014 | An authenticated user shall be able to unfollow another user. | |
| FR-015 | An authenticated user shall be able to list a user’s followers. | Hydrated public profiles; keyset pagination |
| FR-016 | An authenticated user shall be able to list who a user is following. | Hydrated public profiles; keyset pagination |
| FR-017 | An authenticated user shall be able to search users by username or full name. | Min query length 2; max 20 results |
| FR-018 | An authenticated user shall be able to upload and delete a profile avatar. | Stored on Cloudinary; URL on user |
| FR-019 | An authenticated user shall obtain a Cloudinary signed upload for post media. | Issued by content-service |
| FR-020 | An authenticated user shall be able to create a post with a required place `title`, caption, required lat/lng, hashtags, and media. | `title` is the place name; must include visual media per allowed modes |
| FR-021 | Post media shall be one of: (1) 1–10 images; (2) 1–10 images + one audio; (3) exactly one video. | No text-only; no audio-only; no video combined with images or audio |
| FR-022 | Max images per post shall be 10; max video and audio duration shall default to 10 seconds and be configurable. | Config/env on content-service |
| FR-023 | An authenticated user shall be able to fetch a post by id. | Public post payload |
| FR-024 | An authenticated user shall be able to list another user’s (or own) posts. | Publicly visible posts; keyset pagination |
| FR-025 | An authenticated user shall be able to like, dislike, or clear their reaction on a post. | At most one of like/dislike per user per post |
| FR-026 | An authenticated user shall be able to add and list top-level comments on a post. | Keyset pagination; top-level only on list |
| FR-027 | An authenticated user shall be able to bookmark and unbookmark a post. | |
| FR-028 | An authenticated user shall be able to list their own bookmarked posts. | Caller only; not visible to other users |
| FR-029 | An authenticated user shall be able to share a post. | Returns share URL; increments share count |
| FR-030 | An authenticated user shall be able to search posts by text and hashtags. | Full-text over title + caption + hashtags |
| FR-031 | An authenticated user shall be able to list posts for a hashtag. | Keyset pagination |
| FR-031a | An authenticated user shall be able to list posts related to a seed post: remaining search matches for an optional query, then similar posts. | Not the home feed; keyset pagination |
| FR-032 | An authenticated user shall be able to reply to a top-level comment and list those replies. | One level only (no reply-to-reply); keyset pagination |
| FR-033 | An authenticated user shall be able to like, dislike, or clear their reaction on a comment. | At most one of like/dislike per user per comment |
| FR-034 | Each user shall have a reputation score derived from a ledger of reputation events. | Source of truth is the ledger; profile stores denormalized score |
| FR-035 | Reputation shall be derived from a log of post interaction events (likes, comments, shares, bookmarks, …) with per-event time freshness, synced from content-service. | Self-engagement (except post create) shall not write scoring events |
| FR-036 | A user’s reputation level shall be derived from their reputation score using configurable bands. | Level is display-only banding of score |
| FR-037 | An authenticated user shall be able to unlock badges when badge criteria are met. | Badges are awarded once; criteria may use score and/or activity |
| FR-038 | An authenticated user shall be able to fetch their reputation score, level, and earned badges. | Also exposed on public profile (score, level, badges) |
| FR-039 | An authenticated user shall be able to fetch a home feed that blends posts from followed users and nearby posts. | Keyset pagination; lat/lng optional (enables nearby + geo boost) |
| FR-040 | Feed ranking shall order candidate posts by a configurable heuristic score (engagement, freshness, follow boost, geo proximity). | No ML in this slice; formula documented below |
| FR-041 | An authenticated user shall be able to list trending posts. | HN-style score without follow/geo; keyset pagination; default 7-day window |
| FR-042 | An authenticated user shall be able to list trending creators. | Ranked by recent interaction events; top-N; hydrate profiles via identity |
| FR-043 | An authenticated user shall be able to ask a question with a body, optionally linked to a post and/or a place lat/lng. | Lat/lng optional; when set, both required — place location like posts, not GPS-only |
| FR-044 | An authenticated user shall be able to list and fetch questions (optionally filter by `post_id`). | Keyset pagination |
| FR-045 | An authenticated user shall be able to upvote, downvote, or clear their vote on a question or answer. | At most one vote per user per item; authors cannot vote on their own question/answer |
| FR-046 | The author of a question shall be able to mark it closed. | Closed questions reject new answers; votes still allowed |
| FR-047 | An authenticated user shall be able to answer an open question and list answers. | Keyset pagination |

## Feed ranking (non-ML)

Candidates for a feed page come from two pools: **following** (authors the viewer follows) and **nearby** (posts within a configurable radius of the viewer). Pools are merged, de-duplicated, then sorted by score.

Adopt a **Hacker News–style gravity** base (engagement decays with age), extended for YetBota:

```text
engagement = like_count + w_c * comment_count + w_s * share_count
rank = (engagement + 1)^0.8 / (age_hours + 2)^gravity
     * (1 + follow_boost if author is followed else 1)
     * geo_factor
```

| Symbol | Default | Meaning |
| --- | --- | --- |
| `gravity` | `1.5` | How fast older posts fall (higher → newer wins more) |
| `w_c`, `w_s`, `w_b` | `3`, `4`, `3` | Comments / shares / bookmarks count more than a like |
| `follow_boost` | `0.35` | Extra weight for in-network posts |
| `geo_factor` | `1 / (1 + distance_km / d0)` with `d0 = 5` | Closer places rank higher when lat/lng provided; optional |
| Nearby radius | `25 km` (configurable) | Max distance for the nearby pool |

Alternatives considered: Reddit “hot” (`log10(score) + age_term`) — good for forums; HN gravity fits better when freshness matters. Wilson score — better for comment sorting than for a mixed geo/social feed. ML ranking — deferred.

## Reputation defaults (engagement event log)

Content appends rows to `post_interaction_events` on create / like / unlike / dislike /
comment / share / bookmark / unbookmark (self-actions skipped except `create`).
Score is recomputed from that log and synced to identity:

\[
c_i = \frac{w(\mathrm{kind}_i)}{(t_i+4)^{1.5}},\quad
R(u)=\mathrm{round}\Big(40\log_{10}\big(1+(\max(0,\sum c_i))^{0.8}\big)\Big)
\]

| Kind | Weight \(w\) |
| --- | --- |
| create | 0.1 |
| like / unlike | +1 / −1 |
| dislike / undislike | −1 / +1 |
| comment | 3 |
| share | 3 |
| bookmark / unbookmark | +3 / −3 |

One create + like + comment + share from a viewer lands around **~10 pts**, not Contributor. Levels stay 100 / 400 / 1000.
Lookback 90 days (up to 5000 events). Feed still uses denormalized counters for ranking.

Ownership: event log on **content-service**; denormalized score + badges on **identity-service**.
See `docs/reputation-engagement-scoring.md`.

## Related posts (from search)

| Endpoint | Behavior |
| --- | --- |
| `GET /posts/{id}/related?q=` | Other posts matching `q` first (excluding seed); then FTS-similar posts that do **not** match `q`. Without `q`, similar only. |
| Response `source` | `search` \| `similar` \| `mixed` — this is **not** the home feed |

## Trending (non-ML)

| Endpoint | Behavior |
| --- | --- |
| `GET /posts/trending` | Recent posts (default 7 days) ranked with feed HN formula **without** follow/geo boosts; keyset via `cursor` |
| `GET /creators/trending` | Authors ranked by reputation-weighted interaction events in the same window; `{ user_id, score }[]` top-N (no cursor) |

Config under content-service `trending.*` (`lookback_days`, `candidate_limit`, `creator_event_limit`).

## Q&A (Stack Overflow–style)

| Endpoint | Behavior |
| --- | --- |
| `POST /questions` | Create with `body`; optional `post_id`; optional place `lat`/`lng` (both or neither) |
| `GET /questions` | List (optional `post_id` filter); keyset cursor |
| `GET /questions/{id}` | Detail |
| `POST /questions/{id}/vote` | `{ vote: "up"\|"down"\|null }` |
| `POST /questions/{id}/close` | Author only; blocks new answers |
| `POST/GET /questions/{id}/answers` | Create/list answers (create rejected if closed) |
| `POST /answers/{id}/vote` | Same vote shape |

Posts require a place-name **`title`** (NOT NULL). FTS includes title + caption + hashtags. Migration may wipe local post rows to apply NOT NULL.

## Use cases

### UC-001 — Sign up

- **Actor:** User
- **Goal:** Create an account
- **Main flow:** Submit username, password, and full name → store bcrypt hash + full name → return `access_token` + `refresh_token`.
- **Related FRs:** FR-001, FR-003

### UC-002 — Sign in

- **Actor:** User
- **Goal:** Authenticate
- **Main flow:** Verify credentials → return `access_token` + `refresh_token`.
- **Related FRs:** FR-002, FR-004

### UC-003 — Get own profile

- **Actor:** User
- **Goal:** Read own profile
- **Related FRs:** FR-005

### UC-004 — Introspect access token

- **Actor:** Internal service
- **Related FRs:** FR-006

### UC-005 — Log out

- **Related FRs:** FR-007

### UC-006 — Change password

- **Related FRs:** FR-008

### UC-007 — Refresh tokens

- **Related FRs:** FR-009

### UC-008 — Delete account

- **Related FRs:** FR-010

### UC-009 — Update profile

- **Main flow:** PATCH profile; if `phone` is non-empty it must be length 7–15.
- **Related FRs:** FR-011

### UC-010 — View other user profile

- **Actor:** User
- **Goal:** See another user’s public profile
- **Main flow:** `GET /users/{id}` with Bearer → public fields (no password).
- **Related FRs:** FR-012

### UC-011 — Follow / unfollow

- **Actor:** User
- **Goal:** Follow or unfollow another user
- **Main flow:** Authenticated `POST`/`DELETE` `/users/{id}/follow` → SurrealDB `user→follows→user` edge.
- **Related FRs:** FR-013, FR-014

### UC-012 — Followers / following lists

- **Actor:** User
- **Goal:** See followers or following for a user
- **Main flow:** `GET /users/{id}/followers|following?limit=&cursor=` → Surreal edge page → batch Postgres hydrate → `{ users, next_cursor?, total }`.
- **Related FRs:** FR-015, FR-016

### UC-013 — Search users

- **Actor:** User
- **Goal:** Find other users
- **Main flow:** `GET /users/search?q=` → public profiles matching username/full name.
- **Related FRs:** FR-017

### UC-014 — Avatar upload / delete

- **Actor:** User
- **Goal:** Set or clear profile photo
- **Main flow:** `POST /auth/me/avatar` (multipart `file`) uploads to Cloudinary and saves URL; `DELETE /auth/me/avatar` removes it.
- **Related FRs:** FR-018

### UC-015 — Sign post media upload

- **Actor:** User
- **Goal:** Get credentials to upload post media to Cloudinary
- **Main flow:** Authenticated request to content-service → signed upload params for images, video, or audio.
- **Related FRs:** FR-019

### UC-016 — Create post

- **Actor:** User
- **Goal:** Publish a location post with media
- **Main flow:** Upload media to Cloudinary with signature → `POST` post with caption, required `lat`/`lng`, hashtags, and media refs in an allowed mode (`images` | `images_audio` | `video`) → content-service validates and stores post + media metadata.
- **Related FRs:** FR-020, FR-021, FR-022

### UC-017 — View post

- **Actor:** User
- **Goal:** Read a single post
- **Main flow:** `GET /posts/{id}` → post, media, counts, hashtags.
- **Related FRs:** FR-023

### UC-018 — List a user’s posts

- **Actor:** User
- **Goal:** See posts by a given user
- **Main flow:** `GET /users/{id}/posts` → publicly visible posts, newest first, keyset pagination.
- **Related FRs:** FR-024

### UC-019 — React to a post

- **Actor:** User
- **Goal:** Like, dislike, or clear reaction
- **Main flow:** Set reaction for current user on post (replaces previous if any).
- **Related FRs:** FR-025

### UC-020 — Comment on a post

- **Actor:** User
- **Goal:** Add or read top-level comments
- **Main flow:** `POST` comment body (no `parent_id`); `GET /posts/{id}/comments` returns top-level comments with reply counts.
- **Related FRs:** FR-026

### UC-020a — Reply to a comment

- **Actor:** User
- **Goal:** Reply under a top-level comment
- **Main flow:** `POST` comment with `parent_id` of a top-level comment on the same post; `GET /comments/{id}/replies` lists direct replies.
- **Related FRs:** FR-032

### UC-020b — React to a comment

- **Actor:** User
- **Goal:** Like, dislike, or clear reaction on a comment
- **Main flow:** Set reaction for current user on comment (replaces previous if any).
- **Related FRs:** FR-033

### UC-021 — Bookmark posts

- **Actor:** User
- **Goal:** Save posts for later / manage bookmarks
- **Main flow:** Bookmark or unbookmark; `GET` own bookmarks only (not exposed as another user’s list).
- **Related FRs:** FR-027, FR-028

### UC-022 — Share a post

- **Actor:** User
- **Goal:** Share a post link
- **Main flow:** Share action → canonical share URL returned; share count incremented.
- **Related FRs:** FR-029

### UC-023 — Search posts / by hashtag

- **Actor:** User
- **Goal:** Find posts by text or tag
- **Main flow:** Search query over caption + hashtags; or list posts for one hashtag.
- **Related FRs:** FR-030, FR-031

### UC-024 — Earn reputation

- **Actor:** User (passive)
- **Goal:** Gain standing from posts and engagement
- **Main flow:** Someone engages with the user’s posts (like, comment, share, bookmark) or the user creates a post → content recomputes engagement-weighted score with freshness → identity syncs denormalized score, level, and badges.
- **Related FRs:** FR-034, FR-035, FR-036

### UC-025 — View reputation and badges

- **Actor:** User
- **Goal:** See score, level, and badges (own or another user’s public profile)
- **Main flow:** Fetch profile fields including reputation score, level, and earned badges.
- **Related FRs:** FR-037, FR-038

### UC-026 — Unlock a badge

- **Actor:** User (passive)
- **Goal:** Receive a badge when criteria are met
- **Main flow:** After a reputation or activity event, badge rules are evaluated; if newly satisfied, badge is awarded once.
- **Related FRs:** FR-037

### UC-027 — Home feed

- **Actor:** User
- **Goal:** Browse posts from people they follow and from nearby places
- **Main flow:** Client sends viewer location → content-service builds following + nearby candidate pools → ranks with the documented heuristic → returns a paginated feed.
- **Related FRs:** FR-039, FR-040

### UC-028 — Trending posts

- **Actor:** User
- **Goal:** Discover posts that are hot right now
- **Main flow:** Client calls `GET /posts/trending` → content-service loads recent posts in the trending window → ranks with HN gravity (no follow/geo) → returns a paginated page.
- **Related FRs:** FR-041

### UC-029 — Trending creators

- **Actor:** User
- **Goal:** Discover creators gaining engagement
- **Main flow:** Client calls `GET /creators/trending` → content-service aggregates recent interaction events by author → returns ranked `user_id` + score; client hydrates public profiles from identity.
- **Related FRs:** FR-042

### UC-030 — Ask a question

- **Actor:** User
- **Goal:** Ask the community about a place or topic
- **Main flow:** Submit body (optional post link and/or place lat/lng like a post) → question is stored open for answers.
- **Related FRs:** FR-043

### UC-031 — Answer and vote on Q&A

- **Actor:** User
- **Goal:** Contribute or rank answers
- **Main flow:** Answer an open question and/or upvote/downvote questions and answers (one vote per user per item).
- **Related FRs:** FR-045, FR-047

### UC-032 — Close a question

- **Actor:** Question author
- **Goal:** Stop new answers
- **Main flow:** Author closes the question → further answers are rejected; existing content and votes remain.
- **Related FRs:** FR-046
