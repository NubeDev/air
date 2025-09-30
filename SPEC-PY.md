# AIR-Py (FastAPI) — Data Science Microservice Specification

## Purpose

A private **data-science microservice** used **only by the Go backend** to handle **file datasets** (CSV/Parquet/JSONL) and post-query analytics (EDA, profiling, validation). Go remains the single public API (OpenAPI), auth, orchestration, storage, and WebSocket hub.

**FastAPI also uses OpenAPI** as the single source of truth for its internal API contract, ensuring consistency between Go backend and Python microservice.

## Responsibilities

* **Load & process files**: parse, clean, filter, aggregate, resample, join (pandas/polars/pyarrow)
* **Schema inference & profiling**: infer dtypes, basic stats, nulls/uniques, outliers, correlations
* **Frame analytics**: EDA/validation on result frames from DB or files
* **Return compact artifacts**: metrics JSON, small samples, and chart hints (no plotting)

## Non-Responsibilities

* No public client access; no auth for end users; no persistence of AIR metadata; no SQL execution

---

## OpenAPI Integration

### Single Source of Truth

FastAPI microservice **extends the main OpenAPI specification** (`api/openapi.yaml`) with internal-only endpoints:

- **Main OpenAPI**: Public endpoints for clients (CLI, web UI, AI tools)
- **FastAPI OpenAPI**: Internal endpoints for file processing (Go → FastAPI)
- **Generated Code**: Both Go and Python use OpenAPI-generated clients and models
- **Consistency**: Same data models, error formats, and authentication patterns

### OpenAPI Structure

The main `api/openapi.yaml` includes both public and internal endpoints:

```yaml
# api/openapi.yaml
paths:
  # Public endpoints (Go backend)
  /v1/datasources:
    get:
      # ... public API
  /v1/reports:
    get:
      # ... public API
      
  # Internal endpoints (FastAPI microservice)
  /v1/py/health:
    get:
      summary: "FastAPI Health Check"
      tags: ["internal", "python"]
      x-internal: true
      # ... internal API
  /v1/py/infer_schema:
    post:
      summary: "Infer File Schema"
      tags: ["internal", "python"]
      x-internal: true
      # ... internal API
  /v1/py/query:
    post:
      summary: "Execute File Query"
      tags: ["internal", "python"]
      x-internal: true
      # ... internal API
```

**Note**: Internal endpoints are marked with `x-internal: true` to distinguish them from public APIs.

### Code Generation

```bash
# Generate Go client for internal FastAPI calls
oapi-codegen -generate client -package pyclient -o internal/python/client.gen.go api/openapi.yaml

# Generate Python models from OpenAPI
datamodel-codegen --input api/openapi.yaml --output python/models/openapi_models.py
```

## Integration with Go Backend

### Call Pattern

* **Go → FastAPI** over REST (internal network) using generated OpenAPI client
* **Go signs requests** with a shared secret/JWT (HMAC) from `data/config.yaml`
* **Go owns jobs & WebSocket**: FastAPI returns quick results or a `job_id`; Go polls results and broadcasts progress via its WS channels

### Data Transport

* **Preferred**: **Arrow IPC** (binary stream) or **Parquet** for large frames
* **Fallback**: gzipped JSON for small previews
* **URIs**: For file datasets, Go passes a datasource id + logical dataset/path; FastAPI reads from local paths or cloud URIs (if configured)

---

## Configuration

### Go Config (`data/config.yaml`)

```yaml
# ... existing config ...

python:
  enabled: true
  base_url: "http://localhost:9001"  # FastAPI server
  timeout: "90s"
  auth_shared_secret: "change-me-in-production"
  resource_limits:
    memory_mb: 2048
    workers: 2

files:
  allowed_ext: [".csv", ".parquet", ".jsonl"]
  max_rows_return_json: 50000
  infer_rows: 20000
  base_path: "/data/files"  # Base path for file datasets
```

### FastAPI Environment Variables

