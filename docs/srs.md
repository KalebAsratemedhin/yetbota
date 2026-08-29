# YetBota — SRS

## Product

YetBota is a crowdsourced location discovery system.

Backend shape: three microservices — `content-service` (Go), `identity-service` (Go), `ai-service` (Python). Auth belongs to `identity-service`.

Other services validate **access** tokens by calling identity **introspect** (not local-only JWT verify).

## Tools and stack

| Layer | Choice | Notes |
| --- | --- | --- |
| Frontend | Next.js | |
| content-service | Go | |
| identity-service | Go | Owns auth + profile |
| ai-service | Python | |
| Database | PostgreSQL | One shared database for all services |
| Migrations (Go) | goose-format SQL | Applied on startup |
| Data access (Go) | sqlc + pgx | |
| Password hashing | bcrypt | |
| Access token | JWT (HS256), short-lived | API calls; `jti` denylist |
| Refresh token | Opaque, stored hashed | Signup/signin/refresh; used only to get new tokens |
| Token validation (cross-service) | Introspect | On **access** tokens |
| API docs | OpenAPI 3 + Swagger UI | `/swagger/` |
| Logging (Go) | log/slog JSON | Access + error logs to stdout; `LOG_LEVEL` |
| Other | TBD | |

## Functional requirements

| ID | Requirement | Notes |
| --- | --- | --- |
| FR-001 | A user shall be able to sign up with a username, password, and full name. | Returns access + refresh tokens; `full_name` required |
| FR-002 | A user shall be able to sign in with username and password. | Returns access + refresh tokens |
| FR-003 | Sign-up shall fail if the username is already taken. | |
| FR-004 | Sign-in shall fail if the username or password is wrong. | |
| FR-005 | An authenticated user shall be able to fetch their profile (`/auth/me`). | Includes username + profile fields |
| FR-006 | Other services shall validate an access token via identity introspect. | |
| FR-007 | An authenticated user shall be able to log out. | Revokes access `jti` and the presented refresh token |
| FR-008 | An authenticated user shall be able to change their password. | |
| FR-009 | A client shall refresh tokens using a valid refresh token. | Rotates refresh; issues new access + refresh |
| FR-010 | An authenticated user shall be able to delete their account. | |
| FR-011 | An authenticated user shall be able to update profile fields: full name, phone, address, bio. | |

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

### UC-003 — Get profile

- **Actor:** User
- **Goal:** Read own profile
- **Main flow:** Bearer access token → return user id, username, full_name, phone, address, bio.
- **Related FRs:** FR-005

### UC-004 — Introspect access token

- **Actor:** Internal service
- **Goal:** Validate an access token
- **Main flow:** POST token → `active` + user identity when valid (not expired/revoked; user exists).
- **Related FRs:** FR-006

### UC-005 — Log out

- **Actor:** User
- **Goal:** End session
- **Main flow:** Bearer access + refresh token in body → revoke access `jti` and invalidate refresh token.
- **Related FRs:** FR-007

### UC-006 — Change password

- **Actor:** User
- **Goal:** Set a new password
- **Related FRs:** FR-008

### UC-007 — Refresh tokens

- **Actor:** User (client)
- **Goal:** Get new tokens without password
- **Main flow:** POST `{ "refresh_token" }` → validate stored hash → revoke old refresh → issue new access + refresh.
- **Related FRs:** FR-009

### UC-008 — Delete account

- **Actor:** User
- **Goal:** Remove account
- **Related FRs:** FR-010

### UC-009 — Update profile

- **Actor:** User
- **Goal:** Edit full name, phone, address, bio
- **Main flow:** Bearer access → PATCH profile fields (any subset) → return updated profile.
- **Related FRs:** FR-011
