#!/bin/bash
set -e

echo "запуск миграций..."

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-reviewer_service}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

DSN="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

echo "применение миграций к базе данных: $DB_NAME"

migrate -path migrations -database "$DSN" up

echo "миграции успешно применены!"