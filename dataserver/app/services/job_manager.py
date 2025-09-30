"""Job management service for async processing."""

import redis
import json
from typing import Dict, Any, Optional, List
from datetime import datetime, timezone
from app.core.config import settings


class JobManager:
    """Manages async job state and progress."""
    
    def __init__(self):
        self.redis_client = redis.from_url(settings.redis_url)
        self.job_counter = 0
    
    def create_job(self, job_type: str, params: Dict[str, Any]) -> int:
        """Create a new job and return token."""
        self.job_counter += 1
        token = self.job_counter
        
        job_data = {
            "token": token,
            "type": job_type,
            "status": "pending",
            "created_at": datetime.now(timezone.utc).isoformat(),
            "params": params,
            "steps": [],
            "data": None,
            "error": None
        }
        
        self.redis_client.setex(
            f"job:{token}",
            3600,  # 1 hour TTL
            json.dumps(job_data)
        )
        
        return token
    
    def update_job_status(self, token: int, status: str, step: Optional[str] = None, 
                         data: Optional[Dict[str, Any]] = None, error: Optional[str] = None):
        """Update job status and add step."""
        job_key = f"job:{token}"
        job_data = self.get_job(token)
        
        if not job_data:
            return
        
        job_data["status"] = status
        
        if step:
            step_data = {
                "step": len(job_data["steps"]) + 1,
                "message": step,
                "timestamp": datetime.now(timezone.utc).isoformat()
            }
            job_data["steps"].append(step_data)
        
        if data:
            job_data["data"] = data
        
        if error:
            job_data["error"] = error
            job_data["status"] = "failed"
        
        self.redis_client.setex(job_key, 3600, json.dumps(job_data))
    
    def get_job(self, token: int) -> Optional[Dict[str, Any]]:
        """Get job data by token."""
        job_data = self.redis_client.get(f"job:{token}")
        if job_data:
            return json.loads(job_data)
        return None
    
    def list_jobs(self) -> List[Dict[str, Any]]:
        """List all jobs."""
        keys = self.redis_client.keys("job:*")
        jobs = []
        for key in keys:
            job_data = self.redis_client.get(key)
            if job_data:
                jobs.append(json.loads(job_data))
        return sorted(jobs, key=lambda x: x["created_at"], reverse=True)
    
    def cancel_job(self, token: int) -> bool:
        """Cancel a running job."""
        job_data = self.get_job(token)
        if job_data and job_data["status"] in ["pending", "running"]:
            self.update_job_status(token, "cancelled", "Job cancelled by user")
            return True
        return False


# Global job manager instance
job_manager = JobManager()
