"""FastAPI application entry point."""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import redis
from app.core.config import settings
from app.api.endpoints import router


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events."""
    # Startup
    print(f"Starting AIR-Py FastAPI server on {settings.host}:{settings.port}")
    
    # Test Redis connection
    try:
        redis_client = redis.from_url(settings.redis_url)
        redis_client.ping()
        print("Redis connection successful")
    except Exception as e:
        print(f"Redis connection failed: {e}")
        print("Continuing without Redis (job management will be limited)")
        # Don't raise - allow server to start without Redis
    
    yield
    
    # Shutdown
    print("Shutting down AIR-Py FastAPI server")


# Create FastAPI app
app = FastAPI(
    title="AIR-Py Data Processing Service",
    description="FastAPI microservice for data processing and analytics",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routes
app.include_router(router, prefix="/v1/py")


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "AIR-Py Data Processing Service",
        "version": "1.0.0",
        "status": "running"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug
    )
