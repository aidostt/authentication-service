# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git wget

WORKDIR /app

# Resolve dependencies in a separate, cacheable layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary so it runs on a minimal base image without a C toolchain.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Database migration tool (pinned).
RUN wget -q https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz \
    && tar -xzf migrate.linux-amd64.tar.gz \
    && mv migrate.linux-amd64 /usr/local/bin/migrate \
    && rm migrate.linux-amd64.tar.gz

# ---- runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates postgresql-client \
    && adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/wait-for-postgres.sh ./wait-for-postgres.sh
RUN chmod +x ./wait-for-postgres.sh

USER app

EXPOSE 5050

# Waits for PostgreSQL, applies migrations, then starts the service.
ENTRYPOINT ["./wait-for-postgres.sh"]
