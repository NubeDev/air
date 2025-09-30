"""Data processing service for file operations."""

import polars as pl
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq
from pathlib import Path
from typing import Dict, Any, List, Optional, Union
import json
import os
from app.core.config import settings


class DataProcessor:
    """Handles data processing operations."""
    
    def __init__(self):
        self.temp_dir = Path(settings.temp_dir)
        self.temp_dir.mkdir(exist_ok=True)
    
    def discover_files(self, datasource_id: str, uri: str, recurse: bool = True, 
                      max_files: Optional[int] = None) -> List[Dict[str, Any]]:
        """Discover files in the given URI."""
        files = []
        path = Path(uri)
        
        if not path.exists():
            raise FileNotFoundError(f"Path not found: {uri}")
        
        if path.is_file():
            if self._is_supported_file(path):
                files.append(self._get_file_info(path))
        else:
            pattern = "**/*" if recurse else "*"
            for file_path in path.glob(pattern):
                if file_path.is_file() and self._is_supported_file(file_path):
                    files.append(self._get_file_info(file_path))
                    if max_files and len(files) >= max_files:
                        break
        
        return files
    
    def infer_schema(self, datasource_id: str, uri: str, sample_files: Optional[int] = None,
                    infer_rows: int = 20000) -> Dict[str, Any]:
        """Infer schema from files."""
        files = self.discover_files(datasource_id, uri, max_files=sample_files)
        
        if not files:
            raise ValueError("No supported files found")
        
        # Sample first file for schema inference
        sample_file = files[0]["path"]
        df = self._load_file(sample_file, n_rows=infer_rows)
        
        schema = {
            "fields": [
                {
                    "name": col,
                    "type": str(df[col].dtype),
                    "nullable": df[col].null_count() > 0
                }
                for col in df.columns
            ]
        }
        
        stats = {
            "rows": df.height,
            "columns": df.width,
            "null_ratio": {col: df[col].null_count() / df.height if df.height > 0 else 0 for col in df.columns}
        }
        
        return {
            "schema": schema,
            "stats": stats
        }
    
    def preview_data(self, datasource_id: str, path: Optional[str] = None, 
                    limit: int = 100) -> Dict[str, Any]:
        """Preview data from files."""
        if not path:
            files = self.discover_files(datasource_id, ".", max_files=1)
            if not files:
                raise ValueError("No files found")
            path = files[0]["path"]
        
        df = self._load_file(path, n_rows=limit)
        
        return {
            "rows": df.to_dicts()[:limit],
            "schema": {
                "fields": [
                    {"name": col, "type": str(df[col].dtype)}
                    for col in df.columns
                ]
            },
            "stats": {"sampled": True, "total_rows": df.height}
        }
    
    def execute_query(self, datasource_id: str, plan: Dict[str, Any], 
                     output_format: str = "arrow") -> Dict[str, Any]:
        """Execute query plan on files."""
        # This is a simplified implementation
        # In production, you'd implement full query execution logic
        
        files = self.discover_files(datasource_id, plan["dataset"])
        if not files:
            raise ValueError("No files found for dataset")
        
        # Load and process first file as example
        df = self._load_file(files[0]["path"])
        
        # Apply filters if specified
        if plan.get("filters"):
            for filter_cond in plan["filters"]:
                col = filter_cond["col"]
                op = filter_cond["op"]
                val = filter_cond["val"]
                
                if op == ">=":
                    df = df.filter(pl.col(col) >= val)
                elif op == "<=":
                    df = df.filter(pl.col(col) <= val)
                elif op == "==":
                    df = df.filter(pl.col(col) == val)
                # Add more filter operations as needed
        
        # Apply select if specified
        if plan.get("select"):
            df = df.select(plan["select"])
        
        # Apply groupby and aggregations if specified
        if plan.get("groupby") and plan.get("aggs"):
            group_cols = plan["groupby"]
            agg_exprs = []
            for agg in plan["aggs"]:
                col = agg["col"]
                fn = agg["fn"]
                if fn == "sum":
                    agg_exprs.append(pl.col(col).sum().alias(f"sum_{col}"))
                elif fn == "count":
                    agg_exprs.append(pl.col(col).count().alias(f"count_{col}"))
                # Add more aggregation functions as needed
            
            df = df.group_by(group_cols).agg(agg_exprs)
        
        # Apply limit
        if plan.get("limit"):
            df = df.limit(plan["limit"])
        
        if output_format == "arrow":
            # Convert to Arrow format
            arrow_table = df.to_arrow()
            return {
                "format": "arrow",
                "data": arrow_table,
                "metrics": {
                    "rows": df.height,
                    "columns": df.width
                }
            }
        else:
            # Return as JSON
            return {
                "format": "json",
                "rows": df.to_dicts(),
                "schema": {
                    "fields": [
                        {"name": col, "type": str(df[col].dtype)}
                        for col in df.columns
                    ]
                },
                "metrics": {
                    "rows": df.height,
                    "columns": df.width
                }
            }
    
    def analyze_data(self, frame_ref: Optional[Dict[str, Any]] = None,
                    datasource_id: Optional[str] = None, plan: Optional[Dict[str, Any]] = None,
                    job_kind: str = "eda", options: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Analyze data for EDA, profiling, validation, etc."""
        # Load data based on input
        if frame_ref:
            # Load from frame reference
            df = self._load_from_frame_ref(frame_ref)
        elif datasource_id and plan:
            # Execute query plan
            result = self.execute_query(datasource_id, plan, "json")
            df = pl.DataFrame(result["rows"])
        else:
            raise ValueError("Either frame_ref or datasource_id+plan must be provided")
        
        analysis = {
            "metrics": {
                "rows": df.height,
                "columns": df.width,
                "memory_usage": df.estimated_size()
            },
            "issues": [],
            "chart_hints": [],
            "sample": df.head(50).to_dicts(),
            "schema": {
                "fields": [
                    {"name": col, "type": str(df[col].dtype)}
                    for col in df.columns
                ]
            }
        }
        
        if job_kind in ["eda", "profile"]:
            # Add correlation analysis
            numeric_cols = df.select(pl.col(pl.NUMERIC_DTYPES)).columns
            if len(numeric_cols) > 1:
                corr_matrix = df.select(numeric_cols).corr()
                analysis["metrics"]["correlations"] = corr_matrix.to_dicts()
        
        if job_kind in ["validate", "profile"]:
            # Add data quality checks
            for col in df.columns:
                null_count = df[col].null_count()
                if null_count > 0:
                    analysis["issues"].append({
                        "code": "MISSING_VALUES",
                        "detail": f"Column {col} has {null_count} missing values"
                    })
        
        return analysis
    
    def _is_supported_file(self, file_path: Path) -> bool:
        """Check if file has supported extension."""
        return file_path.suffix.lower() in settings.allowed_extensions_list
    
    def _get_file_info(self, file_path: Path) -> Dict[str, Any]:
        """Get file information."""
        stat = file_path.stat()
        return {
            "path": str(file_path),
            "name": file_path.name,
            "size": stat.st_size,
            "modified": stat.st_mtime,
            "extension": file_path.suffix
        }
    
    def _load_file(self, file_path: str, n_rows: Optional[int] = None) -> pl.DataFrame:
        """Load file into Polars DataFrame."""
        path = Path(file_path)
        
        if path.suffix.lower() == ".csv":
            return pl.read_csv(file_path, n_rows=n_rows)
        elif path.suffix.lower() == ".parquet":
            return pl.read_parquet(file_path, n_rows=n_rows)
        elif path.suffix.lower() == ".jsonl":
            return pl.read_ndjson(file_path, n_rows=n_rows)
        else:
            raise ValueError(f"Unsupported file format: {path.suffix}")
    
    def _load_from_frame_ref(self, frame_ref: Dict[str, Any]) -> pl.DataFrame:
        """Load DataFrame from frame reference."""
        # This would implement loading from Arrow, Parquet, or JSON references
        # For now, return empty DataFrame
        return pl.DataFrame()


# Global data processor instance
data_processor = DataProcessor()
