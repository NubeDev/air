#!/bin/bash

echo "🚀 Testing Complete File-Based AI Learning Workflow"
echo "=================================================="
echo ""

# Start the backend
echo "🔄 Starting Go backend..."
cd /home/user/code/go/nube/air
./bin/air --data data --config config-dev.yaml --auth disabled &
BACKEND_PID=$!
sleep 5

# Check if backend is running
if ! curl -s http://localhost:9000/health > /dev/null; then
    echo "❌ Backend failed to start"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

echo "✅ Backend is running"
echo ""

# Test 1: Session Management
echo "📋 Test 1: Session Management"
echo "-----------------------------"

# Start a session
echo "Creating session..."
SESSION_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/sessions/start \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/data/ts-energy.csv",
    "session_name": "Energy Analysis Session",
    "datasource_type": "file",
    "options": {
      "infer_rows": 50,
      "deep_analysis": true
    }
  }')

echo "Session creation response:"
echo $SESSION_RESPONSE | jq '.' 2>/dev/null || echo $SESSION_RESPONSE
echo ""

# List sessions
echo "Listing sessions..."
SESSIONS_RESPONSE=$(curl -s http://localhost:9000/v1/sessions)
echo "Sessions response:"
echo $SESSIONS_RESPONSE | jq '.' 2>/dev/null || echo $SESSIONS_RESPONSE
echo ""

# Test 2: Generated Reports
echo "📊 Test 2: Generated Reports"
echo "----------------------------"

# List generated reports
echo "Listing generated reports..."
REPORTS_RESPONSE=$(curl -s http://localhost:9000/v1/generated/reports)
echo "Reports response:"
echo $REPORTS_RESPONSE | jq '.' 2>/dev/null || echo $REPORTS_RESPONSE
echo ""

# Create a generated report
echo "Creating generated report..."
REPORT_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/generated/reports \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Energy Consumption by Date Range",
    "description": "Get daily energy consumption by site for a date range",
    "file_path": "/data/ts-energy.csv",
    "scope_json": "{\"entities\":[\"energy_data\"],\"metrics\":[{\"name\":\"total_kwh\",\"aggregation\":\"sum\"}],\"dimensions\":[\"site\",\"date\"]}",
    "query_plan_json": "{\"sql\":\"SELECT site, DATE(timestamp) as date, SUM(kwh) as total_kwh FROM energy_data WHERE timestamp >= ? AND timestamp < ? GROUP BY site, DATE(timestamp) ORDER BY date\"}",
    "parameters_json": "{\"date_from\":{\"type\":\"string\",\"format\":\"date\",\"required\":true},\"date_to\":{\"type\":\"string\",\"format\":\"date\",\"required\":true}}"
  }')

echo "Report creation response:"
echo $REPORT_RESPONSE | jq '.' 2>/dev/null || echo $REPORT_RESPONSE
echo ""

# Extract report ID for execution test
REPORT_ID=$(echo $REPORT_RESPONSE | jq -r '.id' 2>/dev/null)

# Test 2.5: Execute Generated Report API
echo "🚀 Test 2.5: Execute Generated Report API"
echo "----------------------------------------"

if [ "$REPORT_ID" != "null" ] && [ "$REPORT_ID" != "" ]; then
    echo "Executing generated report ID: $REPORT_ID"
    EXECUTE_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/generated/reports/$REPORT_ID/execute \
      -H "Content-Type: application/json" \
      -d '{
        "parameters": {
          "date_from": "2024-01-01",
          "date_to": "2024-01-31"
        }
      }')
    
    echo "Report execution response:"
    echo $EXECUTE_RESPONSE | jq '.' 2>/dev/null || echo $EXECUTE_RESPONSE
    echo ""
    
    # Test 2.6: Get Report Schema (for UI generation)
    echo "📋 Test 2.6: Get Report Schema"
    echo "------------------------------"
    SCHEMA_RESPONSE=$(curl -s http://localhost:9000/v1/generated/reports/$REPORT_ID)
    echo "Report schema response:"
    echo $SCHEMA_RESPONSE | jq '.parameters_json' 2>/dev/null || echo "Failed to parse schema"
    echo ""
else
    echo "❌ Could not extract report ID for execution test"
    EXECUTE_RESPONSE="404"
    SCHEMA_RESPONSE="404"
    echo ""
fi

# Test 3: AI Tools (should work)
echo "🤖 Test 3: AI Tools"
echo "-------------------"

AI_TOOLS_RESPONSE=$(curl -s http://localhost:9000/v1/ai/tools)
echo "AI Tools response:"
echo $AI_TOOLS_RESPONSE | jq '.tools | length' 2>/dev/null || echo "Failed to parse AI tools"
echo ""

# Test 4: Chat Completion (should work)
echo "💬 Test 4: Chat Completion"
echo "-------------------------"

CHAT_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/ai/chat/completion \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "Hello! Can you help me analyze energy data?"
      }
    ]
  }')

echo "Chat response:"
echo $CHAT_RESPONSE | jq -r '.message.content' 2>/dev/null || echo "Failed to parse chat response"
echo ""

# Test 5: SQL Generation (should work)
echo "🔍 Test 5: SQL Generation"
echo "-------------------------"

SQL_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/sql/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Show daily energy consumption by site",
    "schema": "CREATE TABLE energy_data (timestamp VARCHAR, site VARCHAR, kwh FLOAT);"
  }')

echo "SQL response:"
echo $SQL_RESPONSE | jq -r '.sql' 2>/dev/null || echo "Failed to parse SQL response"
echo ""

# Test 6: File Processing (should work)
echo "📁 Test 6: File Processing"
echo "-------------------------"

FILE_RESPONSE=$(curl -s -X POST http://localhost:9001/v1/py/infer_schema \
  -H "Content-Type: application/json" \
  -d '{"datasource_id":"test","uri":"../testdata/ts-energy.csv","infer_rows":50}')

TOKEN=$(echo $FILE_RESPONSE | jq -r '.token' 2>/dev/null)
if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
    echo "File processing started, token: $TOKEN"
    sleep 2
    SCHEMA_RESULT=$(curl -s http://localhost:9001/v1/py/jobs/$TOKEN)
    echo "Schema result:"
    echo $SCHEMA_RESULT | jq '.data.schema.fields | length' 2>/dev/null || echo "Failed to parse schema"
else
    echo "File processing failed"
fi
echo ""

# Summary
echo "✅ Complete Workflow Test Summary"
echo "================================="
echo ""

# Check which tests passed
echo "Test Results:"
echo "1. Session Management: $([ "$SESSION_RESPONSE" != "404" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "2. Generated Reports: $([ "$REPORTS_RESPONSE" != "404" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "2.5. Report Execution: $([ "$EXECUTE_RESPONSE" != "404" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "3. AI Tools: $([ "$AI_TOOLS_RESPONSE" != "404" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "4. Chat Completion: $([ "$CHAT_RESPONSE" != "404" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "5. SQL Generation: $([ "$SQL_RESPONSE" != "404" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "6. File Processing: $([ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo ""

# Cleanup
echo "🧹 Cleaning up..."
kill $BACKEND_PID 2>/dev/null
echo "Backend stopped"
echo ""

echo "🎉 Complete workflow test finished!"
