# YetBota backend — always run from this directory: `cd backend && make <target>`
#
# Later: add content-service / ai-service under `run` so one command starts all APIs.

COMPOSE := docker compose
IDENTITY_DIR := identity-service
AIR := $(shell command -v air 2>/dev/null || echo "$(shell go env GOPATH)/bin/air")

.PHONY: help tools db-up db-down db-logs db-ps identity identity-once run up down swagger

help:
	@echo "YetBota backend"
	@echo ""
	@echo "  make tools         Install Air (hot reload)"
	@echo "  make up            Start Postgres, then run APIs with hot reload"
	@echo "  make run           Run all API services with hot reload (identity for now)"
	@echo "  make identity      Run identity-service with Air (restarts on code change)"
	@echo "  make identity-once Run identity-service once (no reload)"
	@echo "  make db-up         Start shared Postgres"
	@echo "  make db-down       Stop Postgres (keeps volume)"
	@echo "  make db-logs       Tail Postgres logs"
	@echo "  make db-ps         Show compose status"
	@echo "  make down          Stop Postgres"
	@echo "  make swagger       Print Swagger UI URL"
	@echo ""

tools:
	go install github.com/air-verse/air@latest
	@echo "Air installed to $$(go env GOPATH)/bin/air — ensure that dir is on PATH"

db-up:
	$(COMPOSE) up -d
	@echo "Postgres: localhost:5433  db=yetbota"

db-down down:
	$(COMPOSE) down

db-logs:
	$(COMPOSE) logs -f postgres

db-ps:
	$(COMPOSE) ps

identity: tools-check
	cd $(IDENTITY_DIR) && $(AIR)

identity-once:
	cd $(IDENTITY_DIR) && go run ./cmd/

# One entry point for all HTTP services (hot reload). Extend when more services exist.
run: identity

up: db-up
	$(MAKE) run

swagger:
	@echo "http://localhost:8081/swagger/index.html"

tools-check:
	@command -v $(AIR) >/dev/null 2>&1 || $(MAKE) tools
