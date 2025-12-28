#!/bin/bash
# DevHive Database Migration Script
# This script runs all migrations in order against your Neon database

set -e

if [ -z "$1" ]; then
    echo "❌ Usage: ./migrate.sh <database-url>"
    echo "   Example: ./migrate.sh 'postgresql://user:pass@host/db?sslmode=require'"
    exit 1
fi

DATABASE_URL="$1"

echo "🚀 DevHive Database Migration Script"
echo "====================================="
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "❌ psql not found. Please install PostgreSQL client tools"
    echo "   Or use Neon SQL Editor at: https://console.neon.tech"
    exit 1
fi

# Get all migration files in order
cd cmd/devhive-api/migrations

for file in $(ls -1 [0-9][0-9][0-9]_*.sql | sort -V); do
    echo "📝 Running migration: $file"

    if psql "$DATABASE_URL" -f "$file" 2>&1 | grep -q "already exists\|duplicate key"; then
        echo "   ⚠️  Already applied (skipping)"
    elif psql "$DATABASE_URL" -f "$file" > /dev/null 2>&1; then
        echo "   ✅ Success"
    else
        echo "   ❌ Failed"
        exit 1
    fi
done

echo ""
echo "✅ All migrations completed successfully!"
echo ""
