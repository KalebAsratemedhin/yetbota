PYTHON ?= python3
VENV := .venv
BIN := $(VENV)/bin
PYTHONPATH_FOR_RUN := src

.PHONY: install lint format typecheck run docker-build clean venv

venv:
	@test -d $(VENV) || $(PYTHON) -m venv $(VENV)
	@$(BIN)/pip install --upgrade pip wheel >/dev/null

install: venv
	$(BIN)/pip install -r requirements-dev.txt

lint:
	$(BIN)/ruff check src
	$(BIN)/ruff format --check src

format:
	$(BIN)/ruff format src
	$(BIN)/ruff check --fix src

typecheck:
	PYTHONPATH=$(PYTHONPATH_FOR_RUN) $(BIN)/mypy

run:
	PYTHONPATH=$(PYTHONPATH_FOR_RUN) $(BIN)/python -m main

docker-build:
	docker build -t yetbota/ai-service:dev .

clean:
	rm -rf .mypy_cache .ruff_cache dist build *.egg-info
	find . -type d -name __pycache__ -exec rm -rf {} +
