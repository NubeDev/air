#!/usr/bin/env bash
set -e

echo "Checking AIR environment..."

# Check if docker is available
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ docker not found"
    exit 1
fi

# Check if docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ docker not running"
    exit 1
fi

echo "✅ Docker is available and running"

# Check if go is available
if ! command -v go >/dev/null 2>&1; then
    echo "❌ go not found"
    exit 1
fi

echo "✅ Go is available"

# Check if analytics databases are reachable
echo "Checking analytics databases..."

# Check TimescaleDB
if command -v pg_isready >/dev/null 2>&1; then
    if pg_isready -h localhost -p 5432 -d energy -U air >/dev/null 2>&1; then
        echo "✅ TimescaleDB is reachable"
    else
        echo "⚠️  TimescaleDB not reachable (run 'make db' to start)"
    fi
else
    echo "⚠️  pg_isready not found, skipping TimescaleDB check"
fi

# Check PostgreSQL
if command -v pg_isready >/dev/null 2>&1; then
    if pg_isready -h localhost -p 5433 -d sales -U reporter >/dev/null 2>&1; then
        echo "✅ PostgreSQL is reachable"
    else
        echo "⚠️  PostgreSQL not reachable (run 'make db' to start)"
    fi
else
    echo "⚠️  pg_isready not found, skipping PostgreSQL check"
fi

# Check MySQL
if command -v mysql >/dev/null 2>&1; then
    if mysql -h localhost -P 3306 -u air -pair -e "SELECT 1" >/dev/null 2>&1; then
        echo "✅ MySQL is reachable"
    else
        echo "⚠️  MySQL not reachable (run 'make db' to start)"
    fi
else
    echo "⚠️  mysql client not found, skipping MySQL check"
fi

# Check Ollama (optional)
if command -v curl >/dev/null 2>&1; then
    if curl -sSf http://localhost:11434/api/tags >/dev/null 2>&1; then
        echo "✅ Ollama is reachable"
    else
        echo "⚠️  Ollama not reachable (optional for local AI models)"
    fi
else
    echo "⚠️  curl not found, skipping Ollama check"
fi

echo ""
echo "Environment check completed!"
echo ""
echo "To start analytics databases: make db"
echo "To start AIR server: make run-dev"
echo "To build CLI: make cli"
