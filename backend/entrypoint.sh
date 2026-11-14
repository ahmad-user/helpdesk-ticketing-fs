#!/bin/sh
set -e

# Run database migrations if AUTO_MIGRATE is enabled (default: true)
if [ "$AUTO_MIGRATE" != "false" ]; then
    echo "[INFO] Running database migrations up..."
    /app/migrate -direction=up -path=/app/migrations || {
        echo "[ERROR] Database migration failed"
        exit 1
    }
fi

# Run database seeding if AUTO_SEED is enabled (default: true)
if [ "$AUTO_SEED" != "false" ]; then
    echo "[INFO] Running database initial seeding..."
    /app/seed || {
        echo "[ERROR] Database seeding failed"
        exit 1
    }
fi

echo "[INFO] Starting Helpdesk API Server..."
exec "$@"
