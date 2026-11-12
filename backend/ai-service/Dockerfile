# Production image for ai-service (Render / GHCR).
# Default embeddings are Gemini API — no local ONNX download required at runtime.
FROM python:3.13-slim-bookworm

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --create-home --uid 10001 --shell /usr/sbin/nologin appuser

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app ./app
COPY alembic.ini ./alembic.ini
COPY alembic ./alembic
COPY api ./api

USER appuser
EXPOSE 8080
ENV HTTP_HOST=0.0.0.0
ENV HTTP_PORT=8080
# Render sets PORT; fall back to 8080 locally / on GHCR smoke runs.
CMD ["sh", "-c", "uvicorn app.main:app --host 0.0.0.0 --port ${PORT:-8080}"]
