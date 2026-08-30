# Multi-stage image for content-service (Render / GHCR).
FROM golang:1.25-alpine AS build
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/content-service ./cmd/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 10001 appuser

WORKDIR /app
COPY --from=build /out/content-service /app/content-service
COPY config.yaml /app/config.yaml
COPY migrations /app/migrations

USER appuser
EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV MIGRATE_ON_START=true

ENTRYPOINT ["/app/content-service", "-config", "/app/config.yaml"]
