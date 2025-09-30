# AIR Stack - Running & Debugging Guide

## Overview

This guide covers how to start, stop, debug, and monitor the complete AIR stack consisting of:
- **Go Backend** (Port 9000) - Main API server with SQLite control plane
- **FastAPI Server** (Port 9001) - Python data processing microservice
- **Redis** (Port 6379) - Job queue and caching for FastAPI
- **Analytics Databases** - TimescaleDB/PostgreSQL/MySQL (optional for development)

---

## Quick Start

### Start Everything
```bash
# Start both Go backend and FastAPI servers
make dev-all

# View logs from both servers
make logs-all
```

### Stop Everything
```bash
# Kill all processes on ports 9000 and 9001
make clean-ports
```

---

## Individual Service Management

### Go Backend Server (Port 9000)

**Start:**
```bash
make dev-backend
```

**View Logs:**
```bash
make logs-backend
```

**Stop:**
```bash
# Kill process on port 9000
fuser -k 9000/tcp
```

**Manual Start (for debugging):**
```bash
go run ./cmd/api --data data --config config-dev.yaml --auth
```

**Configuration:**
- Config file: `data/config-dev.yaml`
- Database: `data/air.db` (SQLite)
- Logs: Colored console output with service tags
- Auth: Disabled in dev mode (`--auth` flag)

### FastAPI Server (Port 9001)

**Start:**
```bash
make dev-data
```

**View Logs:**
```bash
make logs-data
```

**Stop:**
```bash
# Kill process on port 9001
fuser -k 9001/tcp
```

**Manual Start (for debugging):**
```bash
cd dataserver
source venv/bin/activate
python -m app.main
```

**Configuration:**
- Environment: `dataserver/env.example`
- Redis: `redis://localhost:6379/0`
- Logs: FastAPI/Uvicorn standard output

### Redis (Port 6379)

**Start with Docker:**
```bash
docker run --name air-redis -p 6379:6379 redis:7-alpine
```

**Stop:**
```bash
docker stop air-redis
docker rm air-redis
```

**Check Status:**
```bash
redis-cli ping
```

---

## Development Workflow

### 1. Full Stack Development
```bash
# Terminal 1: Start both servers
make dev-all

# Terminal 2: Monitor logs
make logs-all

# Terminal 3: Run tests/API calls
curl http://localhost:9000/health
curl http://localhost:9001/v1/py/health
```

### 2. Backend-Only Development
```bash
# Start only Go backend
make dev-backend

# View Go logs
make logs-backend

# Test Go API
curl http://localhost:9000/health
curl -X POST http://localhost:9000/v1/fastapi/test/energy
```

### 3. Data Processing Development
```bash
# Start only FastAPI
make dev-data

# View FastAPI logs
make logs-data

# Test FastAPI directly
curl http://localhost:9001/v1/py/health
curl -X POST http://localhost:9001/v1/py/infer_schema \
  -H "Content-Type: application/json" \
  -d '{"datasource_id":"test","uri":"../testdata/ts-energy.csv","infer_rows":50}'
```

---

## Debugging Guide

### Go Backend Debugging

**1. Enable Debug Logging:**
```yaml
# data/config-dev.yaml
telemetry:
  level: "debug"  # Change from "info" to "debug"
```

**2. Run with Verbose Output:**
```bash
# Run directly to see all output
go run ./cmd/api --data data --config config-dev.yaml --auth

# Or with debug flags
go run ./cmd/api --data data --config config-dev.yaml --auth -v
```

**3. Common Debug Endpoints:**
```bash
# Health check
curl http://localhost:9000/health

# FastAPI integration test
curl -X POST http://localhost:9000/v1/fastapi/test/energy

# File discovery test
curl -X POST http://localhost:9000/v1/fastapi/test/discover
```

**4. Database Debugging:**
```bash
# Check SQLite database
sqlite3 data/air.db ".tables"
sqlite3 data/air.db "SELECT * FROM datasources;"
```

### FastAPI Debugging

**1. Enable Debug Mode:**
```python
# dataserver/app/main.py
if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=9001,
        reload=True,  # Enable auto-reload
        log_level="debug"  # Enable debug logging
    )
```

**2. Run with Debug Output:**
```bash
cd dataserver
source venv/bin/activate
python -m app.main --log-level debug
```

