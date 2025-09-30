# AIR Implementation Plan - Database + File Processing

## 🎯 **Architecture Overview**

AIR supports **dual processing paths** for both database and file-based data analysis:

### **📊 Database Processing Path**
- **Datasources**: TimescaleDB, PostgreSQL, MySQL, etc.
- **Models**: `Datasource`, `Scope`, `ScopeVersion`, `Report`, `ReportVersion`, `ReportRun`
- **Endpoints**: `/v1/datasources/*`, `/v1/reports/*`, `/v1/scopes/*`
- **Processing**: Direct SQL execution on databases
- **Status**: **Partially implemented** (datasource management works, reports/scopes are stubs)

### **📁 File Processing Path**
- **Files**: CSV, Parquet, JSONL files
- **Models**: `Session`, `GeneratedReport`, `ReportExecution`
- **Endpoints**: `/v1/sessions/*`, `/v1/generated/reports/*`
- **Processing**: Python FastAPI backend handles file processing
- **Status**: **Fully implemented** (sessions and generated reports work)

### **🤖 AI Services (Shared)**
- **AI Tools**: `/v1/ai/tools`, `/v1/ai/chat/completion`
- **SQL Generation**: `/v1/sql/generate`
- **Status**: **Fully implemented**

## 🚀 **Implementation Tasks**

### **Phase 1: Complete Database Workflow** 
**Priority: HIGH** - Database processing is core functionality

#### **1.1 Implement Reports Service** (`internal/services/reports_service.go`)
- [ ] `CreateScope(req CreateScopeRequest) (*Scope, error)`
- [ ] `GetScope(id uint) (*Scope, error)`
- [ ] `CreateScopeVersion(scopeID uint, req CreateScopeVersionRequest) (*ScopeVersion, error)`
- [ ] `CreateReport(req CreateReportRequest) (*Report, error)`
- [ ] `GetReport(key string) (*Report, error)`
- [ ] `CreateReportVersion(reportKey string, req CreateReportVersionRequest) (*ReportVersion, error)`
- [ ] `RunReport(reportKey string, req RunReportRequest) (*ReportRun, error)`
- [ ] `ExportReport(reportKey string, format string) ([]byte, error)`

#### **1.2 Implement AI Service Database Methods** (`internal/services/ai_service.go`)
- [ ] `BuildIR(req BuildIRRequest) (map[string]interface{}, error)` - Convert scope to IR
- [ ] `GenerateSQLFromIR(req GenerateSQLRequest) (string, map[string]interface{}, error)` - Generate SQL from IR
- [ ] `AnalyzeRun(req AnalyzeRunRequest) (*ReportAnalysis, error)` - AI analysis of results

#### **1.3 Implement Datasource Service Learning** (`internal/services/datasource_service.go`)
- [ ] `LearnDatasource(req LearnDatasourceRequest) error` - Learn schema from database
- [ ] `GetSchema(datasourceID string) ([]SchemaNote, error)` - Get learned schema

### **Phase 2: Complete File Learning Workflow**
**Priority: HIGH** - Interactive learning is core to the file processing experience

#### **2.1 Add Session Learning Endpoints** (`cmd/api/handlers/sessions/`)
- [ ] `POST /v1/sessions/{id}/ask` - Ask questions about data
- [ ] `POST /v1/sessions/{id}/scope/build` - Build analysis scope
- [ ] `POST /v1/sessions/{id}/scope/refine` - Refine scope with feedback
- [ ] `POST /v1/sessions/{id}/query/generate` - Generate query plan
- [ ] `POST /v1/sessions/{id}/execute` - Execute analysis
- [ ] `POST /v1/sessions/{id}/analyze` - AI analysis of results
- [ ] `POST /v1/sessions/{id}/save` - Save as reusable API