FastAPI reads the same values via environment variables:
- `PY_AUTH_SECRET` → `python.auth_shared_secret`
- `PY_MAX_ROWS_JSON` → `files.max_rows_return_json`
- `PY_INFER_ROWS` → `files.infer_rows`
- `PY_BASE_PATH` → `files.base_path`
- `PY_MEMORY_MB` → `python.resource_limits.memory_mb`
- `PY_WORKERS` → `python.resource_limits.workers`
- `REDIS_URL` → `redis://localhost:6379/0` (for job queue)
- `CELERY_BROKER_URL` → `redis://localhost:6379/1` (for Celery)

---

## Async Processing with Tokens

### Token System

All data processing operations that may take time return a **token** (incremental integer starting from 1) and process asynchronously:

- **Token Generation**: Auto-incrementing integer (1, 2, 3, ...)
- **Status Tracking**: Real-time progress updates via token lookup
- **Step-by-Step Updates**: Detailed progress with timestamps and data
- **Result Storage**: Final results stored and retrievable by token

### Job Status Response Format

```json
{
  "token": 123,
  "status": "running|completed|failed",
  "steps": [
    {
      "step": "file_found",
      "message": "Found people.csv",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "processing_started", 
      "message": "Started processing: time taken 10 sec",
      "timestamp": "2025-09-30T10:00:10Z",
      "duration_ms": 10000
    },
    {
      "step": "analysis_complete",
      "message": "Completed data analysis",
      "timestamp": "2025-09-30T10:01:00Z", 
      "duration_ms": 50000
    }
  ],
  "data": {
    "result": "actual_data_here"
  },
  "code": 200,
  "error": null
}
```

## FastAPI Endpoints (Private; Called by Go)

### 1. Health Check

**`GET /v1/py/health`**

**Response:**
```json
{
  "status": "ok",
  "versions": {
    "pandas": "2.1.0",
    "polars": "0.20.0",
    "pyarrow": "13.0.0"
  },
  "memory_usage_mb": 256,
  "workers_active": 2,
  "active_jobs": 5,
  "next_token": 124
}
```

### 2. Job Status Lookup

**`GET /v1/py/jobs/{token}`**

**Response:**
```json
{
  "token": 123,
  "status": "running",
  "steps": [
    {
      "step": "file_found",
      "message": "Found people.csv",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "processing_started",
      "message": "Started processing: time taken 10 sec", 
      "timestamp": "2025-09-30T10:00:10Z",
      "duration_ms": 10000
    }
  ],
  "data": null,
  "code": 200,
  "error": null
}
```

### 3. File Discovery (Async)

**`POST /v1/py/discover`**

**Request:**
```json
{
  "datasource_id": "energy-files",
  "uri": "/data/files/energy",
  "recurse": true,
  "max_files": 1000
}
```

**Response:**
```json
{
  "token": 123,
  "status": "started",
  "message": "File discovery job started"
}
```

**Poll for results using token:**
**`GET /v1/py/jobs/123`**

**Response (in progress):**
```json
{
  "token": 123,
  "status": "running",
  "steps": [
    {
      "step": "scanning_directory",
      "message": "Scanning /data/files/energy",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "files_found",
      "message": "Found 15 files, processing...",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    }
  ],
  "data": null,
  "code": 200,
  "error": null
}
```

**Response (completed):**
```json
{
  "token": 123,
  "status": "completed",
  "steps": [
    {
      "step": "scanning_directory",
      "message": "Scanning /data/files/energy",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "files_found",
      "message": "Found 15 files, processing...",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    },
    {
      "step": "discovery_complete",
      "message": "File discovery completed",
      "timestamp": "2025-09-30T10:00:10Z",
      "duration_ms": 10000
    }
  ],
  "data": {
    "files": [
      {
        "path": "/data/files/energy/2025/01/measurements.parquet",
        "size_bytes": 1048576,
        "modified": "2025-01-15T10:30:00Z",
        "estimated_rows": 50000
      }
    ],
    "total_files": 1,
    "total_size_bytes": 1048576
  },
  "code": 200,
  "error": null
}
```

