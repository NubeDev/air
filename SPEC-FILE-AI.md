# AIR File-Based AI Learning Sessions - Specification

## Overview

This specification defines the file-based AI learning session workflow for AIR, which provides a simplified, interactive way for users to understand and test the full AIR system using file datasets. The Python FastAPI backend handles all file processing, while the Go backend manages the learning sessions and AI interactions.

## Core Workflow

**File Upload** → **AI Learning** → **Interactive Q&A** → **Scope Building** → **Query Generation** → **Execution** → **API Generation**

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Go Backend    │    │  Python FastAPI │    │   File Storage  │
│   (Sessions)    │◄──►│   (Processing)  │◄──►│   (CSV/Parquet) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   AI Services   │    │   Data Analysis │
│ (Llama3/SQLCoder)│    │   (Pandas/Arrow)│
└─────────────────┘    └─────────────────┘
```

## Python FastAPI Backend Responsibilities

The Python FastAPI backend (`dataserver/`) handles **ALL** file-related operations:

### File Processing
- **File Reading**: CSV, Parquet, JSONL file parsing and validation
- **Schema Inference**: Automatic detection of data types, relationships, and patterns
- **Data Profiling**: Statistical analysis, data quality assessment, and insights generation
- **Query Execution**: Process file-based queries using Pandas/Arrow for performance
- **Data Sampling**: Extract representative samples for analysis and preview
- **Format Conversion**: Convert between different file formats as needed

### Data Analysis Services
- **Exploratory Data Analysis (EDA)**: Generate comprehensive data insights
- **Statistical Analysis**: Correlations, distributions, outliers, trends
- **Time Series Analysis**: Handle temporal data patterns and seasonality
- **Data Quality Assessment**: Missing values, duplicates, consistency checks
- **Visualization Data**: Prepare data for charts and graphs

### File Management
- **File Validation**: Ensure files are readable and properly formatted
- **Metadata Extraction**: File size, row count, column information
- **Caching**: Cache processed data for performance
- **Error Handling**: Graceful handling of malformed or corrupted files

## Go Backend Responsibilities

The Go backend manages the learning session lifecycle and AI interactions:

### Session Management
- **Session Creation**: Initialize learning sessions with file metadata
- **State Management**: Track session progress and status
- **User Interactions**: Handle natural language questions and feedback
- **AI Coordination**: Orchestrate AI services for learning and analysis

### AI Integration
- **Learning Analysis**: Use AI to understand file structure and content
- **Scope Building**: Convert user questions into structured analysis plans
- **Query Generation**: Use SQLCoder to create file processing queries
- **Result Analysis**: AI-powered validation and insights on results

## Learning Session Workflow

### 1. Session Initialization

```bash
# Start a new learning session
POST /v1/sessions/start
{
  "file_path": "/data/energy.csv",
  "session_name": "Energy Analysis Session",
  "datasource_type": "file",
  "options": {
    "infer_rows": 10000,
    "deep_analysis": true,
    "generate_insights": true
  }
}

# Response
{
  "session_id": "sess_123",
  "status": "initializing",
  "file_info": {
    "path": "/data/energy.csv",
    "size": "2.5MB",
    "estimated_rows": 50000
  },
  "estimated_duration": "30s"
}
```

### 2. AI Learning Phase

The Python backend performs comprehensive file analysis:

```python
# Python FastAPI endpoints called by Go backend
POST /v1/py/learn/file
{
  "file_path": "/data/energy.csv",
  "session_id": "sess_123",
  "options": {
    "infer_rows": 10000,
    "deep_analysis": true,
    "generate_insights": true
  }
}

# Response
{
  "schema": {
    "columns": [
      {"name": "timestamp", "type": "datetime", "nullable": false},
      {"name": "site_id", "type": "string", "nullable": false},
      {"name": "kwh", "type": "float", "nullable": false}
    ],
    "primary_keys": ["timestamp", "site_id"],
    "relationships": [],
    "time_columns": ["timestamp"]
  },
  "statistics": {
    "row_count": 50000,
    "memory_usage": "2.1MB",
    "data_quality": {
      "missing_values": 0,
      "duplicates": 0,
      "outliers": 15
    }
  },
  "insights": [
    "Time series data with hourly measurements",
    "Energy consumption ranges from 0.5 to 15.2 kWh",
    "Data spans 6 months with consistent hourly intervals",
    "3 distinct sites with varying consumption patterns"
  ],
  "sample_data": [
    {"timestamp": "2024-01-01T00:00:00Z", "site_id": "A", "kwh": 12.5},
    {"timestamp": "2024-01-01T01:00:00Z", "site_id": "A", "kwh": 11.8}
  ]
}
```

### 3. Interactive Q&A Phase

```bash
# Ask questions about the data
POST /v1/sessions/{session_id}/ask
{
  "question": "What's the energy consumption trend by month?",
  "context": "I want to see monthly trends with line charts",
  "session_id": "sess_123"
}

