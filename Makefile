.PHONY: check dev-backend dev-ui db down openapi-gen cli build clean test

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
	@echo "Starting AIR backend server..."
	go run ./cmd/api --data data --config config-dev.yaml --auth

dev-ui:
	@echo "UI not implemented yet"

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
	@echo "  dev-backend - Start backend server in development mode"
	@echo "  dev-ui      - Start UI (not implemented)"
	@echo "  db          - Start analytics databases with Docker"
	@echo "  down        - Stop analytics databases"
	@echo "  openapi-gen - Generate OpenAPI client/server code"
	@echo "  cli         - Build CLI tool"
	@echo "  build       - Build all binaries"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Install dependencies"
	@echo "  run-dev     - Run with authentication disabled"
	@echo "  run         - Run with default configuration"
	@echo "  help        - Show this help"