**3. Test Individual Endpoints:**
```bash
# Health check
curl http://localhost:9001/v1/py/health

# Schema inference
curl -X POST http://localhost:9001/v1/py/infer_schema \
  -H "Content-Type: application/json" \
  -d '{"datasource_id":"test","uri":"../testdata/ts-energy.csv","infer_rows":50}'

# Check job status
curl http://localhost:9001/v1/py/jobs/1
```

**4. Redis Debugging:**
```bash
# Connect to Redis
redis-cli

# List all keys
KEYS *

# Check job data
GET job:1
```

---

## Log Analysis

### Go Backend Logs
```
10:24:33 [SERV] [INFO] Initializing AIR server
10:24:33 [DATA] [INFO] Database connected successfully
10:24:33 [HTTP] [INFO] HTTP request status=200 duration=1.4ms
```

**Log Format:** `time [SERVICE][LEVEL] message`
- **SERV** = Server initialization
- **DATA** = Database operations
- **HTTP** = HTTP requests
- **AI** = AI/FastAPI integration
- **AUTH** = Authentication
- **CONF** = Configuration

### FastAPI Logs
```
INFO:     127.0.0.1:34748 - "GET /v1/py/health HTTP/1.1" 200 OK
INFO:     Started server process [12345]
INFO:     Waiting for application startup.
```

**Log Format:** Standard FastAPI/Uvicorn logs with request details

---

## Troubleshooting

### Port Conflicts
```bash
# Check what's using ports
lsof -i :9000
lsof -i :9001
lsof -i :6379

# Kill specific processes
fuser -k 9000/tcp
fuser -k 9001/tcp
fuser -k 6379/tcp
```

### Service Not Starting
```bash
# Check if services are running
ps aux | grep "go run"
ps aux | grep "python -m app.main"
ps aux | grep redis

# Check port availability
netstat -tulpn | grep :9000
netstat -tulpn | grep :9001
```

### Database Issues
```bash
# Check SQLite database
ls -la data/air.db
sqlite3 data/air.db ".schema"

# Reset database (WARNING: deletes all data)
rm data/air.db
make dev-backend  # Will recreate database
```

### FastAPI Connection Issues
```bash
# Check if FastAPI is responding
curl -v http://localhost:9001/v1/py/health

# Check Redis connection
redis-cli ping

# Restart FastAPI
make clean-ports
make dev-data
```

---

## Performance Monitoring

### Go Backend Metrics
```bash
# Check response times in logs
grep "duration=" logs/air.log

# Monitor memory usage
ps aux | grep "go run" | awk '{print $4, $6}'
```

### FastAPI Metrics
```bash
# Check job processing times
curl http://localhost:9001/v1/py/jobs/1 | jq '.steps[].duration_ms'

# Monitor Redis memory
redis-cli info memory
```

---

## Development Tips

### 1. Hot Reloading
- **Go**: Use `air` tool for hot reloading (not implemented yet)
- **FastAPI**: Built-in reload with `reload=True`

### 2. Testing
```bash
# Test Go API
curl http://localhost:9000/health

# Test FastAPI
curl http://localhost:9001/v1/py/health

# Test integration
curl -X POST http://localhost:9000/v1/fastapi/test/energy
```

### 3. Data Testing
```bash
# Use test data
curl -X POST http://localhost:9001/v1/py/infer_schema \
  -H "Content-Type: application/json" \
  -d '{"datasource_id":"energy","uri":"../testdata/ts-energy.csv","infer_rows":50}'
```

### 4. Log Monitoring
```bash
# Follow Go logs
make logs-backend

# Follow FastAPI logs
make logs-data

# Follow both
make logs-all
```

---

## Production Considerations

### Security
- Enable authentication: Remove `--auth` flag
- Use proper JWT secrets
- Restrict network access

### Performance
- Use production Redis configuration
- Configure proper database connections
- Set appropriate log levels

### Monitoring
- Use structured logging (JSON format)
- Set up health check endpoints
- Monitor resource usage

---

## Makefile Reference

```bash
# Development
make dev-backend    # Start Go backend
make dev-data       # Start FastAPI
make dev-all        # Start both

# Logging
make logs-backend   # Tail Go logs
make logs-data      # Tail FastAPI logs
make logs-all       # Tail both

# Utilities
make clean-ports    # Kill all processes
make check          # Check environment
make help           # Show all targets
```

This guide should help you effectively manage, debug, and monitor the AIR stack during development! 🚀