### 4. Schema Inference & Profiling (Async)

**`POST /v1/py/infer_schema`**

**Request:**
```json
{
  "datasource_id": "energy-files",
  "uri": "/data/files/energy/measurements.parquet",
  "sample_files": 3,
  "infer_rows": 20000
}
```

**Response:**
```json
{
  "token": 124,
  "status": "started",
  "message": "Schema inference job started"
}
```

**Poll for results using token:**
**`GET /v1/py/jobs/124`**

**Response (in progress):**
```json
{
  "token": 124,
  "status": "running",
  "steps": [
    {
      "step": "file_found",
      "message": "Found measurements.parquet",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "reading_file",
      "message": "Reading file header and sample data",
      "timestamp": "2025-09-30T10:00:02Z",
      "duration_ms": 2000
    },
    {
      "step": "analyzing_schema",
      "message": "Analyzing data types and structure",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    }
  ],
  "data": null,
  "code": 200,
  "error": null
}
```

**Response (completed):**
```json
{
  "token": 124,
  "status": "completed",
  "steps": [
    {
      "step": "file_found",
      "message": "Found measurements.parquet",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "reading_file",
      "message": "Reading file header and sample data",
      "timestamp": "2025-09-30T10:00:02Z",
      "duration_ms": 2000
    },
    {
      "step": "analyzing_schema",
      "message": "Analyzing data types and structure",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    },
    {
      "step": "schema_complete",
      "message": "Schema inference completed",
      "timestamp": "2025-09-30T10:00:08Z",
      "duration_ms": 8000
    }
  ],
  "data": {
    "schema": {
      "fields": [
        {
          "name": "timestamp",
          "type": "timestamp",
          "nullable": false
        },
        {
          "name": "site_id",
          "type": "string",
          "nullable": false
        },
        {
          "name": "kwh",
          "type": "float64",
          "nullable": true
        },
        {
          "name": "cost",
          "type": "float64",
          "nullable": true
        }
      ]
    },
    "stats": {
      "rows": 123456,
      "columns": 4,
      "null_ratio": {
        "kwh": 0.01,
        "cost": 0.02
      },
      "unique_counts": {
        "site_id": 25
      },
      "memory_usage_mb": 45.2
    },
    "sample_preview": [
      {
        "timestamp": "2025-01-01T00:00:00Z",
        "site_id": "site_001",
        "kwh": 150.5,
        "cost": 22.58
      }
    ]
  },
  "code": 200,
  "error": null
}
```

### 4. Data Preview

**`POST /v1/py/preview`**

**Request:**
```json
{
  "datasource_id": "energy-files",
  "path": "/data/files/energy/measurements.parquet",
  "limit": 100
}
```

**Response:**
```json
{
  "rows": [
    {
      "timestamp": "2025-01-01T00:00:00Z",
      "site_id": "site_001",
      "kwh": 150.5,
      "cost": 22.58
    }
  ],
  "schema": {
    "fields": [...]
  },
  "stats": {
    "sampled": true,
    "total_rows": 123456,
    "sample_size": 100
  }
}
```

### 5. Query Execution (File Processing) - Async

**`POST /v1/py/query`**

**Request:**
```json
{
  "datasource_id": "energy-files",
  "plan": {
    "dataset": "energy_parquet",
    "filters": [
      {
        "col": "timestamp",
        "op": ">=",
        "val": "2025-09-01T00:00:00Z"
      }
    ],
    "select": ["timestamp", "site_id", "kwh", "cost"],
    "groupby": ["site_id", "day"],
    "aggs": [
      {
        "col": "kwh",
        "fn": "sum"
      },
      {
        "col": "cost",
        "fn": "sum"
      }
    ],
    "grain": "1 day",
    "order": [
      {
        "col": "day",
        "dir": "asc"
      }
    ],
    "limit": 50000
  },
  "output": "arrow"
}
```

