#!/bin/bash

# AIR-Py FastAPI Data Processing Server Startup Script

echo "Starting AIR-Py FastAPI Data Processing Server..."

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo "Python 3 is required but not installed."
    exit 1
fi

# Check if virtual environment exists, create if not
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install dependencies
echo "Installing dependencies..."
pip install -r requirements.txt

# Create temp directory
mkdir -p /tmp/air-py

# Start the server
echo "Starting FastAPI server on port 9001..."
python -m app.main
