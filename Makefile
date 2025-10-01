.PHONY: check dev-backend logs-backend dev-ui logs-ui dev-data logs-data dev-all dev-backend-ui logs-all logs-backend-ui db down db-test down-test seed-test-db openapi-gen cli build clean clean-ports test deps deps-ui deps-python deps-all help

# Default target
all: check build

# Check environment and dependencies
check:
	@echo "Checking environment..."
	@command -v docker >/dev/null || (echo "docker not found" && exit 1)
	@docker info >/dev/null || (echo "docker not running" && exit 1)
	@command -v go >/dev/null || (echo "go not found" && exit 1)
	@echo "Environment check passed"

# Development targets
dev-backend:
	@echo "Killing existing processes on port 9000..."
	@-pkill -f "go run ./cmd/api" 2>/dev/null || true
	@-pkill -f "bin/air" 2>/dev/null || true
	@-fuser -k 9000/tcp 2>/dev/null || true
	@sleep 1
	@echo "Starting AIR backend server..."
	go run ./cmd/api --data data --config config-dev.yaml --auth disabled &
	@echo "Go backend server started in background. Use 'make logs-backend' to view logs."

logs-backend:
	@echo "Tailing Go backend server logs..."
	@ps aux | grep "go run ./cmd/api" | grep -v grep | head -1 | awk '{print $$2}' | xargs -I {} tail -f /proc/{}/fd/1 2>/dev/null || echo "Go backend server not running. Start with 'make dev-backend'"

dev-ui:
	@echo "Killing existing processes on port 9002..."
	@-pkill -f "vite" 2>/dev/null || true
	@-fuser -k 9002/tcp 2>/dev/null || true
	@-fuser -k 3001/tcp 2>/dev/null || true
	@sleep 1
	@echo "Starting AIR UI on port 9002..."
	cd air-ui && npm run dev -- --port 9002 &
	@echo "UI server started in background. Use 'make logs-ui' to view logs."

dev-data:
	@echo "Killing existing processes on port 9001..."
	@-pkill -f "python -m app.main" 2>/dev/null || true
	@-fuser -k 9001/tcp 2>/dev/null || true
	@sleep 1
	@echo "Starting AIR-Py FastAPI data processing server..."
	cd dataserver && bash -c "source venv/bin/activate && python -m app.main" &
	@echo "FastAPI server started in background. Use 'make logs-data' to view logs."

logs-data:
	@echo "Tailing FastAPI server logs..."
	@ps aux | grep "python -m app.main" | grep -v grep | head -1 | awk '{print $$2}' | xargs -I {} tail -f /proc/{}/fd/1 2>/dev/null || echo "FastAPI server not running. Start with 'make dev-data'"

logs-ui:
	@echo "Tailing UI server logs..."
	@ps aux | grep "vite" | grep -v grep | head -1 | awk '{print $$2}' | xargs -I {} tail -f /proc/{}/fd/1 2>/dev/null || echo "UI server not running. Start with 'make dev-ui'"

# Combined development targets
dev-all:
	@echo "Starting Go backend, FastAPI, and UI servers..."
	@make clean-ports
	@make dev-backend
	@make dev-data
	@make dev-ui
	@echo "All servers started. Use 'make logs-all' to view combined logs."

dev-backend-ui:
	@echo "Starting Go backend and UI servers..."
	@make clean-ports
	@make dev-backend
	@make dev-ui
	@echo "Backend and UI started. Use 'make logs-backend-ui' to view logs."

# Force restart everything
restart:
	@echo "Force restarting all services..."
	@make clean-ports
	@sleep 2
	@make dev-all

logs-all:
	@echo "Tailing all server logs..."
	@echo "=== Go Backend Logs ==="
	@make logs-backend &
	@echo "=== FastAPI Logs ==="
	@make logs-data &
	@echo "=== UI Logs ==="
	@make logs-ui

logs-backend-ui:
	@echo "Tailing backend and UI logs..."
	@echo "=== Go Backend Logs ==="
	@make logs-backend &
	@echo "=== UI Logs ==="
	@make logs-ui

# Database targets
db:
	@echo "Starting analytics databases..."
	docker compose -f deploy/docker-compose.yml up -d

down:
	@echo "Stopping analytics databases..."
	docker compose -f deploy/docker-compose.yml down

# Test database targets
db-test:
	@echo "Starting TimescaleDB test database with Rubix OS schema..."
	docker compose -f deploy/docker-compose.yml up -d timescale-test
	@echo "TimescaleDB test database started on port 5434"
	@echo "Connection details:"
	@echo "  Host: localhost"
	@echo "  Port: 5434"
	@echo "  Database: rubix_test"
	@echo "  Username: rubix"
	@echo "  Password: rubix"
	@echo "  Connection string: postgres://rubix:rubix@localhost:5434/rubix_test"

down-test:
	@echo "Stopping TimescaleDB test database..."
	docker compose -f deploy/docker-compose.yml stop timescale-test
	@echo "To remove the test database completely (including data), run:"
	@echo "  docker compose -f deploy/docker-compose.yml down timescale-test -v"

# Seed test database with fake data
seed-test-db:
	@echo "Seeding test database with fake data..."
	cd deploy && ./seed-database.sh