**Response:**
```json
{
  "token": 125,
  "status": "started",
  "message": "Query execution job started"
}
```

**Poll for results using token:**
**`GET /v1/py/jobs/125`**

**Response (in progress):**
```json
{
  "token": 125,
  "status": "running",
  "steps": [
    {
      "step": "file_found",
      "message": "Found energy_parquet files",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "loading_data",
      "message": "Loading data from files",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    },
    {
      "step": "applying_filters",
      "message": "Applying filters and transformations",
      "timestamp": "2025-09-30T10:00:15Z",
      "duration_ms": 15000
    },
    {
      "step": "aggregating",
      "message": "Performing aggregations",
      "timestamp": "2025-09-30T10:00:30Z",
      "duration_ms": 30000
    }
  ],
  "data": null,
  "code": 200,
  "error": null
}
```

**Response (completed):**
```json
{
  "token": 125,
  "status": "completed",
  "steps": [
    {
      "step": "file_found",
      "message": "Found energy_parquet files",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "loading_data",
      "message": "Loading data from files",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    },
    {
      "step": "applying_filters",
      "message": "Applying filters and transformations",
      "timestamp": "2025-09-30T10:00:15Z",
      "duration_ms": 15000
    },
    {
      "step": "aggregating",
      "message": "Performing aggregations",
      "timestamp": "2025-09-30T10:00:30Z",
      "duration_ms": 30000
    },
    {
      "step": "query_complete",
      "message": "Query execution completed",
      "timestamp": "2025-09-30T10:00:45Z",
      "duration_ms": 45000
    }
  ],
  "data": {
    "rows": [
      {
        "day": "2025-09-01",
        "site_id": "site_001",
        "kwh_sum": 3600.5,
        "cost_sum": 540.75
      }
    ],
    "schema": {
      "fields": [...]
    },
    "metrics": {
      "rows": 30,
      "bytes": 2048,
      "execution_time_ms": 45000
    }
  },
  "code": 200,
  "error": null
}
```

**For Arrow output, additional endpoint:**
**`GET /v1/py/jobs/{token}/download`**

- **Content-Type**: `application/vnd.apache.arrow.stream`
- **Headers**: `X-Arrow-Schema`, `X-Arrow-Rows`, `X-Arrow-Bytes`
- **Body**: Arrow IPC stream (only available when job is completed)

### 6. Frame Analysis (EDA/Validation) - Async

**`POST /v1/py/analyze`**

**Request (with frame reference):**
```json
{
  "frame_ref": {
    "type": "arrow",
    "data": "base64_encoded_arrow_data"
  },
  "job_kind": "eda",
  "options": {
    "correlation_threshold": 0.7,
    "outlier_method": "iqr"
  }
}
```

**Request (with query plan):**
```json
{
  "datasource_id": "energy-files",
  "plan": {
    "dataset": "energy_parquet",
    "filters": [...],
    "select": [...]
  },
  "job_kind": "profile",
  "options": {
    "include_correlations": true,
    "include_outliers": true
  }
}
```

**Response:**
```json
{
  "token": 126,
  "status": "started",
  "message": "Analysis job started"
}
```

**Poll for results using token:**
**`GET /v1/py/jobs/126`**

**Response (in progress):**
```json
{
  "token": 126,
  "status": "running",
  "steps": [
    {
      "step": "loading_data",
      "message": "Loading data for analysis",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "basic_stats",
      "message": "Computing basic statistics",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    },
    {
      "step": "correlation_analysis",
      "message": "Analyzing correlations",
      "timestamp": "2025-09-30T10:00:10Z",
      "duration_ms": 10000
    },
    {
      "step": "outlier_detection",
      "message": "Detecting outliers",
      "timestamp": "2025-09-30T10:00:15Z",
      "duration_ms": 15000
    }
  ],
  "data": null,
  "code": 200,
  "error": null
}
```

