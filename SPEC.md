# AIR (AI Reporter) - Project Specification

## Overview

AIR is a natural language to SQL reporting system that uses AI to convert business questions into structured reports. It features a SQLite control-plane for metadata management and connects to multiple external analytics databases (TimescaleDB/PostgreSQL/MySQL) for data querying.

## Core Workflow

**Natural Language** → **Scope (Markdown)** → **IR (JSON)** → **SQL (per-engine)** → **Execute** → **AI QA** → **Saved Report** with parameters

## Multi-Datasource Architecture

AIR supports multiple analytics database sources and file datasets from day one:
- **SQLite Control-Plane**: Stores all AIR metadata (scopes, reports, runs, analyses)
- **Multiple Analytics Sources**: Register and connect to many external databases
- **File Datasets**: Process CSV/Parquet/JSONL files via FastAPI microservice
- **Redis Integration**: WebSocket management, live chat, and real-time features
- **Read-Only Operations**: All analytics databases are accessed read-only
- **Engine-Agnostic IR**: Intermediate Representation works across all database types
- **Dialect-Aware SQL**: Generate engine-specific SQL for each database type
- **Unified Interface**: Same API for both database and file-based data sources

## Architecture

```
air/
├─ api/
│  └─ openapi.yaml               # REST contract (public + internal FastAPI endpoints)
├─ cmd/
│  ├─ api/                       # main.go (Gin server bootstrap)
│  └─ cli/                       # Cobra CLI (calls REST/WS; no business logic)
├─ internal/
│  ├─ config/                    # Viper loader (YAML only), validation
│  ├─ store/                     # GORM (SQLite) models + repos (scopes/reports/runs/notes)
│  ├─ datasource/                # registry, connectors, health, pooling for multiple DBs
│  ├─ learn/                     # learn per datasource
│  ├─ scope/                     # scope → IR
│  ├─ sqlgen/                    # IR → SQL (dialect-aware)
│  ├─ reports/                   # defs, params, bind to datasource, runs
│  ├─ qa/                        # deterministic checks + LLM analysis
│  ├─ llm/                       # adapters (OpenAI, Ollama), router (auto/manual)
│  ├─ python/                    # FastAPI client for file processing
│  ├─ redis/                     # Redis client and WebSocket management
│  ├─ websocket/                 # WebSocket hub with Redis backend
│  ├─ transport/
│  │  ├─ rest/                   # Gin handlers (generated stubs + thin glue)
│  │  └─ ws/                     # WebSocket handlers (chat/progress/token streaming)
│  └─ telemetry/                 # logging/metrics
├─ clients/
│  └─ go/                        # OpenAPI-generated Go client (used by CLI)
├─ python/                       # FastAPI microservice for file processing
│  ├─ main.py                    # FastAPI application
│  ├─ models/                    # Pydantic models
│  ├─ services/                  # Business logic
│  ├─ requirements.txt           # Python dependencies
│  └─ Dockerfile                 # Container configuration
├─ deploy/
│  ├─ docker-compose.yml         # Postgres/Timescale + FastAPI for local testing
│  └─ init/                      # DB bootstrap (extensions)
├─ scripts/
│  ├─ check.sh                   # docker/pg/ollama/fastapi checks
│  └─ devdb.sh                   # spin up analytics DB for dev
├─ Makefile
├─ SPEC-PY.md                    # FastAPI microservice specification
└─ data/
   ├─ config.yaml                # primary config (Viper loads this)
   └─ config-dev.yaml            # development config
```

## Configuration

### Primary Config (`data/config.yaml`)

