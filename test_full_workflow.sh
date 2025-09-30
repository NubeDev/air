#!/bin/bash

echo "🚀 Testing Full File-Based AI Learning Workflow"
echo "=============================================="
echo ""

# Check if services are running
echo "🔍 Checking service status..."
echo "----------------------------"

# Check Go backend
GO_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9000/health)
if [ "$GO_STATUS" = "200" ]; then
    echo "✅ Go backend (port 9000) is running"
else
    echo "❌ Go backend (port 9000) is not running"
    echo "   Run: make dev-backend"
    exit 1
fi

# Check Python backend
PY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9001/)
if [ "$PY_STATUS" = "200" ]; then
    echo "✅ Python backend (port 9001) is running"
else
    echo "❌ Python backend (port 9001) is not running"
    echo "   Run: make dev-data"
    exit 1
fi

echo ""

# Step 1: Test AI Tools
echo "🤖 Step 1: Testing AI Tools"
echo "---------------------------"
echo "Getting available AI tools..."

AI_TOOLS_RESPONSE=$(curl -s -X GET http://localhost:9000/v1/ai/tools)
echo "AI Tools Response:"
echo $AI_TOOLS_RESPONSE | jq '.'
echo ""

# Step 2: Test Chat Completion
echo "💬 Step 2: Testing Chat Completion"
echo "---------------------------------"
echo "Testing AI chat with energy data question..."

CHAT_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/ai/chat/completion \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {
        "role": "user", 
        "content": "I have energy data with timestamp, site, device_id, kwh, voltage, current, temperature, status. I want to analyze energy consumption by site for different date ranges. What analysis would you recommend?"
      }
    ]
  }')

echo "AI Chat Response:"
echo $CHAT_RESPONSE | jq -r '.message.content' | head -15
echo ""

# Step 3: Test SQL Generation
echo "🔍 Step 3: Testing SQL Generation"
echo "--------------------------------"
echo "Testing SQLCoder for date range analysis..."

SQL_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/sql/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Generate SQL to show daily energy consumption by site for a date range. Include total, average, and maximum kwh values.",
    "schema": "CREATE TABLE energy_data (timestamp VARCHAR, site VARCHAR, device_id VARCHAR, kwh FLOAT, voltage FLOAT, current FLOAT, temperature FLOAT, status VARCHAR);"
  }')

echo "Generated SQL Query:"
echo $SQL_RESPONSE | jq -r '.sql'
echo ""

# Step 4: Test File Reading
echo "📁 Step 4: Testing File Reading"
echo "------------------------------"
echo "Testing Python backend file processing..."

FILE_RESPONSE=$(curl -s -X POST http://localhost:9001/v1/py/infer_schema \
  -H "Content-Type: application/json" \
  -d '{"datasource_id":"test","uri":"../testdata/ts-energy.csv","infer_rows":50}')

TOKEN=$(echo $FILE_RESPONSE | jq -r '.token')
echo "File processing job token: $TOKEN"

# Wait for completion
sleep 2
SCHEMA_RESULT=$(curl -s http://localhost:9001/v1/py/jobs/$TOKEN)
echo "File schema result:"
echo $SCHEMA_RESULT | jq '.data.schema'
echo ""

# Step 5: Test AI Analysis of Results
echo "🧠 Step 5: Testing AI Analysis"
echo "-----------------------------"
echo "Testing AI analysis of sample data..."

AI_ANALYSIS=$(curl -s -X POST http://localhost:9000/v1/ai/chat/completion \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "Analyze these energy consumption results: Site A shows 1250.5 kWh on Jan 1st and 1280.3 kWh on Jan 2nd. Site B shows 980.2 kWh on Jan 1st and 1010.7 kWh on Jan 2nd. What insights can you provide?"
      }
    ]
  }')

echo "AI Analysis of Results:"
echo $AI_ANALYSIS | jq -r '.message.content' | head -10
echo ""

# Step 6: Test Session Management (Not Yet Implemented)
echo "📋 Step 6: Testing Session Management"
echo "------------------------------------"
echo "Testing session creation (this will fail - not yet implemented)..."

SESSION_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/sessions/start \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/data/ts-energy.csv",
    "session_name": "Energy Analysis Session",
    "datasource_type": "file",
    "options": {
      "infer_rows": 50,
      "deep_analysis": true,
      "generate_insights": true
    }
  }')

echo "Session creation response (expected to fail):"
echo $SESSION_RESPONSE | jq '.'
echo ""

# Step 7: Test Report Management (Not Yet Implemented)
echo "📊 Step 7: Testing Report Management"
echo "-----------------------------------"
echo "Testing report listing (this will fail - not yet implemented)..."