**Response (completed):**
```json
{
  "token": 126,
  "status": "completed",
  "steps": [
    {
      "step": "loading_data",
      "message": "Loading data for analysis",
      "timestamp": "2025-09-30T10:00:00Z",
      "duration_ms": 0
    },
    {
      "step": "basic_stats",
      "message": "Computing basic statistics",
      "timestamp": "2025-09-30T10:00:05Z",
      "duration_ms": 5000
    },
    {
      "step": "correlation_analysis",
      "message": "Analyzing correlations",
      "timestamp": "2025-09-30T10:00:10Z",
      "duration_ms": 10000
    },
    {
      "step": "outlier_detection",
      "message": "Detecting outliers",
      "timestamp": "2025-09-30T10:00:15Z",
      "duration_ms": 15000
    },
    {
      "step": "analysis_complete",
      "message": "Analysis completed",
      "timestamp": "2025-09-30T10:00:20Z",
      "duration_ms": 20000
    }
  ],
  "data": {
    "metrics": {
      "rows": 10000,
      "columns": 12,
      "memory_usage_mb": 45.2,
      "correlations": [
        ["kwh", "cost", 0.86],
        ["timestamp", "kwh", 0.12]
      ]
    },
    "issues": [
      {
        "code": "OUTLIERS",
        "severity": "warning",
        "detail": "kwh 3-sigma outliers detected on 2025-09-07",
        "affected_rows": 5
      },
      {
        "code": "HIGH_NULL_RATIO",
        "severity": "info",
        "detail": "cost column has 15% null values",
        "affected_rows": 1500
      }
    ],
    "chart_hints": [
      {
        "type": "line",
        "x": "timestamp",
        "y": "kwh",
        "title": "Energy Usage Over Time"
      },
      {
        "type": "bar",
        "x": "site_id",
        "y": "sum(cost)",
        "title": "Total Cost by Site"
      },
      {
        "type": "scatter",
        "x": "kwh",
        "y": "cost",
        "title": "Energy vs Cost Correlation"
      }
    ],
    "sample": [
      {
        "timestamp": "2025-09-01T00:00:00Z",
        "site_id": "site_001",
        "kwh": 150.5,
        "cost": 22.58
      }
    ],
    "schema": {
      "fields": [...]
    }
  },
  "code": 200,
  "error": null
}
```

---

## Job Management System

### Token Generation
- **Auto-incrementing**: Tokens start at 1 and increment for each new job
- **Persistent Storage**: Tokens stored in Redis with job metadata
- **Status Tracking**: Real-time updates as jobs progress through steps
- **Result Storage**: Final results stored and retrievable by token

### Job Lifecycle
1. **Job Creation**: POST request returns token immediately
2. **Background Processing**: Celery worker processes job asynchronously
3. **Step Updates**: Worker updates job status with detailed steps
4. **Result Storage**: Final results stored in Redis with TTL
5. **Cleanup**: Jobs cleaned up after configurable TTL (default 24h)

### Job Storage Schema
```json
{
  "token": 123,
  "status": "running|completed|failed",
  "created_at": "2025-09-30T10:00:00Z",
  "updated_at": "2025-09-30T10:01:00Z",
  "job_type": "discover|infer_schema|query|analyze",
  "steps": [...],
  "data": {...},
  "error": null,
  "ttl": 86400
}
```

### Worker Configuration
- **Celery Workers**: Configurable number of background workers
- **Memory Limits**: Per-worker memory constraints
- **Timeout**: Maximum job execution time
- **Retry Logic**: Automatic retry for transient failures
- **Dead Letter Queue**: Failed jobs moved to DLQ for inspection

---

## Security & Limits

### Authentication

* **HMAC/JWT**: All requests require `Authorization: Bearer <token>` header
* **Shared Secret**: Token signed with `python.auth_shared_secret` from config
* **Request Validation**: FastAPI validates token signature and expiration

### Resource Limits

* **Memory**: Per-request limit from `python.resource_limits.memory_mb`
* **Rows**: JSON responses capped at `files.max_rows_return_json`
* **Time**: Request timeout from `python.timeout`
* **Workers**: Pool size limited by `python.resource_limits.workers`

