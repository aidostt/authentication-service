#!/bin/sh

# Waits for PostgreSQL, applies migrations, then starts the service.

set -e

host="${POSTGRES_HOST}"
port="${POSTGRES_PORT}"

until pg_isready -h "$host" -p "$port"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - running migrations"
# This service shares the dev database with reservation-service, so it tracks
# its migration versions in its own table to avoid clashing with theirs.
migrate -path /app/migrations -database "postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable&x-migrations-table=auth_schema_migrations" up

>&2 echo "Migrations completed - starting application"
exec ./api
