# AIR Implementation Plan - Database + File Processing

## 🎯 Architecture Overview

AIR supports dual processing paths for both database and file-based data analysis.

- DB execution is done in Go against registered analytics datasources (SQLite, Postgres, MySQL, Timescale).
- File reads and ad-hoc file processing are done in Python FastAPI (no ingestion by default).
- Scope generation/refinement uses Llama/ChatGPT; SQL generation uses SQLCoder only.

## 🚦 Phase 0: Datasource Registration (SQLite + PG + MySQL + Files)
- Register analytics datasources via Go using `/v1/datasources` and `internal/datasource` registry:
  - SQLite: `dsn: "file:/data/analytics.db?_fk=1"`
  - Postgres: `dsn: "postgres://user:pass@host:5432/db?sslmode=disable"`
  - MySQL: `dsn: "user:pass@tcp(host:3306)/db"`
  - Files: `base_path: "/data/files"`
- Ensure separate SQLite analytics DB (not `air.db`).

## 🚀 Implementation Tasks

### Phase 1: Complete Database Workflow (HIGH)

#### 1.1 Reports Service (`internal/services/reports_service.go`)
- [ ] `CreateScope(req CreateScopeRequest) (*Scope, error)`
- [ ] `GetScope(id uint) (*Scope, error)`
- [ ] `CreateScopeVersion(scopeID uint, req CreateScopeVersionRequest) (*ScopeVersion, error)`
- [ ] `CreateReport(req CreateReportRequest) (*Report, error)`
- [ ] `GetReport(key string) (*Report, error)`
- [ ] `CreateReportVersion(reportKey string, req CreateReportVersionRequest) (*ReportVersion, error)`
- [ ] `RunReport(reportKey string, req RunReportRequest) (*ReportRun, error)`
- [ ] `ExportReport(reportKey string, format string) ([]byte, error)`

#### 1.2 AI Service (DB) (`internal/services/ai_service.go`)
- [ ] `BuildIR(req BuildIRRequest)` → use Llama/ChatGPT to turn scope markdown into IR JSON
- [ ] `GenerateSQLFromIR(req GenerateSQLRequest)` → SQLCoder only; no heuristic fallback
- [ ] `AnalyzeRun(runID, req)` → use Llama/ChatGPT to analyze execution results

#### 1.3 Datasource Learning (`internal/services/datasource_service.go`)
- [ ] `LearnDatasource(req)` → learn schema from DB; store `SchemaNote`
- [ ] `GetSchema(datasourceID)` → return learned schema

### Phase 2: Complete File Learning Workflow (HIGH)

- File reads stay in Python. No ingestion into DB unless explicitly requested later.

#### 2.1 Session Learning Endpoints (`cmd/api/handlers/sessions/`)
- [ ] `POST /v1/sessions/{id}/ask` → Llama/ChatGPT Q&A
- [ ] `POST /v1/sessions/{id}/scope/build` → Draft scope (LLM)
- [ ] `POST /v1/sessions/{id}/scope/refine` → Refine scope (LLM)
- [ ] `POST /v1/sessions/{id}/query/generate` → Build file query plan (LLM-assisted IR→plan)
- [ ] `POST /v1/sessions/{id}/execute` → Call Python to execute plan
- [ ] `POST /v1/sessions/{id}/analyze` → LLM analysis on results
- [ ] `POST /v1/sessions/{id}/save` → Save as GeneratedReport

#### 2.2 Session Service (`internal/services/session_service.go`)
- [ ] `AskQuestion`, `BuildScope`, `RefineScope`
- [ ] `GenerateQuery`, `ExecuteAnalysis`, `AnalyzeResults`, `SaveAsAPI`

#### 2.3 FastAPI Integration (`internal/services/fastapi_client.go`)
- [ ] `LearnFile` (infer schema, profiling)
- [ ] `ExecuteFileQuery` (run query plan)
- [ ] `AnalyzeFileResults`

### Phase 3: Unified Learning Pattern (MEDIUM)
- [ ] `learning_service.go` unifies DB/file learning
- [ ] `scope_builder.go` centralizes scope→IR (LLM)
- [ ] `query_generator.go` centralizes IR→SQL (SQLCoder) or IR→plan (files)

### Phase 4: UI Integration (MEDIUM)
- [ ] `GET /v1/generated/reports/{id}/schema` (JSON Schema for forms)
- [ ] `GET /v1/reports/{id}/schema`
- [ ] Convert OpenAPI → JSON Schema; client-side validation and hints

## 📁 File Structure (Keep Current Layout)

```
cmd/api/
├── handlers/
│   ├── ai/           # AI tools, chat, SQL generation
│   ├── db/           # Database datasource management
│   ├── fastapi/      # Python FastAPI integration
│   ├── generated_reports/  # File-based generated reports
│   ├── health/       # Health checks
│   ├── reports/      # Database reports and scopes
│   ├── sessions/     # File-based learning sessions
│   └── websocket/    # WebSocket handlers
├── middleware/       # Auth, logging middleware
├── routes/          # Route definitions
├── main.go
└── server.go

internal/
├── auth/            # JWT authentication
├── config/          # Configuration management
├── datasource/      # Database connectors
├── llm/            # Ollama AI client
├── logger/         # Logging
├── redis/          # Redis client
├── services/       # Business logic services
├── store/          # Database models
└── websocket/      # WebSocket hub

dataserver/         # Python FastAPI backend (DON'T CHANGE)
├── app/
│   ├── api/        # API endpoints
│   ├── core/       # Configuration
│   ├── models/     # Pydantic models
│   └── services/   # Data processing
└── requirements.txt
```

## ✅ Success Criteria

### Database Processing
- [ ] Datasources registered (SQLite, Postgres, MySQL, Files)
- [ ] Scope → IR (LLM)
- [ ] IR → SQL (SQLCoder)
- [ ] SQL executes on selected datasource via Go
- [ ] Reports saved and executed with parameters

### File Processing
- [ ] Sessions over files
- [ ] Scope/IR via LLM
- [ ] Plan executed by Python
- [ ] Generated APIs saved/executed

### Unified
- [ ] Same 8-step workflow for DB and files
- [ ] No heuristic data; outputs from models or real execution only

## Notes
- Go executes SQL only on registered DBs.
- Python handles file reads and execution for ad‑hoc file analysis.
- No heuristic SQL fallback; errors bubble up to client with guidance.

Last Updated: 2025-09-30
Status: Planning Phase
Priority: Implement Phase 0–2 first
