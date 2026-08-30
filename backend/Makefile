# YetBota backend — always run from this directory: `cd backend && make <target>`

COMPOSE := docker compose
IDENTITY_DIR := identity-service
CONTENT_DIR := content-service
AIR := $(shell command -v air 2>/dev/null || echo "$(shell go env GOPATH)/bin/air")

.PHONY: help tools db-up db-down db-logs db-ps identity identity-once content content-once run up down swagger

help:
	@echo "YetBota backend"
	@echo ""
	@echo "  make tools         Install Air (hot reload)"
	@echo "  make up            Start Postgres, then run APIs with hot reload"
	@echo "  make run           Run identity + content with Air"
	@echo "  make identity      Run identity-service with Air"
	@echo "  make identity-once Run identity-service once (no reload)"
	@echo "  make content       Run content-service with Air"
	@echo "  make content-once  Run content-service once (no reload)"
	@echo "  make db-up         Start Postgres + SurrealDB"
	@echo "  make db-down       Stop compose services (keeps volumes)"
	@echo "  make db-logs       Tail Postgres logs"
	@echo "  make db-ps         Show compose status"
	@echo "  make down          Stop compose services"
	@echo "  make swagger       Print Swagger UI URLs"
	@echo ""

tools:
	go install github.com/air-verse/air@latest
	@echo "Air installed to $$(go env GOPATH)/bin/air — ensure that dir is on PATH"

db-up:
	$(COMPOSE) up -d
	@echo "Postgres: localhost:5433  db=yetbota"
	@echo "SurrealDB: ws://localhost:8000  ns=yetbota db=main (root/root)"
	@echo "  Shared graph: identity writes follows; other services may read."

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

content: tools-check
	cd $(CONTENT_DIR) && $(AIR)

content-once:
	cd $(CONTENT_DIR) && go run ./cmd/

# One entry point for all HTTP services (hot reload).
run: tools-check
	$(MAKE) -j2 identity content

up: db-up
	$(MAKE) run

swagger:
	@echo "identity: http://localhost:8081/swagger/index.html"
	@echo "content:  http://localhost:8082/swagger/index.html"

tools-check:
	@command -v $(AIR) >/dev/null 2>&1 || $(MAKE) tools