```yaml
server:
  host: 0.0.0.0
  port: 9000
  ws_enabled: true
  auth:
    enabled: true
    jwt_secret: "your-secret-key-here"
    token_expiry: "24h"

control_plane:            # AIR's own metadata store (GORM -> SQLite)
  driver: sqlite          # fixed to sqlite for MVP
  dsn: "file:air.db?_fk=1"

analytics_sources:        # list of external, READ-ONLY engines
  - id: "ts-dev"          # unique key
    kind: "timescaledb"   # "timescaledb" | "postgres" | "mysql"
    dsn: "postgres://user:pass@localhost:5432/energy?sslmode=disable"
    display_name: "Timescale Dev"
    default: true
  - id: "pg-sales"
    kind: "postgres"
    dsn: "postgres://reporter:***@pg:5432/sales"
    display_name: "Sales Warehouse (PG)"
  - id: "mysql-ops"
    kind: "mysql"
    dsn: "user:pass@tcp(localhost:3306)/ops"
    display_name: "Ops MySQL"
  - id: "energy-files"    # file-based datasource
    kind: "files"
    base_path: "/data/files/energy"
    display_name: "Energy Files (CSV/Parquet)"

python:                   # FastAPI microservice configuration
  enabled: true
  base_url: "http://localhost:9001"
  timeout: "90s"
  auth_shared_secret: "change-me-in-production"
  resource_limits:
    memory_mb: 2048
    workers: 2

redis:                    # Redis configuration for WebSocket and caching
  enabled: true
  url: "redis://localhost:6379/0"
  password: ""            # optional
  db: 0
  max_retries: 3
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  pool_size: 10
  min_idle_conns: 5

websocket:                # WebSocket configuration
  enabled: true
  buffer_size: 1024
  read_buffer_size: 4096
  write_buffer_size: 4096
  handshake_timeout: "10s"
  ping_period: "54s"
  pong_wait: "60s"
  max_message_size: 512
  enable_compression: true

chat:                     # Live chat configuration
  enabled: true
  message_retention: "24h"
  typing_timeout: "5s"
  presence_timeout: "5m"
  max_room_size: 100
  ai_streaming: true
  ai_response_timeout: "30s"

files:                    # file processing configuration
  allowed_ext: [".csv", ".parquet", ".jsonl"]
  max_rows_return_json: 50000
  infer_rows: 20000
  base_path: "/data/files"

models:
  chat_primary: "openai"        # openai | llama3
  chat_backup:  "llama3"
  sql_primary:  "sqlcoder"
  openai:
    model: "gpt-4o-mini"
    api_key: ""                 # optional if not using OpenAI
  ollama:
    host: "http://localhost:11434"
    llama3_model: "llama3"
    sqlcoder_model: "sqlcoder"
  embeddings:
    provider: "openai"          # or "ollama"
    model: "text-embedding-3-small"

safety:
  default_row_limit: 5000
  max_row_limit: 100000
  enforce_time_filter_days: 370

telemetry:
  level: "info"
```

## FastAPI Integration

AIR includes a **private FastAPI microservice** for file dataset processing:

### Purpose
- **File Processing**: Handle CSV/Parquet/JSONL datasets that can't be queried via SQL
- **Data Science Operations**: EDA, profiling, validation, and analytics on file data
- **Internal Service**: Only accessible by the Go backend, not public clients
- **Unified Experience**: Same natural language interface for both databases and files

### Key Features
- **Schema Inference**: Automatically detect file structure and data types
- **Query Execution**: Process file datasets using query plans (filters, aggregations, joins)
- **Data Analysis**: Perform EDA, outlier detection, correlation analysis
- **Multiple Formats**: Support CSV, Parquet, and JSONL files
- **Resource Management**: Memory limits, worker pools, and timeout controls
- **Arrow Integration**: Efficient binary data transfer for large datasets
- **OpenAPI Integration**: Uses same OpenAPI spec as Go backend for consistency

### Integration Pattern
1. **Go Backend** receives natural language query
2. **Go Backend** determines if query targets files or databases
3. **For Files**: Go calls FastAPI with query plan
4. **FastAPI** processes files and returns results
5. **Go Backend** streams results to client via WebSocket
6. **Unified Storage**: All results stored in SQLite with same metadata

### Security
- **Internal Only**: No public access to FastAPI service
- **HMAC Authentication**: Go signs requests with shared secret
- **Resource Limits**: Memory, row, and time constraints per request
- **Path Validation**: Restricted file access within configured base paths

### File-Based AI Learning Sessions

AIR includes a **simplified learning workflow** that demonstrates the full system using file datasets:

- **Interactive Learning**: Users can upload files and interact with AI to understand data
- **Scope Building**: AI helps build analysis plans through natural language conversation
- **Query Generation**: SQLCoder generates file processing queries
- **API Generation**: Successful analyses become reusable API endpoints
- **Educational Value**: Easy way to understand AIR's capabilities with familiar file data

See [SPEC-FILE-AI.md](./SPEC-FILE-AI.md) for detailed file-based AI learning session specification.

See [SPEC-PY.md](./SPEC-PY.md) for detailed FastAPI microservice specification.

## Redis Integration & Live Chat

AIR includes **Redis integration** for WebSocket management and live chat functionality:

### Purpose
- **WebSocket Management**: Handle real-time connections with persistence and scalability
- **Live Chat**: Enable real-time AI conversations with streaming responses
- **Session Management**: Track user sessions, presence, and typing indicators
- **Message Queuing**: Queue AI requests and stream responses back to users
- **Multi-Instance Support**: Scale horizontally with shared state

### Key Features
- **Real-time Messaging**: Instant message delivery and AI responses
- **AI Streaming**: Stream AI responses as they're generated (token by token)
- **Presence Management**: Track who's online/offline in real-time
- **Typing Indicators**: Show when users are typing
- **Message Persistence**: Store chat history for reconnection
- **Room Management**: Support for group chats and channels
- **Rate Limiting**: Prevent AI API abuse per user
- **Session Recovery**: Reconnect users to their chat sessions

### Redis Data Structures

```redis
# Chat messages (sorted set by timestamp)
ZADD chat:user:123 1640995200 "msg_id_1"
HSET chat:msg:msg_id_1 content "Hello AI!" user_id "123" type "user" timestamp "2025-01-01T00:00:00Z"

# Active WebSocket sessions
HSET sessions:user:123 ws_id "ws_456" last_seen 1640995200 room "general"

# User presence (who's online)
SADD online_users "user_123"
EXPIRE online_users 300  # 5 minute timeout

# Typing indicators
SET typing:user:123 "general" EX 5  # 5 second timeout

# AI response queue
LPUSH ai_queue:user:123 "ai_request_id_789"
HSET ai_request:ai_request_id_789 prompt "What is energy consumption?" user_id "123" status "processing"

# Chat rooms
SADD room:general:members "user_123" "user_456"
HSET room:general:info name "General Chat" created_at "2025-01-01T00:00:00Z"
```

### WebSocket Channels

```go
// Real-time channels for different purposes
channels := map[string]chan []byte{
    "chat:user:123",           // Personal chat messages
    "chat:room:general",       // Group chat messages
    "ai:stream:user:123",      // AI streaming responses
    "typing:user:123",         // Typing indicators
    "presence:online",         // Online users list
    "system:notifications",    // System notifications
}
```

### Live Chat Integration

**1. Real-time AI Conversations**
```javascript
// Frontend WebSocket connection
const ws = new WebSocket('ws://localhost:9000/v1/ws');

// Send user message
ws.send(JSON.stringify({
    type: "chat_message",
    content: "What's our energy consumption trend?",
    user_id: "user_123",
    room: "general"
}));

// Receive AI responses as they stream
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === "ai_stream") {
        appendToChat(data.content); // Partial AI response
    } else if (data.type === "ai_complete") {
        showCompleteResponse(data.response);
    }
};
```

**2. AI Streaming Backend**
```go
// Stream AI responses to WebSocket
func (h *WebSocketHub) StreamAIResponse(userID string, prompt string) {
    // Send to AI service for processing
    response := h.aiService.StreamChat(prompt)
    
    for chunk := range response.Stream {
        h.SendToUser(userID, WebSocketMessage{
            Type: "ai_stream",
            Content: chunk,
            Timestamp: time.Now(),
        })
    }
    
    // Mark as complete
    h.SendToUser(userID, WebSocketMessage{
        Type: "ai_complete",
        Response: response.FullResponse,
        Timestamp: time.Now(),
    })
}
```

