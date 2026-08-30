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
| content-service | Go | Owns posts, media metadata, reactions, comments, bookmarks, share, post search |
| identity-service | Go | Owns auth + profile + follow writes |
| ai-service | Python | |
| Database | PostgreSQL | Shared; users/auth + posts/content tables (service-owned) |
| Graph DB | SurrealDB | Shared; follow relationships. Local: Docker in `backend/docker-compose.yml` |
| Object / media | Cloudinary | Avatars (identity); post images/video/audio (content, signed upload) |
| Migrations (Go) | goose-format SQL | Applied on startup |
| Data access (Go) | sqlc + pgx | Postgres |
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
| FR-020 | An authenticated user shall be able to create a post with caption, required lat/lng, hashtags, and media. | Must include visual media: image(s) and/or video per allowed modes |
| FR-021 | Post media shall be one of: (1) 1–10 images; (2) 1–10 images + one audio; (3) exactly one video. | No text-only; no audio-only; no video combined with images or audio |
| FR-022 | Max images per post shall be 10; max video and audio duration shall default to 10 seconds and be configurable. | Config/env on content-service |
| FR-023 | An authenticated user shall be able to fetch a post by id. | Public post payload |
| FR-024 | An authenticated user shall be able to list another user’s (or own) posts. | Publicly visible posts; keyset pagination |
| FR-025 | An authenticated user shall be able to like, dislike, or clear their reaction on a post. | At most one of like/dislike per user per post |
| FR-026 | An authenticated user shall be able to add and list top-level comments on a post. | Keyset pagination; top-level only on list |
| FR-027 | An authenticated user shall be able to bookmark and unbookmark a post. | |
| FR-028 | An authenticated user shall be able to list their own bookmarked posts. | Caller only; not visible to other users |
| FR-029 | An authenticated user shall be able to share a post. | Returns share URL; increments share count |
| FR-030 | An authenticated user shall be able to search posts by text and hashtags. | Full-text over caption + hashtags |
| FR-031 | An authenticated user shall be able to list posts for a hashtag. | Keyset pagination |
| FR-032 | An authenticated user shall be able to reply to a top-level comment and list those replies. | One level only (no reply-to-reply); keyset pagination |
| FR-033 | An authenticated user shall be able to like, dislike, or clear their reaction on a comment. | At most one of like/dislike per user per comment |

Feed / home timeline is **out of scope** for this SRS slice.

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