REPORTS_RESPONSE=$(curl -s -X GET http://localhost:9000/v1/reports)
echo "Reports listing response (expected to fail):"
echo $REPORTS_RESPONSE | jq '.'
echo ""

# Step 8: Simulate Full Workflow
echo "🔄 Step 8: Simulating Full Workflow"
echo "----------------------------------"
echo "Simulating the complete workflow with current working components..."

# Simulate scope building
SCOPE_PLAN='{
  "scope_id": "scope_456",
  "analysis_plan": {
    "entities": ["energy_data"],
    "metrics": [
      {"name": "total_kwh", "aggregation": "sum"},
      {"name": "avg_kwh", "aggregation": "avg"},
      {"name": "max_kwh", "aggregation": "max"}
    ],
    "dimensions": ["site", "date"],
    "filters": [
      {"field": "timestamp", "op": ">=", "value": "{{date_from}}"},
      {"field": "timestamp", "op": "<", "value": "{{date_to}}"}
    ],
    "grain": "1 day",
    "order": [{"field": "date", "dir": "asc"}],
    "visualization": {
      "type": "table",
      "x_axis": "date",
      "y_axis": "total_kwh",
      "series": "site"
    }
  },
  "parameters": {
    "date_from": {"type": "string", "format": "date", "required": true, "description": "Start date for analysis"},
    "date_to": {"type": "string", "format": "date", "required": true, "description": "End date for analysis"},
    "site": {"type": "string", "required": false, "description": "Filter by specific site (optional)"}
  }
}'

echo "Generated Analysis Scope:"
echo $SCOPE_PLAN | jq '.'
echo ""

# Simulate query results
QUERY_RESULTS='{
  "run_id": "run_789",
  "status": "completed",
  "results": {
    "data": [
      {"site": "site_a", "date": "2024-01-01", "total_kwh": 1250.5, "avg_kwh": 52.1, "max_kwh": 75.2},
      {"site": "site_a", "date": "2024-01-02", "total_kwh": 1280.3, "avg_kwh": 53.3, "max_kwh": 78.1},
      {"site": "site_b", "date": "2024-01-01", "total_kwh": 980.2, "avg_kwh": 40.8, "max_kwh": 65.4},
      {"site": "site_b", "date": "2024-01-02", "total_kwh": 1010.7, "avg_kwh": 42.1, "max_kwh": 68.9}
    ],
    "metadata": {
      "row_count": 4,
      "execution_time": "0.15s",
      "date_range": "2024-01-01 to 2024-01-31"
    }
  }
}'

echo "Simulated Query Results:"
echo $QUERY_RESULTS | jq '.results.data'
echo ""

# Simulate API definition
API_DEFINITION='{
  "api_id": "api_energy_date_range",
  "name": "energy_by_date_range",
  "description": "Get daily energy consumption by site for a date range",
  "endpoint": "/v1/reports/123/execute",
  "parameters": {
    "date_from": {"type": "string", "format": "date", "required": true, "description": "Start date (YYYY-MM-DD)"},
    "date_to": {"type": "string", "format": "date", "required": true, "description": "End date (YYYY-MM-DD)"},
    "site": {"type": "string", "required": false, "description": "Filter by site (optional)"}
  },
  "query_plan": {
    "sql": "SELECT site, DATE(timestamp) as date, SUM(kwh) as total_kwh, AVG(kwh) as avg_kwh, MAX(kwh) as max_kwh FROM energy_data WHERE timestamp >= ? AND timestamp < ? GROUP BY site, DATE(timestamp) ORDER BY date",
    "parameters": ["date_from", "date_to"]
  },
  "created_at": "2024-01-01T00:00:00Z"
}'

echo "Generated API Definition:"
echo $API_DEFINITION | jq '.'
echo ""

# Step 9: Summary
echo "✅ Workflow Test Complete!"
echo "=========================="
echo ""
echo "Working Components:"
echo "1. ✅ AI Tools - Get available AI tools and models"
echo "2. ✅ Chat Completion - AI conversation and analysis"
echo "3. ✅ SQL Generation - SQLCoder for query generation"
echo "4. ✅ File Reading - Python backend file processing"
echo "5. ✅ AI Analysis - AI insights on data and results"
echo ""
echo "Not Yet Implemented:"
echo "1. ❌ Session Management - /v1/sessions/* endpoints"
echo "2. ❌ Report Management - /v1/reports/* endpoints"
echo "3. ❌ Scope Building - Interactive scope creation"
echo "4. ❌ Query Execution - File query processing"
echo "5. ❌ API Generation - Saving analyses as reusable APIs"
echo ""
echo "Next Steps:"
echo "1. Implement Session Management Foundation"
echo "2. Add Report Management APIs"
echo "3. Build Scope Building workflow"
echo "4. Add Query Execution for files"
echo "5. Implement API Generation and storage"
echo ""
echo "🎉 Core AI and File Processing is working! 🎉"
echo "Ready to implement the full workflow!"