# AI Response
{
  "understanding": "You want to analyze monthly energy consumption trends across all sites",
  "suggested_analysis": {
    "aggregation": "monthly",
    "metrics": ["total_kwh", "avg_kwh"],
    "dimensions": ["month", "site_id"],
    "visualization": "line_chart"
  },
  "follow_up_questions": [
    "Do you want to see trends for all sites or specific ones?",
    "Should we include year-over-year comparisons?",
    "Any specific time range you're interested in?"
  ]
}
```

### 4. Scope Building

```bash
# Build analysis scope based on AI understanding
POST /v1/sessions/{session_id}/scope/build
{
  "question": "Show monthly energy consumption trends by site",
  "requirements": {
    "time_range": "last_6_months",
    "sites": "all",
    "visualization": "line_chart",
    "comparison": "year_over_year"
  },
  "session_id": "sess_123"
}

# Response
{
  "scope_id": "scope_456",
  "analysis_plan": {
    "entities": ["energy_data"],
    "metrics": [
      {"name": "total_kwh", "aggregation": "sum"},
      {"name": "avg_kwh", "aggregation": "avg"}
    ],
    "dimensions": ["month", "site_id"],
    "filters": [
      {"field": "timestamp", "op": ">=", "value": "{{date_from}}"},
      {"field": "timestamp", "op": "<", "value": "{{date_to}}"}
    ],
    "grain": "1 month",
    "order": [{"field": "month", "dir": "asc"}],
    "visualization": {
      "type": "line_chart",
      "x_axis": "month",
      "y_axis": "total_kwh",
      "series": "site_id"
    }
  },
  "parameters": {
    "date_from": {"type": "string", "format": "date", "required": true},
    "date_to": {"type": "string", "format": "date", "required": true},
    "site_id": {"type": "string", "required": false}
  }
}
```

### 5. Query Generation

The Go backend uses SQLCoder to generate file processing queries:

```bash
# Generate query plan for file processing
POST /v1/sessions/{session_id}/query/generate
{
  "scope_id": "scope_456",
  "session_id": "sess_123"
}

# Response
{
  "query_id": "query_789",
  "query_plan": {
    "operations": [
      {
        "type": "filter",
        "condition": "timestamp >= '{{date_from}}' AND timestamp < '{{date_to}}'"
      },
      {
        "type": "groupby",
        "columns": ["month", "site_id"],
        "aggregations": {
          "total_kwh": "sum",
          "avg_kwh": "mean"
        }
      },
      {
        "type": "sort",
        "columns": ["month", "site_id"]
      }
    ]
  },
  "estimated_rows": 18,
  "complexity": "low"
}
```

### 6. Query Execution

The Python backend executes the query against the file:

```bash
# Execute query on file data
POST /v1/py/execute/query
{
  "session_id": "sess_123",
  "query_id": "query_789",
  "parameters": {
    "date_from": "2024-01-01",
    "date_to": "2024-06-30"
  }
}

# Response
{
  "run_id": "run_101",
  "status": "completed",
  "results": {
    "data": [
      {"month": "2024-01", "site_id": "A", "total_kwh": 8920.5, "avg_kwh": 12.0},
      {"month": "2024-01", "site_id": "B", "total_kwh": 7560.2, "avg_kwh": 10.2}
    ],
    "metadata": {
      "row_count": 18,
      "execution_time": "0.15s",
      "memory_used": "2.1MB"
    }
  },
  "visualization_data": {
    "chart_type": "line_chart",
    "x_axis": "month",
    "y_axis": "total_kwh",
    "series": "site_id"
  }
}
```

### 7. AI Analysis and Validation

```bash
# Get AI analysis of results
POST /v1/sessions/{session_id}/analyze
{
  "run_id": "run_101",
  "session_id": "sess_123"
}

# Response
{
  "analysis_id": "analysis_202",
  "insights": [
    "Site A shows consistent consumption around 12 kWh/month",
    "Site B has more variability with peak consumption in March",
    "Overall trend shows 5% increase from January to June",
    "Data quality is excellent with no missing values"
  ],
  "recommendations": [
    "Consider investigating Site B's March spike",
    "Add year-over-year comparison for better context",
    "Include weather data correlation analysis"
  ],
  "data_quality": {
    "completeness": 100,
    "consistency": 95,
    "accuracy": 98
  }
}
```

### 8. API Generation

```bash
# Save analysis as reusable API
POST /v1/sessions/{session_id}/save
{
  "api_name": "energy_monthly_trends",
  "description": "Monthly energy consumption trends by site",
  "scope_id": "scope_456",
  "session_id": "sess_123"
}

