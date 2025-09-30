"""Pydantic models for API request/response validation."""

from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any, Union
from enum import Enum


class JobStatus(str, Enum):
    """Job status enumeration."""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


class JobStep(BaseModel):
    """Individual step in job execution."""
    step: int
    message: str
    timestamp: str
    duration_ms: Optional[int] = None


class JobStatusResponse(BaseModel):
    """Job status response."""
    token: int
    status: JobStatus
    steps: List[JobStep]
    data: Optional[Dict[str, Any]] = None
    code: int = 200
    error: Optional[str] = None


class HealthResponse(BaseModel):
    """Health check response."""
    status: str = "ok"
    versions: Dict[str, str]


class DiscoverRequest(BaseModel):
    """Discover files request."""
    datasource_id: str
    uri: str
    recurse: bool = True
    max_files: Optional[int] = None


class DiscoverResponse(BaseModel):
    """Discover files response."""
    token: int


class InferSchemaRequest(BaseModel):
    """Infer schema request."""
    datasource_id: str
    uri: str
    sample_files: Optional[int] = None
    infer_rows: int = 20000


class InferSchemaResponse(BaseModel):
    """Infer schema response."""
    token: int


class PreviewRequest(BaseModel):
    """Preview data request."""
    datasource_id: str
    path: Optional[str] = None
    limit: int = 100


class PreviewResponse(BaseModel):
    """Preview data response."""
    token: int


class QueryPlan(BaseModel):
    """Query execution plan."""
    dataset: str
    filters: Optional[List[Dict[str, Any]]] = None
    select: Optional[List[str]] = None
    groupby: Optional[List[str]] = None
    aggs: Optional[List[Dict[str, str]]] = None
    grain: Optional[str] = None
    order: Optional[List[Dict[str, str]]] = None
    limit: int = 50000


class QueryRequest(BaseModel):
    """Query data request."""
    datasource_id: str
    plan: QueryPlan
    output: str = "arrow"  # "arrow" | "json"


class QueryResponse(BaseModel):
    """Query data response."""
    token: int


class AnalyzeRequest(BaseModel):
    """Analyze data request."""
    frame_ref: Optional[Dict[str, Any]] = None
    datasource_id: Optional[str] = None
    plan: Optional[QueryPlan] = None
    job_kind: str  # "eda" | "profile" | "validate" | "transform"
    options: Optional[Dict[str, Any]] = None


class AnalyzeResponse(BaseModel):
    """Analyze data response."""
    token: int


class ErrorResponse(BaseModel):
    """Error response."""
    error: Dict[str, Any]
