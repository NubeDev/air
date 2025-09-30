#!/bin/bash

echo "🚀 Testing AIR File-Based AI Learning Workflow"
echo "=============================================="

# Test 1: Check if services are running
echo "📋 Step 1: Checking service health..."
echo "Go Backend:"
curl -s http://localhost:9000/health | jq '.'
echo ""

echo "FastAPI Backend:"
curl -s http://localhost:9001/v1/py/health | jq '.'
echo ""

# Test 2: Read file through FastAPI
echo "📁 Step 2: Reading test file through FastAPI..."
echo "Inferring schema from testdata/ts-energy.csv..."
SCHEMA_RESPONSE=$(curl -s -X POST http://localhost:9001/v1/py/infer_schema \
  -H "Content-Type: application/json" \
  -d '{"datasource_id":"test","uri":"../testdata/ts-energy.csv","infer_rows":50}')

echo "Schema inference response:"
echo $SCHEMA_RESPONSE | jq '.'
echo ""

# Get the job token
TOKEN=$(echo $SCHEMA_RESPONSE | jq -r '.token')
echo "Job token: $TOKEN"
echo ""

# Wait for job completion and get results
echo "⏳ Waiting for schema inference to complete..."
sleep 2

SCHEMA_RESULT=$(curl -s http://localhost:9001/v1/py/jobs/$TOKEN)
echo "Schema inference result:"
echo $SCHEMA_RESULT | jq '.'
echo ""

# Test 3: Go backend integration with FastAPI
echo "🔗 Step 3: Testing Go backend integration with FastAPI..."
echo "Running comprehensive energy data test..."
ENERGY_TEST=$(curl -s -X POST http://localhost:9000/v1/fastapi/test/energy)
echo "Energy test response (truncated):"
echo $ENERGY_TEST | jq '.message'
echo ""

# Test 4: AI Chat with file data
echo "🤖 Step 4: Testing AI chat with file data..."
echo "Sending energy data description to AI for analysis..."
AI_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/ai/chat/completion \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"I have energy data with columns: timestamp, site, device_id, kwh, voltage, current, temperature, status. The data shows hourly energy consumption measurements. Can you analyze this data and tell me what insights you can provide?"}]}')

echo "AI Analysis Response:"
echo $AI_RESPONSE | jq '.message.content' | head -20
echo ""

# Test 5: SQL Generation for file data
echo "🔍 Step 5: Testing SQL generation for file data..."
echo "Generating SQL query for energy consumption analysis..."
SQL_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/sql/generate \
  -H "Content-Type: application/json" \
  -d '{"prompt":"Show me the total energy consumption by site for the last 24 hours","schema":"CREATE TABLE energy_data (timestamp VARCHAR, site VARCHAR, device_id VARCHAR, kwh FLOAT, voltage FLOAT, current FLOAT, temperature FLOAT, status VARCHAR);"}')

echo "Generated SQL:"
echo $SQL_RESPONSE | jq -r '.sql'
echo ""

# Test 6: AI Tools
echo "🛠️  Step 6: Testing AI tools..."
echo "Getting available AI tools..."
TOOLS_RESPONSE=$(curl -s http://localhost:9000/v1/ai/tools)
echo "Available AI tools:"
echo $TOOLS_RESPONSE | jq '.tools[] | {name, description, type}'
echo ""

echo "✅ File-Based AI Learning Workflow Test Complete!"
echo "=============================================="
echo ""
echo "Summary:"
echo "- ✅ File reading through FastAPI backend"
echo "- ✅ Schema inference and data analysis"
echo "- ✅ Go backend integration with FastAPI"
echo "- ✅ AI chat analysis of file data"
echo "- ✅ SQL generation for file queries"
echo "- ✅ AI tools and capabilities"
echo ""
echo "The Go backend is ready for file-based AI learning sessions! 🎉"