### Input Validation

* **File Extensions**: Only allowlisted extensions from `files.allowed_ext`
* **Path Security**: Validate file paths against `files.base_path`
* **Plan Validation**: Strict schema validation for query plans
* **No Arbitrary Code**: Only predefined operations allowed

---

## Error Model

Consistent error envelope for Go to pass upstream:

```json
{
  "error": {
    "code": "VALIDATION_ERROR|IO_ERROR|TIMEOUT|RESOURCE_CAP|AUTH_ERROR",
    "message": "Human-readable error message",
    "details": {
      "field": "specific field that failed validation",
      "constraint": "what constraint was violated"
    },
    "request_id": "req_12345"
  }
}
```

**Error Codes:**
- `VALIDATION_ERROR`: Invalid request format or parameters
- `IO_ERROR`: File system or network I/O failure
- `TIMEOUT`: Request exceeded time limit
- `RESOURCE_CAP`: Memory or row limit exceeded
- `AUTH_ERROR`: Invalid or missing authentication token

---

## How Go Backend Uses AIR-Py

### File Dataset Learning

1. **Go** calls `/v1/py/infer_schema` with file path
2. **FastAPI** analyzes file structure and returns schema + stats
3. **Go** converts schema to Markdown format
4. **Go** stores schema notes in SQLite with `datasource_id`

### File Dataset Querying

1. **Go** receives natural language query
2. **Go** converts to IR (Intermediate Representation)
3. **Go** calls `/v1/py/query` with IR converted to query plan
4. **FastAPI** processes files and returns Arrow stream or JSON
5. **Go** streams results to client via WebSocket
6. **Go** stores samples and metadata in SQLite

### Post-Query Analysis

1. **Go** executes SQL query on database
2. **Go** calls `/v1/py/analyze` with result frame
3. **FastAPI** performs EDA and returns insights
4. **Go** incorporates analysis into QA verdict
5. **Go** stores analysis results in SQLite

### Long-Running Operations

1. **Go** calls `/v1/py/discover` or large `/v1/py/query`
2. **FastAPI** returns `job_id` immediately
3. **Go** polls `/v1/py/jobs/{job_id}` for status
4. **Go** broadcasts progress via WebSocket to clients
5. **Go** handles final results when job completes

---

## Project Structure

```
python/
├─ main.py                    # FastAPI application entry point
├─ config.py                  # Configuration management
├─ requirements.txt           # Python dependencies
├─ pyproject.toml            # Poetry configuration
├─ Dockerfile                # Container configuration
├─ models/                   # Pydantic models
│  ├─ __init__.py
│  ├─ openapi_models.py      # Generated from OpenAPI spec
│  ├─ requests.py            # Request models
│  ├─ responses.py           # Response models
│  └─ jobs.py                # Job/token models
├─ services/                 # Business logic
│  ├─ __init__.py
│  ├─ file_service.py        # File processing logic
│  ├─ schema_service.py      # Schema inference
│  ├─ query_service.py       # Query execution
│  ├─ analysis_service.py    # EDA and validation
│  └─ job_service.py         # Async job management
├─ api/                      # API endpoints
│  ├─ __init__.py
│  ├─ health.py              # Health check endpoints
│  ├─ files.py               # File processing endpoints
│  ├─ schema.py              # Schema inference endpoints
│  ├─ query.py               # Query execution endpoints
│  ├─ analysis.py            # Analysis endpoints
│  └─ jobs.py                # Job status endpoints
├─ core/                     # Core functionality
│  ├─ __init__.py
│  ├─ auth.py                # Authentication middleware
│  ├─ limits.py              # Resource limit enforcement
│  ├─ storage.py             # Job storage and retrieval
│  └─ utils.py               # Utility functions
├─ workers/                  # Background workers
│  ├─ __init__.py
│  ├─ file_worker.py         # File processing worker
│  ├─ query_worker.py        # Query execution worker
│  └─ analysis_worker.py     # Analysis worker
├─ tests/                    # Test suite
│  ├─ __init__.py
│  ├─ conftest.py            # Test configuration
│  ├─ test_api/              # API tests
│  ├─ test_services/         # Service tests
│  └─ fixtures/              # Test data fixtures
└─ scripts/                  # Utility scripts
   ├─ generate_models.py     # Generate models from OpenAPI
   └─ run_tests.py           # Test runner
```

