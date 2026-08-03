# ai-service

YetBota AI service. Owns content duplicate detection (UC-005), the RAG
assistant (UC-014), and profanity screening for community Q&A
(UC-009/UC-010). Built with FastAPI + gRPC, clean architecture.

See `docs/ai-service-backlog.txt` at the repo root for the full plan.

## Prerequisites

- Python 3.11 or newer
- `make`

That's it — no global tooling beyond stock Python.

## Quickstart

```sh
# 1. Create a venv and install runtime + dev deps.
make install

# 2. Generate Python proto stubs into ../common/proto/generated/python/.
make proto

# 3. Copy .env.example to .env and fill in secrets (Gemini API key,
#    Weaviate URL, RabbitMQ URI).
cp .env.example .env

# 4. Run tests.
make test

# 5. Run the service (HTTP on :8080, gRPC on :9090).
make run
```

`make install` creates `.venv/` in the project directory and installs
from `requirements-dev.txt` (which extends `requirements.txt`). The venv
is gitignored.

## Layout

```
src/
  domain/         entities + ports — pure, no I/O.
  application/    use cases — depend on domain only.
  infrastructure/ adapters (Gemini, Weaviate, RabbitMQ, config, logging).
  interfaces/     inbound: FastAPI routes, gRPC handlers, ingest worker.
  main.py         composition root.
```

The dependency rule (enforced by `make contracts` via import-linter):
`domain` imports nothing outside itself; `application` imports only
`domain`; `infrastructure` and `interfaces` may import either; `main.py`
is the only place that wires everything together.

## Proto stubs

Proto definitions live in `../common/proto/definition/ai/`. Generated
Python lands in `../common/proto/generated/python/` and is added to
`PYTHONPATH` by `make run` and `make test`. In Docker, `make proto` runs
during the build so the image is self-contained.
