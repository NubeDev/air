.PHONY: check dev-backend logs-backend dev-ui dev-data logs-data dev-all logs-all db down openapi-gen cli build clean clean-ports test

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
	@echo "Killing existing process on port 9000..."
	@-fuser -k 9000/tcp 2>/dev/null || true
	@echo "Starting AIR backend server..."
	go run ./cmd/api --data data --config config-dev.yaml --auth &
	@echo "Go backend server started in background. Use 'make logs-backend' to view logs."

logs-backend:
	@echo "Tailing Go backend server logs..."
	@ps aux | grep "go run ./cmd/api" | grep -v grep | head -1 | awk '{print $$2}' | xargs -I {} tail -f /proc/{}/fd/1 2>/dev/null || echo "Go backend server not running. Start with 'make dev-backend'"

dev-ui:
	@echo "UI not implemented yet"

dev-data:
	@echo "Killing existing process on port 9001..."
	@-fuser -k 9001/tcp 2>/dev/null || true
	@echo "Starting AIR-Py FastAPI data processing server..."
	cd dataserver && bash -c "source venv/bin/activate && python -m app.main" &
	@echo "FastAPI server started in background. Use 'make logs-data' to view logs."

logs-data:
	@echo "Tailing FastAPI server logs..."
	@ps aux | grep "python -m app.main" | grep -v grep | head -1 | awk '{print $$2}' | xargs -I {} tail -f /proc/{}/fd/1 2>/dev/null || echo "FastAPI server not running. Start with 'make dev-data'"

# Combined development targets
dev-all:
	@echo "Starting both Go backend and FastAPI servers..."
	@make dev-backend
	@make dev-data
	@echo "Both servers started. Use 'make logs-all' to view combined logs."

logs-all:
	@echo "Tailing both server logs..."
	@echo "=== Go Backend Logs ==="
	@make logs-backend &
	@echo "=== FastAPI Logs ==="
	@make logs-data

# Database targets
db:
	@echo "Starting analytics databases..."
	docker compose -f deploy/docker-compose.yml up -d

down:
	@echo "Stopping analytics databases..."
	docker compose -f deploy/docker-compose.yml down

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
	@echo "Killing processes on ports 9000 and 9001..."
	@-fuser -k 9000/tcp 2>/dev/null || true
	@-fuser -k 9001/tcp 2>/dev/null || true

# Test targets
test:
	@echo "Running tests..."
	go test ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Run with auth disabled (development)
run-dev: build
	@echo "Running AIR with authentication disabled..."
	./bin/air --data data --config config-dev.yaml --auth

# Run with default config
run: build
	@echo "Running AIR..."
	./bin/air --data data --config config.yaml

# Help
help:
	@echo "Available targets:"
	@echo "  check       - Check environment and dependencies"
	@echo "  dev-backend  - Start backend server in development mode"
	@echo "  logs-backend - Tail Go backend server logs"
	@echo "  dev-ui       - Start UI (not implemented)"
	@echo "  dev-data     - Start FastAPI data processing server"
	@echo "  logs-data    - Tail FastAPI server logs"
	@echo "  dev-all      - Start both Go backend and FastAPI servers"
	@echo "  logs-all     - Tail both server logs"
	@echo "  db          - Start analytics databases with Docker"
	@echo "  down        - Stop analytics databases"
	@echo "  openapi-gen - Generate OpenAPI client/server code"
	@echo "  cli         - Build CLI tool"
	@echo "  build       - Build all binaries"
	@echo ""
	@echo "Quick Start:"
	@echo "  make dev-all     # Start both servers"
	@echo "  make logs-all    # View logs"
	@echo "  make clean-ports # Stop everything"
	@echo ""
	@echo "Health Checks:"
	@echo "  curl http://localhost:9000/health      # Go backend"
	@echo "  curl http://localhost:9001/v1/py/health # FastAPI"
	@echo ""
	@echo "For detailed guide, see SPEC-RUNNING.md"