# Response
{
  "api_id": "api_303",
  "endpoint": "/v1/reports/303",
  "documentation": {
    "description": "Get monthly energy consumption trends by site",
    "parameters": {
      "date_from": {
        "type": "string",
        "format": "date",
        "description": "Start date for analysis",
        "required": true,
        "example": "2024-01-01"
      },
      "date_to": {
        "type": "string", 
        "format": "date",
        "description": "End date for analysis",
        "required": true,
        "example": "2024-06-30"
      },
      "site_id": {
        "type": "string",
        "description": "Filter by specific site (optional)",
        "required": false,
        "example": "A"
      }
    },
    "response_format": {
      "data": "Array of monthly trend data",
      "metadata": "Execution metadata and statistics",
      "visualization": "Chart configuration data"
    }
  }
}
```

## API Endpoints

### Session Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/sessions/start` | Start new learning session |
| GET | `/v1/sessions/{id}` | Get session details |
| GET | `/v1/sessions/{id}/status` | Get session status |
| DELETE | `/v1/sessions/{id}` | End learning session |

### Interactive Learning

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/sessions/{id}/ask` | Ask questions about data |
| POST | `/v1/sessions/{id}/scope/build` | Build analysis scope |
| GET | `/v1/sessions/{id}/scope/{scope_id}` | Get scope details |
| POST | `/v1/sessions/{id}/scope/refine` | Refine scope with feedback |

### Query Processing

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/sessions/{id}/query/generate` | Generate query plan |
| POST | `/v1/sessions/{id}/execute` | Execute analysis |
| GET | `/v1/sessions/{id}/results/{run_id}` | Get execution results |
| POST | `/v1/sessions/{id}/analyze` | AI analysis of results |

### API Generation

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/sessions/{id}/save` | Save as reusable API |
| GET | `/v1/reports` | List generated APIs |
| GET | `/v1/reports/{id}` | Get API details |
| POST | `/v1/reports/{id}/execute` | Execute saved API |

## Python FastAPI Integration

### New Endpoints for File Learning

```python
# File learning and analysis
@app.post("/v1/py/learn/file")
async def learn_file(request: FileLearnRequest):
    """Comprehensive file analysis and schema inference"""
    
@app.post("/v1/py/execute/query") 
async def execute_file_query(request: FileQueryRequest):
    """Execute query plan against file data"""
    
@app.get("/v1/py/sessions/{session_id}/insights")
async def get_file_insights(session_id: str):
    """Get AI-generated insights about file data"""
    
@app.post("/v1/py/sessions/{session_id}/analyze")
async def analyze_file_results(request: FileAnalysisRequest):
    """Analyze query results with AI"""
```

### Data Models

```python
class FileLearnRequest(BaseModel):
    file_path: str
    session_id: str
    options: FileLearnOptions

class FileLearnOptions(BaseModel):
    infer_rows: int = 10000
    deep_analysis: bool = True
    generate_insights: bool = True

class FileQueryRequest(BaseModel):
    session_id: str
    query_id: str
    parameters: Dict[str, Any]

class FileAnalysisRequest(BaseModel):
    session_id: str
    run_id: str
    analysis_type: str = "comprehensive"
```

## Benefits

### For Users
- **🎯 Simple Learning**: Easy to understand workflow with familiar file data
- **🔄 Interactive**: Real-time feedback and refinement at each step
- **📚 Educational**: See how AI learns and builds analysis plans
- **💡 Practical**: Generate real, usable APIs from analysis

### For Developers
- **🚀 Scalable**: Same patterns work for database sources
- **🔧 Modular**: Clear separation between Go and Python responsibilities
- **📊 Powerful**: Full data analysis capabilities with file processing
- **🎨 Flexible**: Easy to extend with new file formats and analysis types

## Implementation Phases

### Phase 1: Core Session Management
- Session lifecycle and state management
- Basic file learning integration
- Simple Q&A functionality

### Phase 2: Interactive Scope Building
- AI-powered scope generation
- User feedback and refinement
- Query plan generation

### Phase 3: Query Execution
- File query processing
- Result validation and analysis
- Visualization data preparation

### Phase 4: API Generation
- Reusable API creation
- Documentation generation
- Parameter validation

### Phase 5: Advanced Features
- Multi-file sessions
- Advanced analytics
- Real-time collaboration