**3. Presence & Typing Indicators**
```go
// User typing indicator
func (h *WebSocketHub) SetUserTyping(userID string, room string, isTyping bool) {
    if isTyping {
        h.redis.Set(fmt.Sprintf("typing:%s", userID), room, 5*time.Second)
    } else {
        h.redis.Del(fmt.Sprintf("typing:%s", userID))
    }
    
    h.BroadcastToRoom(room, WebSocketMessage{
        Type: "typing",
        UserID: userID,
        Room: room,
        IsTyping: isTyping,
    })
}
```

### Security & Performance
- **Authentication**: JWT tokens for WebSocket connections
- **Rate Limiting**: Per-user limits on AI requests and messages
- **Message Validation**: Sanitize and validate all incoming messages
- **Connection Limits**: Maximum connections per user and room
- **Memory Management**: Automatic cleanup of old messages and sessions

## Data Model (SQLite with GORM)

### Core Tables

- `datasources(id TEXT PK, kind TEXT, dsn TEXT, display_name TEXT, is_default BOOL, created_at, updated_at)`
- `schema_notes(id, datasource_id, object TEXT, chunk INT, md TEXT, md_hash TEXT, created_at)`
- `scopes(id, name, status, created_at, updated_at)`
- `scope_versions(id, scope_id, version, scope_md TEXT, ir_json JSON, created_at)`
- `reports(id, key UNIQUE, title, owner, archived, created_at, updated_at)`
- `report_versions(id, report_id, version, scope_version_id, datasource_id TEXT NULL, def_json JSON, checksum TEXT, status, created_at)`
- `report_runs(id, report_id, report_version_id, datasource_id, params_json JSON, sql_text TEXT, row_count INT, started_at, finished_at, status, error_text)`
- `report_samples(run_id, seq, row_json JSON, PRIMARY KEY(run_id, seq))`
- `report_analyses(id, run_id, model_used, rubric_version, verdict_json JSON, analysis_md TEXT, created_at)`

### Multi-Datasource Notes

- `datasources` table stores all registered analytics database connections
- `schema_notes` now includes `datasource_id` to associate schema info with specific databases
- `report_versions.datasource_id` is optional:
  - If set → report is **bound** to a specific datasource
  - If null → report is **portable**; datasource chosen at runtime
- `report_runs` includes `datasource_id` to track which database was used for execution

### Report Definition Schema

Stored in `report_versions.def_json`:

```json
{
  "version": "1",
  "ir": {
    "entities": ["energy.measurements"],
    "metrics": [{"name":"kwh","agg":"sum"}],
    "dimensions": ["site","day"],
    "filters": [
      {"field":"timestamp","op":">=","value":"{{dateFrom}}"},
      {"field":"timestamp","op":"<","value":"{{dateTo}}"}
    ],
    "grain": "1 day",
    "order": [{"field":"day","dir":"asc"}],
    "limit": 5000
  },
  "params": {
    "exposed": ["dateFrom","dateTo","name","people"],
    "schema": {
      "type": "object",
      "properties": {
        "dateFrom": {"type":"string","format":"date-time"},
        "dateTo":   {"type":"string","format":"date-time"},
        "name":     {"type":"string"},
        "people":   {"type":"array","items":{"type":"string"}}
      },
      "required": ["dateFrom","dateTo"]
    },
    "defaults": {
      "dateFrom": "2025-09-01T00:00:00Z",
      "dateTo":   "2025-10-01T00:00:00Z"
    }
  },
  "visuals": [
    {"type":"line", "x":"day", "y":"kwh", "series":"site"}
  ],
  "notes": "Daily energy use by site with date range parameters."
}
```

## Workflows

### A) Learn Data Sources (Multi-Datasource + Files)
- Register multiple analytics databases via `/v1/datasources`
- Register file-based datasources with base paths
- Introspect each datasource individually via `POST /v1/learn?datasource_id=...`
- **For Databases**: Collect tables/columns/FKs/index hints per datasource
- **For Files**: Call FastAPI to infer schema from file samples
- Generate schema notes (Markdown) → store in SQLite with `datasource_id`
- Special handling for TimescaleDB hypertables and time columns
- Health checks and connection pooling per datasource

