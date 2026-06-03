#!/bin/sh

set -e

# Формируем DSN, если не передан явно
if [ -z "$POSTGRES_DSN" ]; then
    POSTGRES_USER="${POSTGRES_USER:-postgres}"
    POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
    POSTGRES_DB="${POSTGRES_DB:-queue}"
    POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
    POSTGRES_PORT="${POSTGRES_PORT:-5432}"
    POSTGRES_DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
fi

# Путь к миграциям (относительно WORKDIR)
MIGRATIONS_PATH="db/postgres/migration"

# Выполняем миграции
echo "Applying migrations..."
migrate -path "$MIGRATIONS_PATH" -database "$POSTGRES_DSN" -verbose up