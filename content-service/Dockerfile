FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/content-service ./cmd/main.go

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/content-service .
COPY config.yaml .
COPY migrations ./migrations
EXPOSE 9966
CMD ["./content-service"]