# Code generation
openapi-gen:
	@echo "Generating OpenAPI code..."
	@command -v ~/go/bin/oapi-codegen >/dev/null || (echo "oapi-codegen not found, install with: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest" && exit 1)
	~/go/bin/oapi-codegen -generate types,server -package restapi -o internal/transport/rest/openapi.gen.go api/openapi.yaml
	~/go/bin/oapi-codegen -generate client -package apiclient -o clients/go/client.gen.go api/openapi.yaml

# Build targets
cli:
	@echo "Building CLI..."
	go build -o bin/aircli ./cmd/cli

build: cli
	@echo "Building API server..."
	go build -o bin/air ./cmd/api

# Clean up
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f air.db

clean-ports:
	@echo "Killing all AIR processes..."
	@-pkill -f "go run ./cmd/api" 2>/dev/null || true
	@-pkill -f "bin/air" 2>/dev/null || true
	@-pkill -f "python -m app.main" 2>/dev/null || true
	@-pkill -f "vite" 2>/dev/null || true
	@-fuser -k 9000/tcp 2>/dev/null || true
	@-fuser -k 9001/tcp 2>/dev/null || true
	@-fuser -k 9002/tcp 2>/dev/null || true
	@-fuser -k 3001/tcp 2>/dev/null || true
	@echo "All AIR processes stopped"

# Test targets
test:
	@echo "Running tests..."
	go test ./...

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download

deps-ui:
	@echo "Installing UI dependencies..."
	@command -v npm >/dev/null || (echo "npm not found, please install Node.js" && exit 1)
	cd air-ui && npm install
	@echo "UI dependencies installed successfully"

deps-python:
	@echo "Installing Python dependencies..."
	@command -v python3 >/dev/null || (echo "python3 not found" && exit 1)
	cd dataserver && python3 -m venv venv
	cd dataserver && bash -c "source venv/bin/activate && pip install -r requirements.txt"
	@echo "Python dependencies installed successfully"

deps-all: deps deps-ui deps-python
	@echo "All dependencies installed successfully"

# Run with auth disabled (development)
run-dev: build
	@echo "Killing existing processes on port 9000..."
	@-pkill -f "go run ./cmd/api" 2>/dev/null || true
	@-pkill -f "bin/air" 2>/dev/null || true
	@-fuser -k 9000/tcp 2>/dev/null || true
	@sleep 1
	@echo "Running AIR with authentication disabled..."
	./bin/air --data data --config config-dev.yaml --auth disabled

# Run with default config
run: build
	@echo "Killing existing processes on port 9000..."
	@-pkill -f "go run ./cmd/api" 2>/dev/null || true
	@-pkill -f "bin/air" 2>/dev/null || true
	@-fuser -k 9000/tcp 2>/dev/null || true
	@sleep 1
	@echo "Running AIR..."
	./bin/air --data data --config config.yaml

# Help
help:
	@echo "Available targets:"
	@echo "  check          - Check environment and dependencies"
	@echo "  dev-backend    - Start Go backend server in development mode"
	@echo "  logs-backend   - Tail Go backend server logs"
	@echo "  dev-ui         - Start React UI development server"
	@echo "  logs-ui        - Tail UI server logs"
	@echo "  dev-data       - Start FastAPI data processing server"
	@echo "  logs-data      - Tail FastAPI server logs"
	@echo "  dev-all        - Start all servers (backend, data, UI)"
	@echo "  logs-all       - Tail all server logs"
	@echo "  dev-backend-ui - Start backend and UI only"
	@echo "  logs-backend-ui- Tail backend and UI logs"
	@echo "  restart        - Force restart all services"
	@echo "  clean-ports    - Kill all AIR processes"
	@echo "  build          - Build CLI and API server"
	@echo "  run-dev        - Run with auth disabled"
	@echo "  db             - Start analytics databases"
	@echo "  down           - Stop analytics databases"
	@echo "  db-test        - Start TimescaleDB test database with Rubix OS schema"
	@echo "  down-test      - Stop TimescaleDB test database"
	@echo "  seed-test-db   - Populate test database with fake data (buildings, devices, history)"
	@echo "  openapi-gen    - Generate OpenAPI client/server code"
	@echo "  deps           - Install Go dependencies"
	@echo "  deps-ui        - Install UI dependencies (Node.js)"
	@echo "  deps-python    - Install Python dependencies"
	@echo "  deps-all       - Install all dependencies"
	@echo ""
	@echo "Quick Start:"
	@echo "  make deps-all       # Install all dependencies first"
	@echo "  make dev-all        # Start all servers (backend, data, UI)"
	@echo "  make dev-backend-ui # Start backend and UI only"
	@echo "  make logs-all       # View all logs"
	@echo "  make clean-ports    # Stop everything"
	@echo ""
	@echo "Health Checks:"
	@echo "  curl http://localhost:9000/v1/reports  # Go backend"
	@echo "  curl http://localhost:9002             # UI"
	@echo "  curl http://localhost:9001/v1/py/health # FastAPI"
	@echo ""
	@echo "For detailed guide, see SPEC-RUNNING.md"
