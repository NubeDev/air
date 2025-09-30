"""Configuration management for AIR-Py FastAPI service."""

from pydantic_settings import BaseSettings
from typing import List, Optional
import os


class Settings(BaseSettings):
    """Application settings."""
    
    # Server configuration
    host: str = "0.0.0.0"
    port: int = 9001
    debug: bool = False
    
    # Authentication (HMAC with Go backend)
    auth_shared_secret: str = "change-me"
    
    # Redis and Celery
    redis_url: str = "redis://localhost:6379/0"
    celery_broker_url: str = "redis://localhost:6379/0"
    celery_result_backend: str = "redis://localhost:6379/0"
    
    # Data processing limits
    max_rows_return_json: int = 50000
    infer_rows: int = 20000
    max_memory_mb: int = 2048
    max_workers: int = 2
    
    # File processing
    allowed_extensions: str = ".csv,.parquet,.jsonl"
    
    @property
    def allowed_extensions_list(self) -> List[str]:
        """Convert comma-separated string to list."""
        return [ext.strip() for ext in self.allowed_extensions.split(",")]
    temp_dir: str = "/tmp/air-py"
    
    # Go backend communication
    go_backend_url: str = "http://localhost:9000"
    
    class Config:
        env_file = ".env"
        case_sensitive = False


# Global settings instance
settings = Settings()
