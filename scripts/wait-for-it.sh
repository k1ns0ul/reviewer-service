#!/bin/bash
set -e

HOST="$1"
PORT="$2"
TIMEOUT="${3:-30}"

echo "Ожидание $HOST:$PORT..."

for i in $(seq 1 $TIMEOUT); do
    if nc -z "$HOST" "$PORT" 2>/dev/null; then
        echo "$HOST:$PORT доступен!"
        exit 0
    fi
    echo "Попытка $i/$TIMEOUT..."
    sleep 1
done

echo "Timeout: $HOST:$PORT недоступен после $TIMEOUT секунд"
exit 1