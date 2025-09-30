# AIR API Server

This directory contains the AIR API server implementation, broken down into focused, manageable files.

## File Structure

### `main.go`
- Entry point for the API server
- Handles command-line flags and configuration loading
- Creates and starts the server

### `server.go`
- Contains the `Server` struct and related functionality
- Handles server initialization, database setup, and cleanup
- Manages the overall server lifecycle

### `middleware.go`
- Contains global middleware setup
- CORS configuration
- Other cross-cutting concerns

### `routes.go`
- Defines all API routes and route groups
- Organizes endpoints by functionality (datasources, reports, etc.)
- Sets up authentication middleware per route group

### `handlers.go`
- Contains all HTTP request handlers
- Organized by functionality (datasource, report, scope handlers, etc.)
- Currently contains placeholder implementations

## Architecture

The server follows a clean separation of concerns:

1. **Configuration**: Loaded from YAML files with environment variable support
2. **Database**: SQLite control-plane with GORM for metadata management
3. **Datasources**: Multi-datasource registry for analytics databases
4. **Authentication**: JWT-based auth with optional disable flag
5. **Routes**: RESTful API with organized route groups
6. **Handlers**: Focused request handlers for each endpoint

## Development

### Building
```bash
go build -o bin/air ./cmd/api
```

### Running
```bash
# With authentication disabled (development)
./bin/air --config configs/air-dev.yaml

# With authentication enabled
./bin/air --config configs/air.yaml
```

### Testing
```bash
# Health check
curl http://localhost:8080/health

# List datasources
curl http://localhost:8080/v1/datasources
```

## Next Steps

The current implementation provides a solid foundation with:
- ✅ Multi-datasource support
- ✅ JWT authentication with disable flag
- ✅ Clean file organization
- ✅ Placeholder handlers for all endpoints

Future development should focus on implementing the actual business logic in the handler functions.
