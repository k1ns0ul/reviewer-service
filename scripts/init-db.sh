#!/bin/bash
set -e

echo "инициализация базы данных..."

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-reviewer_service}"

export PGPASSWORD=$DB_PASSWORD

echo "ожидание доступности PostgreSQL..."
until psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c '\q' 2>/dev/null; do
  echo "PostgreSQL недоступен - ожидание..."
  sleep 2
done

echo "PostgreSQL доступен!"

if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo "база данных '$DB_NAME' уже существует"
else
    echo "создание базы данных '$DB_NAME'..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"
    echo "база данных '$DB_NAME' создана!"
fi

echo "инициализация завершена!"