#### **2.2 Add Session Learning Service** (`internal/services/session_service.go`)
- [ ] `AskQuestion(sessionID uint, question string) (*ChatResponse, error)`
- [ ] `BuildScope(sessionID uint, req BuildScopeRequest) (*Scope, error)`
- [ ] `RefineScope(sessionID uint, scopeID uint, feedback string) (*Scope, error)`
- [ ] `GenerateQuery(sessionID uint, scopeID uint) (*QueryPlan, error)`
- [ ] `ExecuteAnalysis(sessionID uint, queryID uint, params map[string]interface{}) (*ExecutionResult, error)`
- [ ] `AnalyzeResults(sessionID uint, runID uint) (*AnalysisResult, error)`
- [ ] `SaveAsAPI(sessionID uint, req SaveAsAPIRequest) (*GeneratedReport, error)`

#### **2.3 Add Python FastAPI Integration** (`internal/services/fastapi_client.go`)
- [ ] `LearnFile(sessionID string, filePath string, options FileLearnOptions) (*FileLearnResponse, error)`
- [ ] `ExecuteFileQuery(sessionID string, queryPlan QueryPlan, params map[string]interface{}) (*ExecutionResult, error)`
- [ ] `AnalyzeFileResults(sessionID string, results ExecutionResult) (*AnalysisResult, error)`

### **Phase 3: Unified Learning Pattern**
**Priority: MEDIUM** - Both paths should follow the same workflow

#### **3.1 Create Unified Learning Interface**
- [ ] `internal/services/learning_service.go` - Unified interface for both DB and file learning
- [ ] `internal/services/scope_builder.go` - Unified scope building logic
- [ ] `internal/services/query_generator.go` - Unified query generation logic

#### **3.2 Add Learning Workflow Endpoints**
- [ ] `POST /v1/learn/{datasource_id}` - Start learning session for database
- [ ] `POST /v1/learn/file` - Start learning session for file
- [ ] `POST /v1/learn/{session_id}/ask` - Ask questions (unified)
- [ ] `POST /v1/learn/{session_id}/scope` - Build scope (unified)
- [ ] `POST /v1/learn/{session_id}/execute` - Execute analysis (unified)

### **Phase 4: UI Integration**
**Priority: MEDIUM** - Dynamic form generation for generated reports

#### **4.1 Add Schema Endpoints**
- [ ] `GET /v1/generated/reports/{id}/schema` - Get JSON Schema for form generation
- [ ] `GET /v1/reports/{id}/schema` - Get JSON Schema for database reports

#### **4.2 Add Form Generation Support**
- [ ] Convert OpenAPI parameter schemas to JSON Schema format
- [ ] Add form validation based on parameter schemas
- [ ] Add example data and descriptions for form fields

## 📁 **File Structure (Keep Current Layout)**

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

## 🎯 **Success Criteria**

### **Database Processing Complete When:**
- [ ] User can register database datasources
- [ ] User can learn schema from databases
- [ ] User can create scopes and reports
- [ ] User can generate and execute SQL queries
- [ ] User can save analyses as reusable APIs
- [ ] User can execute saved APIs with parameters

### **File Processing Complete When:**
- [ ] User can start learning sessions with files
- [ ] User can ask questions about file data
- [ ] User can build analysis scopes interactively
- [ ] User can generate and execute file queries
- [ ] User can save analyses as reusable APIs
- [ ] User can execute saved APIs with parameters

### **Unified Experience When:**
- [ ] Both database and file processing follow the same 8-step workflow
- [ ] UI can dynamically generate forms for any generated API
- [ ] AI provides consistent analysis across both data sources
- [ ] Users can seamlessly switch between database and file analysis

## 🚨 **Important Notes**

1. **Keep Current Go Layout** - Don't change the directory structure
2. **Don't Change Python Stack** - The FastAPI backend is working well
3. **Maintain Dual Support** - Both database and file processing must work
4. **Preserve Existing APIs** - Don't break current functionality
5. **Follow Specs** - Implement according to SPEC.md and SPEC-FILE-AI.md

## 📋 **Next Steps**

1. **Start with Phase 1** - Complete database workflow first
2. **Then Phase 2** - Add file learning workflow
3. **Finally Phase 3** - Unify the experience
4. **Test Everything** - Ensure both paths work end-to-end

---

**Last Updated**: 2025-09-30
**Status**: Planning Phase
**Priority**: Complete missing workflow endpoints for both database and file processing
