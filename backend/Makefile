# YetBota backend — always run from this directory: `cd backend && make <target>`

COMPOSE := docker compose
IDENTITY_DIR := identity-service
CONTENT_DIR := content-service
AI_DIR := ai-service
AIR := $(shell command -v air 2>/dev/null || echo "$(shell go env GOPATH)/bin/air")
AI_UVICORN := $(AI_DIR)/.venv/bin/uvicorn

.PHONY: help tools db-up db-down db-logs db-ps identity identity-once content content-once ai ai-once run up down swagger

help:
	@echo "YetBota backend"
	@echo ""
	@echo "  make tools         Install Air (hot reload)"
	@echo "  make up            Start Docker deps, then run all APIs (identity + content + ai)"
	@echo "  make run           Run identity + content + ai (hot reload)"
	@echo "  make identity      Run identity-service with Air"
	@echo "  make identity-once Run identity-service once (no reload)"
	@echo "  make content       Run content-service with Air"
	@echo "  make content-once  Run content-service once (no reload)"
	@echo "  make ai            Run ai-service with uvicorn --reload"
	@echo "  make ai-once       Run ai-service once (no reload)"
	@echo "  make db-up         Start Postgres + SurrealDB + RabbitMQ + Adminer"
	@echo "  make db-down       Stop compose services (keeps volumes)"
	@echo "  make db-logs       Tail Postgres logs"
	@echo "  make db-ps         Show compose status"
	@echo "  make down          Stop compose services"
	@echo "  make swagger       Print Swagger / health URLs"
	@echo ""

tools:
	go install github.com/air-verse/air@latest
	@echo "Air installed to $$(go env GOPATH)/bin/air — ensure that dir is on PATH"

db-up:
	$(COMPOSE) up -d
	@echo "Postgres: localhost:5433  db=yetbota"
	@echo "SurrealDB: ws://localhost:8000  ns=yetbota db=main (root/root)"
	@echo "RabbitMQ: amqp://yetbota:yetbota@localhost:5672/  UI http://localhost:15672"
	@echo "Adminer: http://localhost:8080"

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

ai: ai-check
	cd $(AI_DIR) && .venv/bin/uvicorn app.main:app --reload --host 0.0.0.0 --port 8083

ai-once: ai-check
	cd $(AI_DIR) && .venv/bin/uvicorn app.main:app --host 0.0.0.0 --port 8083

# One entry point for all HTTP services (hot reload).
run: tools-check ai-check
	$(MAKE) -j3 identity content ai

up: db-up
	$(MAKE) run

swagger:
	@echo "identity: http://localhost:8081/swagger/index.html"
	@echo "content:  http://localhost:8082/swagger/index.html"
	@echo "ai:       http://localhost:8083/healthz"

tools-check:
	@command -v $(AIR) >/dev/null 2>&1 || $(MAKE) tools

ai-check:
	@test -x $(AI_UVICORN) || { \
		echo "ai-service venv missing or incomplete."; \
		echo "  cd $(AI_DIR) && python3 -m venv .venv && . .venv/bin/activate && pip install -r requirements.txt && cp -n .env.example .env"; \
		exit 1; \
	}
