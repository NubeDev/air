"""FastAPI endpoints for data processing."""

from fastapi import APIRouter, HTTPException, Depends, BackgroundTasks
from typing import List, Optional
import asyncio
from datetime import datetime, timezone

from app.models.schemas import (
    HealthResponse, DiscoverRequest, DiscoverResponse,
    InferSchemaRequest, InferSchemaResponse, PreviewRequest, PreviewResponse,
    QueryRequest, QueryResponse, AnalyzeRequest, AnalyzeResponse,
    JobStatusResponse, JobStatus
)
from app.services.job_manager import job_manager
from app.services.data_processor import data_processor
from app.core.config import settings

router = APIRouter()


@router.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint."""
    import polars as pl
    import pandas as pd
    import pyarrow as pa
    
    return HealthResponse(
        status="ok",
        versions={
            "polars": pl.__version__,
            "pandas": pd.__version__,
            "pyarrow": pa.__version__
        }
    )


@router.post("/discover", response_model=DiscoverResponse)
async def discover_files(request: DiscoverRequest, background_tasks: BackgroundTasks):
    """Discover files in a directory."""
    try:
        token = job_manager.create_job("discover", request.dict())
        
        # Start background task
        background_tasks.add_task(
            _process_discover,
            token,
            request.datasource_id,
            request.uri,
            request.recurse,
            request.max_files
        )
        
        return DiscoverResponse(token=token)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/infer_schema", response_model=InferSchemaResponse)
async def infer_schema(request: InferSchemaRequest, background_tasks: BackgroundTasks):
    """Infer schema from files."""
    try:
        token = job_manager.create_job("infer_schema", request.dict())
        
        # Start background task
        background_tasks.add_task(
            _process_infer_schema,
            token,
            request.datasource_id,
            request.uri,
            request.sample_files,
            request.infer_rows
        )
        
        return InferSchemaResponse(token=token)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/preview", response_model=PreviewResponse)
async def preview_data(request: PreviewRequest, background_tasks: BackgroundTasks):
    """Preview data from files."""
    try:
        token = job_manager.create_job("preview", request.dict())
        
        # Start background task
        background_tasks.add_task(
            _process_preview,
            token,
            request.datasource_id,
            request.path,
            request.limit
        )
        
        return PreviewResponse(token=token)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/query", response_model=QueryResponse)
async def query_data(request: QueryRequest, background_tasks: BackgroundTasks):
    """Execute query on files."""
    try:
        token = job_manager.create_job("query", request.dict())
        
        # Start background task
        background_tasks.add_task(
            _process_query,
            token,
            request.datasource_id,
            request.plan.dict(),
            request.output
        )
        
        return QueryResponse(token=token)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/analyze", response_model=AnalyzeResponse)
async def analyze_data(request: AnalyzeRequest, background_tasks: BackgroundTasks):
    """Analyze data for EDA, profiling, validation."""
    try:
        token = job_manager.create_job("analyze", request.dict())
        
        # Start background task
        background_tasks.add_task(
            _process_analyze,
            token,
            request.frame_ref,
            request.datasource_id,
            request.plan.dict() if request.plan else None,
            request.job_kind,
            request.options
        )
        
        return AnalyzeResponse(token=token)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/jobs/{token}", response_model=JobStatusResponse)
async def get_job_status(token: int):
    """Get job status and progress."""
    job_data = job_manager.get_job(token)
    if not job_data:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return JobStatusResponse(
        token=job_data["token"],
        status=JobStatus(job_data["status"]),
        steps=job_data["steps"],
        data=job_data["data"],
        code=200 if job_data["status"] != "failed" else 500,
        error=job_data["error"]
    )


@router.get("/jobs", response_model=List[JobStatusResponse])
async def list_jobs():
    """List all jobs."""
    jobs = job_manager.list_jobs()
    return [
        JobStatusResponse(
            token=job["token"],
            status=JobStatus(job["status"]),
            steps=job["steps"],
            data=job["data"],
            code=200 if job["status"] != "failed" else 500,
            error=job["error"]
        )
        for job in jobs
    ]


@router.delete("/jobs/{token}")
async def cancel_job(token: int):
    """Cancel a running job."""
    success = job_manager.cancel_job(token)
    if not success:
        raise HTTPException(status_code=400, detail="Job cannot be cancelled")
    return {"message": "Job cancelled successfully"}


# Background task functions

async def _process_discover(token: int, datasource_id: str, uri: str, 
                          recurse: bool, max_files: Optional[int]):
    """Process file discovery in background."""
    try:
        job_manager.update_job_status(token, "running", "Starting file discovery")
        
        files = data_processor.discover_files(datasource_id, uri, recurse, max_files)
        
        job_manager.update_job_status(
            token, "completed", 
            f"Found {len(files)} files",
            data={"files": files}
        )
    except Exception as e:
        job_manager.update_job_status(token, "failed", error=str(e))


async def _process_infer_schema(token: int, datasource_id: str, uri: str,
                              sample_files: Optional[int], infer_rows: int):
    """Process schema inference in background."""
    try:
        job_manager.update_job_status(token, "running", "Starting schema inference")
        
        schema_data = data_processor.infer_schema(datasource_id, uri, sample_files, infer_rows)
        
        job_manager.update_job_status(
            token, "completed",
            "Schema inference completed",
            data=schema_data
        )
    except Exception as e:
        job_manager.update_job_status(token, "failed", error=str(e))


async def _process_preview(token: int, datasource_id: str, path: Optional[str], limit: int):
    """Process data preview in background."""
    try:
        job_manager.update_job_status(token, "running", "Loading data preview")
        
        preview_data = data_processor.preview_data(datasource_id, path, limit)
        
        job_manager.update_job_status(
            token, "completed",
            f"Preview loaded: {len(preview_data['rows'])} rows",
            data=preview_data
        )
    except Exception as e:
        job_manager.update_job_status(token, "failed", error=str(e))


async def _process_query(token: int, datasource_id: str, plan: dict, output_format: str):
    """Process query execution in background."""
    try:
        job_manager.update_job_status(token, "running", "Executing query")
        
        query_result = data_processor.execute_query(datasource_id, plan, output_format)
        
        job_manager.update_job_status(
            token, "completed",
            f"Query completed: {query_result['metrics']['rows']} rows",
            data=query_result
        )
    except Exception as e:
        job_manager.update_job_status(token, "failed", error=str(e))


async def _process_analyze(token: int, frame_ref: Optional[dict], datasource_id: Optional[str],
                          plan: Optional[dict], job_kind: str, options: Optional[dict]):
    """Process data analysis in background."""
    try:
        job_manager.update_job_status(token, "running", f"Starting {job_kind} analysis")
        
        analysis_result = data_processor.analyze_data(
            frame_ref, datasource_id, plan, job_kind, options
        )
        
        job_manager.update_job_status(
            token, "completed",
            f"{job_kind.title()} analysis completed",
            data=analysis_result
        )
    except Exception as e:
        job_manager.update_job_status(token, "failed", error=str(e))