### B) Scope with User
- Natural language question → clarification questions → Scope (Markdown)
- Store in `scopes` + `scope_versions` with versioning

### C) IR Build
- Convert scope to Intermediate Representation (IR) JSON
- Store in `scope_versions.ir_json`

### D) Query Generation + Guardrails (Dialect-Aware)
- **For Databases**: Use SQLCoder for SQL generation tuned to specific datasource type
- **For Files**: Convert IR to query plan for FastAPI processing
- Generate engine-specific SQL:
  - Timescale/Postgres: use `time_bucket`, `timestamptz`, etc.
  - MySQL: use `DATE_FORMAT`/window functions as available
- Enforce read-only operations across all engines
- Inject time predicates and LIMIT clauses
- Apply safety constraints per engine type

### E) Execute + Sample + QA (Multi-Datasource + Files)
- **For Databases**: Execute SQL on chosen analytics datasource
- **For Files**: Call FastAPI with query plan, receive Arrow/JSON results
- Persist `report_run` with results and `datasource_id`
- Save sample rows (capped at ~50)
- Run deterministic checks + LLM analysis
- **Enhanced QA**: Use FastAPI for additional EDA and validation on results
- Store QA verdict with file-specific insights

### F) Save Report Definition (Bound or Portable)
- Create `reports` + `report_versions.def_json`
- Define exposed parameters with JSON Schema validation
- Choose binding strategy:
  - **Bound**: Set `datasource_id` → report tied to specific database
  - **Portable**: Leave `datasource_id` null → choose at runtime
- Enable parameterized execution without LLM

## API Surface (OpenAPI/Swagger)

### REST Endpoints

#### Datasources
- `GET /v1/datasources` → list all datasources with health status
- `POST /v1/datasources` → create new datasource connection
- `POST /v1/datasources/{id}/health` → test datasource connection
- `DELETE /v1/datasources/{id}` → remove datasource (if unused)

#### Learn & Schema
- `POST /v1/learn?datasource_id=...` → introspect specific datasource
- `GET /v1/schema/{datasource_id}` → get schema notes for datasource

#### Scope & IR
- `POST /v1/ask` → Natural language → scope draft (Markdown)
- `POST /v1/scopes` / `POST /v1/scopes/{id}/version` → create/approve scope
- `POST /v1/ir/build` → {scope_version_id} → IR JSON

#### SQL Generation
- `POST /v1/sql` → {ir, datasource_id} → {sql, safety_report}

#### Reports
- `POST /v1/reports` → create report (key/title)
- `POST /v1/reports/{key}/versions` → upload def_json (with optional datasource_id)
- `POST /v1/reports/{key}/run` → execute with parameters
  - Bound reports: use stored datasource_id
  - Portable reports: require ?datasource_id=... parameter

#### Analysis & Export
- `POST /v1/runs/{run_id}/analyze` → AI QA verdict
- `GET /v1/reports/{key}/export?format=json|yaml` / `POST /v1/reports/import`
- `GET /v1/ai/tools` → tool/function definitions

### Authentication

- JWT-based authentication with configurable secret
- `--auth disabled` flag to disable authentication for development
- Token-based access control for all API endpoints

### Heavy Processing APIs

For operations that may take longer than 30 seconds (large datasets, complex analysis, bulk operations):

#### Async Processing
- `POST /v1/process/learn` → bulk learn multiple datasources
  - **Input**: `{datasource_ids: ["ds1", "ds2"], options: {deep: true, include_stats: true}}`
  - **Output**: `{job_id: "job_123", estimated_duration: "5m", websocket_channel: "learn/job_123"}`
  - **WebSocket**: Real-time progress via `learn/job_123` channel

- `POST /v1/process/analyze` → comprehensive data analysis
  - **Input**: `{datasource_id: "ds1", scope: "energy_analysis", options: {correlations: true, outliers: true, trends: true}}`
  - **Output**: `{job_id: "job_456", estimated_duration: "3m", websocket_channel: "analyze/job_456"}`

- `POST /v1/process/bulk-reports` → execute multiple reports
  - **Input**: `{reports: [{"key": "daily", "params": {...}}, {"key": "weekly", "params": {...}}], datasource_id: "ds1"}`
  - **Output**: `{job_id: "job_789", estimated_duration: "2m", websocket_channel: "bulk/job_789"}`

