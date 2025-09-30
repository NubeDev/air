# AIR Stack - Quick Start Reference

## 🚀 Essential Commands

### Start Everything
```bash
make dev-all          # Start both servers
make logs-all         # View all logs
```

### Individual Services
```bash
make dev-backend      # Start Go backend only
make dev-data         # Start FastAPI only
make logs-backend     # View Go logs
make logs-data        # View FastAPI logs
```

### Stop Everything
```bash
make clean-ports      # Kill all processes
```

## 🔍 Quick Health Checks

```bash
# Go backend
curl http://localhost:9000/health

# FastAPI server
curl http://localhost:9001/v1/py/health

# Test integration
curl -X POST http://localhost:9000/v1/fastapi/test/energy
```

## 🐛 Common Issues

### Port Already in Use
```bash
make clean-ports      # Kill all processes
make dev-all          # Restart
```

### Services Not Responding
```bash
# Check what's running
ps aux | grep "go run"
ps aux | grep "python -m app.main"

# Check ports
lsof -i :9000
lsof -i :9001
```

### Database Issues
```bash
# Reset SQLite database
rm data/air.db
make dev-backend      # Recreates database
```

## 📊 Log Monitoring

```bash
# Real-time logs
make logs-all         # Both servers
make logs-backend     # Go only
make logs-data        # FastAPI only

# Manual log checking
tail -f /proc/$(pgrep -f "go run")/fd/1
```

## 🧪 Testing

```bash
# Test energy data processing
curl -X POST http://localhost:9000/v1/fastapi/test/energy

# Test file discovery
curl -X POST http://localhost:9000/v1/fastapi/test/discover

# Direct FastAPI test
curl -X POST http://localhost:9001/v1/py/infer_schema \
  -H "Content-Type: application/json" \
  -d '{"datasource_id":"test","uri":"../testdata/ts-energy.csv","infer_rows":50}'
```

## 📁 Key Files

- **Config**: `data/config-dev.yaml`
- **Database**: `data/air.db`
- **Test Data**: `testdata/ts-energy.csv`
- **FastAPI**: `dataserver/app/main.py`
- **Go API**: `cmd/api/main.go`

## 🔧 Development Workflow

1. **Start**: `make dev-all`
2. **Monitor**: `make logs-all` (in another terminal)
3. **Test**: Use curl commands above
4. **Debug**: Check logs for errors
5. **Stop**: `make clean-ports`

---
*For detailed information, see `SPEC-RUNNING.md`*
