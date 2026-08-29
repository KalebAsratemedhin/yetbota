# Multi-stage image for identity-service (Render / GHCR).
FROM golang:1.23-alpine AS build
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/identity-service ./cmd/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 10001 appuser

WORKDIR /app
COPY --from=build /out/identity-service /app/identity-service
COPY config.yaml /app/config.yaml
COPY migrations /app/migrations

USER appuser
EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV MIGRATE_ON_START=true

ENTRYPOINT ["/app/identity-service", "-config", "/app/config.yaml"]