#### Job Management
- `GET /v1/jobs/{job_id}` → get job status and progress
  - **Output**: `{job_id: "job_123", status: "running|completed|failed", progress: 45, current_step: "Analyzing correlations", estimated_remaining: "2m30s"}`
- `DELETE /v1/jobs/{job_id}` → cancel running job
- `GET /v1/jobs` → list all jobs (running, completed, failed)

#### WebSocket Progress Updates
- **Channel**: `process/{job_id}`
- **Message Types**:
  - `status`: `{step: "Connecting to database", progress: 10, timestamp: "2025-09-30T08:30:00Z"}`
  - `progress`: `{step: "Analyzing 1M rows", progress: 45, rows_processed: 450000, timestamp: "2025-09-30T08:31:00Z"}`
  - `result`: `{data: {...}, metrics: {...}, completed_at: "2025-09-30T08:35:00Z"}`
  - `error`: `{error: "Connection timeout", step: "Database query", timestamp: "2025-09-30T08:32:00Z"}`

### WebSocket Streaming

- `GET /v1/ws` → multiplexed channels with Redis backend
- **Channels**: 
  - `learn/<job>` - Database learning progress
  - `run/<run_id>` - Report execution progress  
  - `chat/<session>` - Live chat messages
  - `process/<job_id>` - Heavy processing jobs
  - `ai:stream:<user_id>` - AI response streaming
  - `typing:<user_id>` - Typing indicators
  - `presence:online` - User presence updates
- **Message format**: `{ channel, type: status|token|result|error|ai_stream|typing|presence, payload, ts, user_id? }`
- **Redis Integration**: All WebSocket state persisted in Redis for scalability

## CLI (Cobra)

### Generic Mode (Multi-Datasource)

```bash
# Authentication
aircli --auth disabled post /v1/datasources --json '{"id":"ts-dev","kind":"timescaledb","dsn":"..."}'
aircli --token "jwt-token" post /v1/learn --query datasource_id=ts-dev

# Datasource Management
aircli get /v1/datasources
aircli post /v1/datasources/ts-dev/health

# Report Management
aircli post /v1/reports/my_daily/versions --json @def.json
aircli post /v1/reports/my_daily/run --json @params.json --query datasource_id=pg-sales

# SQL Generation
aircli post /v1/sql --json @ir.json --query datasource_id=mysql-ops

# Heavy Processing
aircli post /v1/process/learn --json '{"datasource_ids":["ts-dev","pg-sales"],"options":{"deep":true}}'
aircli post /v1/process/analyze --json '{"datasource_id":"ts-dev","scope":"energy_analysis"}'
aircli get /v1/jobs/job_123
aircli ws --channel process/job_123

# WebSocket
aircli ws --channel run/<run_id>
```

## Model Routing

- **Chat tasks**: `chat_primary` (OpenAI/Llama3), fallback to `chat_backup`
- **SQL generation**: `sql_primary` (SQLCoder), fallback to chat_primary
- Per-request override: `force_model=openai|llama3|sqlcoder`

## Safety Guardrails (Per Engine)

- Block destructive operations across all engines: `INSERT|UPDATE|DELETE|DROP|ALTER|COPY|CALL`
- Require time predicates on time-series tables
- Enforce row limits and date span constraints
- Engine-specific safety checks:
  - TimescaleDB: hypertable time column validation
  - PostgreSQL: proper timestamp handling
  - MySQL: version-specific feature compatibility
- Optional `EXPLAIN` analysis for query optimization warnings

## Development Setup

### Docker Compose (Analytics DB + Redis)

```yaml
services:
  timescale:
    image: timescale/timescaledb:latest-pg16
    environment:
      POSTGRES_USER: air
      POSTGRES_PASSWORD: air
      POSTGRES_DB: energy
    ports: ["5432:5432"]
    volumes: ["pgdata:/var/lib/postgresql/data"]

  redis:
    image: redis:7-alpine
    container_name: air-redis
    ports: ["6379:6379"]
    volumes: ["redis_data:/data"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  pgdata: {}
  redis_data: {}
```

