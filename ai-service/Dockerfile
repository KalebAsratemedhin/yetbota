FROM python:3.12-slim AS deps
WORKDIR /build
COPY requirements.txt ./
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

FROM python:3.12-slim
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH=/app/src
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --create-home --uid 1000 appuser

WORKDIR /app
ENV PYTHONPATH=/app/src
COPY --from=deps /install /usr/local
COPY config.yaml ./config.yaml
COPY src ./src
EXPOSE 8989
USER appuser
CMD ["python", "-m", "main"]
