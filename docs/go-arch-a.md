# Go Service Architecture

Each Go service follows a layered, go-kit-oriented layout. Dependencies point inward: transport and drivers depend on domain interfaces; usecases depend on those interfaces, not on concrete infra.

## Request flow

```
HTTP → transport (decode) → endpoint → usecase → repository / port
                              ↓
                         transport (encode)
```

Async work (if any) follows the same idea: message → processor → usecase → repository / port.

## Top-level layout

```
service/
├── cmd/                 # Process entrypoints and wiring
│   ├── main.go          # Load config, construct deps, start servers/workers
│   └── http/            # HTTP server bootstrap (listen, shutdown)
├── drivers/             # Concrete infrastructure (adapters)
├── internal/
│   ├── domain/          # Ports: interfaces, entities, domain types
│   ├── services/
│   │   ├── endpoint/    # go-kit endpoints (thin adapters over usecases)
│   │   ├── usecase/     # Business logic
│   │   └── repository/  # Persistence implementing domain ports
│   ├── transport/       # Protocol adapters (HTTP, etc.)
│   ├── messaging/       # Optional: queue names, message shapes
│   └── processors/      # Optional: async consumers / job handlers
└── migrations/          # SQL (or other) schema migrations
```

## Layer responsibilities

### `cmd/`

Composition root only. Wire config, drivers, repositories, usecases, endpoints, and transport. No business rules.

### `internal/domain/<feature>/`

Feature-owned **ports and models**: repository interfaces, session/hash/storage ports, entities, query options. No SQL, HTTP, or vendor SDKs here when avoidable.

### `internal/services/usecase/<feature>/`

Application / business logic. Talks only to domain interfaces. Typical files: `service.go` (interface + constructor), feature methods, `models.go`, helpers.

### `internal/services/repository/<feature>/`

Implements domain repository (or related) interfaces against Postgres, SurrealDB, Redis, etc. Maps between DB models and domain types.

### `internal/services/endpoint/`

go-kit `endpoint.Endpoint` wrappers: validate input, call usecase, map errors. One place to hang middleware (auth, logging, metrics) if needed at this layer.

### `internal/transport/http/<feature>/`

Protocol concern only:

| File        | Role                                      |
|-------------|-------------------------------------------|
| `handler.go`| Route → `kithttp.NewServer(...)`          |
| `decoder.go`| HTTP request → usecase/endpoint request   |
| `encoder.go`| Response / errors → HTTP                  |

Shared helpers (CORS, error encoding, server options) live under `transport/http/shared/`. A top-level `router.go` mounts feature handlers under path prefixes.

### `drivers/`

Replaceable tech details: config, logger, postgres/redis/clients, JWT, middleware, validators, generated DB models, storage SDKs, third-party HTTP clients. Drivers implement domain ports or are constructed in `cmd` and passed in.

### `migrations/`

Versioned schema changes, applied from `cmd` at startup (or via a dedicated migrate command).

## Optional async packages

Use when the service owns background work:

- **`internal/messaging/`** — queue/topic names and message DTOs
- **`internal/processors/`** — consumers that decode messages and call usecases (same business layer as HTTP)

## Conventions

1. **Feature folders** — Prefer `domain/foo`, `usecase/foo`, `repository/foo`, `transport/http/foo` over dumping everything in one package.
2. **Depend on interfaces** — Usecases take domain ports; `cmd` injects driver/repo implementations.
3. **Keep transport dumb** — No business logic in handlers/decoders beyond parsing and status codes.
4. **Keep endpoints thin** — Endpoints adapt and validate; rules live in usecases.
5. **Drivers stay outward** — Swap Redis/S3/JWT without changing usecases, as long as ports stay stable.

## Minimal skeleton for a new feature

```
internal/domain/<feature>/repository.go   # (+ entities if needed)
internal/services/usecase/<feature>/      # service.go, models.go, …
internal/services/repository/<feature>/   # repo.go
internal/services/endpoint/<feature>.go   # or package-level endpoint helpers
internal/transport/http/<feature>/        # handler, decoder, encoder
```

Wire the new pieces in `cmd/main.go` and register routes in `transport/http/router.go`.
