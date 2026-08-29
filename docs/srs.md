# YetBota — SRS

## Product

YetBota is a crowdsourced location discovery system.

Backend shape: three microservices — `content-service` (Go), `identity-service` (Go), `ai-service` (Python). Auth belongs to `identity-service`.

Other services validate **access** tokens by calling identity **introspect** (not local-only JWT verify).

Follow graph is stored in **SurrealDB** (shared). Identity writes follows; other services may **read** the same SurrealDB.

## Tools and stack

| Layer | Choice | Notes |
| --- | --- | --- |
| Frontend | Next.js | |
| content-service | Go | |
| identity-service | Go | Owns auth + profile + follow writes |
| ai-service | Python | |
| Database | PostgreSQL | Shared; users/auth source of truth |
| Graph DB | SurrealDB | Shared; follow relationships. Local: Docker in `backend/docker-compose.yml` |
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
