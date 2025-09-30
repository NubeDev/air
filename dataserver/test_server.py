#!/usr/bin/env python3
"""Simple test script for AIR-Py FastAPI server."""

import requests
import json
import time

def test_health():
    """Test health endpoint."""
    try:
        response = requests.get("http://localhost:9001/v1/py/health")
        if response.status_code == 200:
            data = response.json()
            print("✅ Health check passed")
            print(f"   Status: {data['status']}")
            print(f"   Versions: {data['versions']}")
            return True
        else:
            print(f"❌ Health check failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Health check error: {e}")
        return False

def test_discover():
    """Test discover endpoint."""
    try:
        payload = {
            "datasource_id": "test",
            "uri": ".",
            "recurse": False,
            "max_files": 5
        }
        response = requests.post("http://localhost:9001/v1/py/discover", json=payload)
        if response.status_code == 200:
            data = response.json()
            print("✅ Discover test passed")
            print(f"   Token: {data['token']}")
            return data['token']
        else:
            print(f"❌ Discover test failed: {response.status_code}")
            return None
    except Exception as e:
        print(f"❌ Discover test error: {e}")
        return None

def test_job_status(token):
    """Test job status endpoint."""
    try:
        response = requests.get(f"http://localhost:9001/v1/py/jobs/{token}")
        if response.status_code == 200:
            data = response.json()
            print("✅ Job status test passed")
            print(f"   Status: {data['status']}")
            print(f"   Steps: {len(data['steps'])}")
            return True
        else:
            print(f"❌ Job status test failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Job status test error: {e}")
        return False

def main():
    """Run all tests."""
    print("Testing AIR-Py FastAPI server...")
    print("=" * 50)
    
    # Test health
    if not test_health():
        print("Server not running. Start with: make dev-data")
        return
    
    print()
    
    # Test discover
    token = test_discover()
    if token:
        print()
        
        # Wait a moment for processing
        print("Waiting for job to complete...")
        time.sleep(2)
        
        # Test job status
        test_job_status(token)
    
    print()
    print("=" * 50)
    print("Test completed!")

if __name__ == "__main__":
    main()