### Makefile Targets

```make
check:
	bash scripts/check.sh   # docker, analytics DB reachable, redis up, ollama up, fastapi up

dev-backend:
	go run ./cmd/api --data data --config config-dev.yaml --auth

dev-python:
	cd python && uvicorn main:app --host 0.0.0.0 --port 9001 --reload

dev-ui:
	@echo "UI not implemented yet"

db:
	docker compose -f deploy/docker-compose.yml up -d

down:
	docker compose -f deploy/docker-compose.yml down

openapi-gen:
	oapi-codegen -generate types,server -package restapi -o internal/transport/rest/openapi.gen.go api/openapi.yaml
	oapi-codegen -generate client -package apiclient -o clients/go/client.gen.go api/openapi.yaml
	oapi-codegen -generate client -package pyclient -o internal/python/client.gen.go api/openapi.yaml
	cd python && datamodel-codegen --input ../api/openapi.yaml --output models/openapi_models.py

cli:
	go build -o bin/aircli ./cmd/cli

python-deps:
	cd python && pip install -r requirements.txt

python-test:
	cd python && pytest

build-all: cli python-deps
	go build -o bin/air ./cmd/api
```

## v0.1 Deliverables

- SQLite control-plane with GORM models and migrations
- **Multi-datasource registry** with health checks and connection pooling
- External analytics connectors (Timescale/Postgres/MySQL) with dialect-aware SQL generation
- **FastAPI microservice** for file dataset processing (CSV/Parquet/JSONL)
- **Per-datasource learning** and schema note generation (databases + files)
- Scope management with Markdown storage and versioning
- IR build and storage on scope versions (engine-agnostic)
- **Dialect-aware SQL generation** with SQLCoder and engine-specific safety guardrails
- **File query processing** with Arrow IPC and JSON data transport
- Report execution with sampling and AI QA (multi-datasource + file support)
- **Enhanced QA** with FastAPI-powered EDA and validation
- **Bound and portable report definitions** with parameter schema validation
- Parameterized report execution across multiple datasources and file datasets
- **Redis-powered WebSocket streaming** for real-time updates and live chat
- **Live AI chat** with streaming responses and presence management
- **File-based AI learning sessions** for interactive data exploration and API generation
- OpenAPI specification with generated server stubs and Go client
- CLI that uses REST/WS exclusively with multi-datasource support
- JWT authentication with optional disable flag
- **Unified interface** for both database and file-based data sources

## Key Design Principles

1. **SQLite Control-Plane**: All AIR metadata stored in portable SQLite database
2. **Multi-Datasource Support**: Register and connect to many external analytics databases
3. **File Dataset Support**: Process CSV/Parquet/JSONL files via FastAPI microservice
4. **Read-Only Operations**: All analytics databases accessed read-only with strong guardrails
5. **Engine-Agnostic IR**: Intermediate Representation works across all database types
6. **Dialect-Aware SQL**: Generate engine-specific SQL optimized for each database type
7. **Unified Interface**: Same natural language interface for databases and files
8. **Bound or Portable Reports**: Choose between datasource-specific or cross-database reports
9. **Versioned Definitions**: Report definitions stored as versioned JSON with schema validation
10. **OpenAPI Contract**: Single source of truth for REST/WS/CLI interfaces and internal FastAPI
11. **AI Tool Integration**: OpenAPI serves as tool definition for AI agents
12. **Consistent APIs**: Both Go backend and FastAPI microservice use same OpenAPI specification
13. **Safety First**: Read-only operations with comprehensive, engine-specific guardrails
14. **Internal Services**: FastAPI microservice only accessible by Go backend
15. **Portable Deployment**: SQLite enables simple, portable deployments
16. **Dual API Design**: Standard APIs for quick operations (<30s), Heavy Processing APIs for long-running tasks with WebSocket progress updates
17. **Redis Integration**: WebSocket state, live chat, and real-time features powered by Redis
18. **Live Chat**: Real-time AI conversations with streaming responses and presence management
