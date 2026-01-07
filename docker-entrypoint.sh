#!/bin/sh
set -e

echo "🚀 Starting Identity Service (Go)..."

if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL is not set"
  exit 1
fi

echo "📦 Running database migrations..."
/app/migrate -path /app/sql/schema -database "$DATABASE_URL" up

echo "✅ Migrations applied successfully"

echo "🔌 Starting application..."
exec /app/server
