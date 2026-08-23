# YetBota — SRS

## Product

YetBota is a crowdsourced location discovery system.

Backend shape: three microservices — `content-service` (Go), `identity-service` (Go), `ai-service` (Python). Auth belongs to `identity-service`.

## Tools and stack

| Layer | Choice | Notes |
| --- | --- | --- |
| Frontend | Next.js | |
| content-service | Go | |
| identity-service | Go | Owns auth (signup / sign-in) |
| ai-service | Python | |
| Database | PostgreSQL | One shared database for all services |
| Migrations (Go) | goose | Versioned SQL under each Go service’s `migrations/` |
| Data access (Go) | sqlc + pgx | SQL in `.sql` files → generated Go; pgx driver |
| Other | TBD | Messaging, etc. |

## Functional requirements

| ID | Requirement | Notes |
| --- | --- | --- |
| FR-001 | A user shall be able to sign up with a username and password. | identity-service |
| FR-002 | A user shall be able to sign in with username and password. | identity-service |
| FR-003 | Sign-up shall fail if the username is already taken. | |
| FR-004 | Sign-in shall fail if the username or password is wrong. | Do not reveal which field failed in the client-facing message if that is a later security choice; for now, failure is enough. |

## Use cases

### UC-001 — Sign up

- **Actor:** User
- **Goal:** Create an account with username and password
- **Main flow:**
  1. User submits a username and password.
  2. System checks the username is available.
  3. System stores the new account (password stored securely — mechanism TBD).
  4. User is signed up (session/token handling TBD).
- **Success:** Account exists; user can proceed to sign in (or is already authenticated — TBD).
- **Related FRs:** FR-001, FR-003

### UC-002 — Sign in

- **Actor:** User
- **Goal:** Access the system with an existing username and password
- **Main flow:**
  1. User submits username and password.
  2. System verifies the credentials.
  3. User is signed in (session/token handling TBD).
- **Success:** User is authenticated for subsequent requests.
- **Failures:** Wrong username or password → FR-004
- **Related FRs:** FR-002, FR-004