## Technology Stack

### Core Framework
- **FastAPI**: Modern Python web framework with automatic OpenAPI generation
- **Uvicorn**: ASGI server for production deployment
- **Pydantic**: Data validation and serialization
- **Celery**: Background task processing with Redis broker

### OpenAPI Integration
- **OpenAPI 3.0**: Single source of truth for API contract
- **datamodel-codegen**: Generate Python models from OpenAPI spec
- **oapi-codegen**: Generate Go client for internal FastAPI calls
- **FastAPI Auto-Docs**: Automatic OpenAPI documentation generation

### Data Processing
- **Polars**: Primary DataFrame library (fast, memory-efficient)
- **PyArrow**: Arrow format support and interoperability
- **Pandas**: Fallback for complex operations not supported by Polars

### Async Processing
- **Celery**: Distributed task queue for long-running operations
- **Redis**: Message broker and result backend
- **Token System**: Incremental job tracking with status updates

### Serialization
- **Arrow IPC**: Binary format for large datasets
- **Parquet**: Columnar storage format
- **JSON**: Human-readable format for small datasets

### Development & Testing
- **pytest**: Testing framework
- **pytest-asyncio**: Async testing support
- **httpx**: HTTP client for testing
- **Poetry**: Dependency management

---

## Deployment

### Docker Configuration

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY pyproject.toml poetry.lock ./
RUN pip install poetry && poetry install --no-dev

# Copy application code
COPY . .

# Expose port
EXPOSE 9001

# Run application
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9001"]
```

### Docker Compose Integration

```yaml
services:
  air-go:
    build: .
    ports: ["9000:9000"]
    depends_on: [air-py, redis]
    environment:
      - PYTHON_BASE_URL=http://air-py:9001
      - PYTHON_AUTH_SECRET=shared-secret

  air-py:
    build: ./python
    ports: ["9001:9001"]
    depends_on: [redis]
    volumes:
      - ./data/files:/data/files:ro
    environment:
      - PY_AUTH_SECRET=shared-secret
      - PY_BASE_PATH=/data/files
      - PY_MEMORY_MB=2048
      - PY_WORKERS=2
      - REDIS_URL=redis://redis:6379/0
      - CELERY_BROKER_URL=redis://redis:6379/1

  air-py-worker:
    build: ./python
    command: celery -A workers.celery_app worker --loglevel=info
    depends_on: [redis]
    volumes:
      - ./data/files:/data/files:ro
    environment:
      - PY_AUTH_SECRET=shared-secret
      - PY_BASE_PATH=/data/files
      - PY_MEMORY_MB=2048
      - REDIS_URL=redis://redis:6379/0
      - CELERY_BROKER_URL=redis://redis:6379/1

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

volumes:
  redis_data:
```

---

## Success Criteria

1. **Go remains the only public surface** (OpenAPI & WebSocket)
2. **Files are fully supported only via Python**; no Go file I/O
3. **Uniform outputs** (schema, metrics, sample, chart hints) regardless of DB/file path
4. **Strong guardrails** on memory/rows/time; predictable latency for previews
5. **Arrow for big data**; JSON for small previews
6. **Seamless integration** with existing Go backend workflows
7. **No authentication complexity** for end users (internal service only)
8. **Resource efficiency** with proper memory management and worker pooling
9. **Async processing** with token-based job tracking for long-running operations
10. **Real-time progress updates** with detailed step-by-step status reporting
11. **Scalable architecture** with proper separation of concerns and growth potential
12. **OpenAPI consistency** between Go backend and FastAPI microservice
