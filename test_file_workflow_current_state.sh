#!/usr/bin/env bash
set -euo pipefail

# Config
BASE_URL="http://localhost:9000"
FASTAPI_URL="http://localhost:9001"
CSV_PATH="testdata/ts-energy.csv"

echo "== AIR FILE WORKFLOW DEMO (current state) =="

pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; exit 1; }

jq_exists() {
  command -v jq >/dev/null 2>&1 || { echo "jq required"; exit 1; }
}

curl_json() {
  curl -sS -H "Content-Type: application/json" "$@"
}

jq_exists

echo "1) Health checks"
echo "  Backend health:"
backend_resp=$(curl -sS "$BASE_URL/health")
echo "$backend_resp" | jq .
backend_health=$(echo "$backend_resp" | jq -r .status || true)
[[ "$backend_health" == "healthy" ]] && pass "Backend healthy" || fail "Backend not healthy"

echo "  FastAPI health:"
fastapi_resp=$(curl -sS "$BASE_URL/v1/fastapi/health")
echo "$fastapi_resp" | jq .
fastapi_health=$(echo "$fastapi_resp" | jq -r .status || true)
[[ "$fastapi_health" == "ok" ]] && pass "FastAPI proxy healthy" || fail "FastAPI proxy not healthy"

echo "2) Start file session"
echo "  Request:"
echo "  {\"file_path\": \"$CSV_PATH\", \"session_name\": \"file-demo\", \"datasource_type\": \"file\", \"options\": {}}"
echo "  Response:"
SESSION_JSON=$(curl_json -X POST "$BASE_URL/v1/sessions/start" -d "$(jq -n --arg path "$CSV_PATH" '{file_path: $path, session_name: "file-demo", datasource_type: "file", options: {}}')")
echo "$SESSION_JSON" | jq .
SESSION_ID=$(echo "$SESSION_JSON" | jq -r .id)
[[ "$SESSION_ID" != "null" && -n "$SESSION_ID" ]] && pass "Session created id=$SESSION_ID" || fail "Failed to create session"

echo "3) FastAPI infer schema + preview via helper endpoints"
echo "  Response (truncated):"
ENERGY_TEST=$(curl -sS -X POST "$BASE_URL/v1/fastapi/test/energy" -H "Content-Type: application/json" -d '{}')
echo "$ENERGY_TEST" | jq '{status: .status, message: .message, results_keys: (.results | keys)}'
ENERGY_STATUS=$(echo "$ENERGY_TEST" | jq -r .status)
[[ "$ENERGY_STATUS" == "success" ]] && pass "FastAPI schema+preview+analysis ok" || fail "FastAPI test/energy failed"

echo "4) Create scope and scope version"
echo "  Request: {\"name\":\"energy_file_scope\"}"
echo "  Response:"
SCOPE=$(curl_json -X POST "$BASE_URL/v1/scopes" -d '{"name":"energy_file_scope"}')
echo "$SCOPE" | jq .
SCOPE_ID=$(echo "$SCOPE" | jq -r .id)
[[ "$SCOPE_ID" != "null" && -n "$SCOPE_ID" ]] && pass "Scope created id=$SCOPE_ID" || fail "Create scope failed: $SCOPE"

SCOPE_MD=$(cat <<'EOF'
Goal: Summarize daily energy kWh by site for a date range.
Dataset: energy CSV with columns [timestamp, site, kwh, voltage, current, temperature].
Filters: date_from inclusive.
Select: site, sum(kwh) as total_kwh.
Group by: site.
Order by: site asc.
Limit: 1000.
EOF
)

echo "  Creating scope version with Markdown:"
echo "  $SCOPE_MD"
echo "  Response:"
SCOPE_VER=$(curl_json -X POST "$BASE_URL/v1/scopes/$SCOPE_ID/version" -d "$(jq -n --arg md "$SCOPE_MD" '{scope_md: $md}')")
echo "$SCOPE_VER" | jq .
SCOPE_VER_ID=$(echo "$SCOPE_VER" | jq -r .id)
[[ "$SCOPE_VER_ID" != "null" && -n "$SCOPE_VER_ID" ]] && pass "Scope version created id=$SCOPE_VER_ID" || fail "Create scope version failed: $SCOPE_VER"

echo "5) User provides initial scope feedback"
echo "  👤 User: 'I want to see energy consumption by site, but I also need to filter by device type and see daily totals, not just site totals.'"
echo "  📝 Updating scope with user feedback..."

SCOPE_MD_V2=$(cat <<'EOF'
Goal: Analyze daily energy consumption by site and device type for a date range.
Dataset: energy CSV with columns [timestamp, site, device_id, kwh, voltage, current, temperature, status].
Filters: date_from inclusive, device_type optional filter.
Select: site, device_id, date(timestamp) as date, sum(kwh) as total_kwh, avg(voltage) as avg_voltage.
Group by: site, device_id, date(timestamp).
Order by: site asc, device_id asc, date asc.
Limit: 1000.
EOF
)

echo "  Creating refined scope version:"
echo "  $SCOPE_MD_V2"
echo "  Response:"
SCOPE_VER_V2=$(curl_json -X POST "$BASE_URL/v1/scopes/$SCOPE_ID/version" -d "$(jq -n --arg md "$SCOPE_MD_V2" '{scope_md: $md}')")
echo "$SCOPE_VER_V2" | jq .
SCOPE_VER_V2_ID=$(echo "$SCOPE_VER_V2" | jq -r .id)
[[ "$SCOPE_VER_V2_ID" != "null" && -n "$SCOPE_VER_V2_ID" ]] && pass "Refined scope version created id=$SCOPE_VER_V2_ID" || fail "Create refined scope version failed: $SCOPE_VER_V2"

echo "6) Build IR from refined scope (LLM)"
echo "  🤖 AI: 'Processing your refined requirements...'"
echo "  Request: {\"scope_version_id\": $SCOPE_VER_V2_ID}"
echo "  Response:"
IR_RESP=$(curl_json -X POST "$BASE_URL/v1/ir/build" -d "$(jq -n --argjson id $SCOPE_VER_V2_ID '{scope_version_id: $id}')")
echo "$IR_RESP" | jq .
IR_JSON=$(echo "$IR_RESP" | jq -c .ir 2>/dev/null || true)
[[ -n "$IR_JSON" && "$IR_JSON" != "null" ]] && pass "IR built from refined scope" || fail "Build IR failed: $IR_RESP"

echo "7) Generate SQL from IR (SQLCoder) targeting sqlite-dev if available"
DSID="sqlite-dev"
echo "  Request:"
echo "  {\"ir\": $IR_JSON, \"datasource_id\": \"$DSID\"}"
echo "  Response:"
SQL_REQ=$(jq -n --argjson ir "$IR_JSON" --arg ds "$DSID" '{ir: $ir, datasource_id: $ds}')
SQL_RESP=$(curl_json -X POST "$BASE_URL/v1/sql" -d "$SQL_REQ")
echo "$SQL_RESP" | jq .
SQL_TEXT=$(echo "$SQL_RESP" | jq -r .sql || true)
[[ -n "$SQL_TEXT" && "$SQL_TEXT" != "null" ]] && pass "SQL generated" || echo "⚠️ SQL generation may require SQLCoder model in Ollama"

echo "8) User reviews generated SQL and provides final feedback"
echo "  🤖 AI: 'Here's the SQL I generated based on your requirements:'"
echo "  📋 Generated SQL:"
echo "  $(echo "$SQL_TEXT" | sed 's/^/    /')"
echo ""
echo "  👤 User: 'That looks good, but can you add a filter for device status = active only?'"
echo "  🤖 AI: 'I'll update the scope to include the device status filter...'"

SCOPE_MD_V3=$(cat <<'EOF'
Goal: Analyze daily energy consumption by site and device type for a date range, active devices only.
Dataset: energy CSV with columns [timestamp, site, device_id, kwh, voltage, current, temperature, status].
Filters: date_from inclusive, device_type optional filter, status = 'active'.
Select: site, device_id, date(timestamp) as date, sum(kwh) as total_kwh, avg(voltage) as avg_voltage.
Group by: site, device_id, date(timestamp).
Order by: site asc, device_id asc, date asc.
Limit: 1000.
EOF
)

echo "  Creating final scope version with status filter:"
echo "  $SCOPE_MD_V3"
echo "  Response:"
SCOPE_VER_V3=$(curl_json -X POST "$BASE_URL/v1/scopes/$SCOPE_ID/version" -d "$(jq -n --arg md "$SCOPE_MD_V3" '{scope_md: $md}')")
echo "$SCOPE_VER_V3" | jq .
SCOPE_VER_V3_ID=$(echo "$SCOPE_VER_V3" | jq -r .id)
[[ "$SCOPE_VER_V3_ID" != "null" && -n "$SCOPE_VER_V3_ID" ]] && pass "Final scope version created id=$SCOPE_VER_V3_ID" || fail "Create final scope version failed: $SCOPE_VER_V3"

echo "  🤖 AI: 'Rebuilding IR with your final requirements...'"
echo "  Request: {\"scope_version_id\": $SCOPE_VER_V3_ID}"
echo "  Response:"
IR_RESP_V3=$(curl_json -X POST "$BASE_URL/v1/ir/build" -d "$(jq -n --argjson id $SCOPE_VER_V3_ID '{scope_version_id: $id}')")
echo "$IR_RESP_V3" | jq .
IR_JSON_V3=$(echo "$IR_RESP_V3" | jq -c .ir 2>/dev/null || true)
[[ -n "$IR_JSON_V3" && "$IR_JSON_V3" != "null" ]] && pass "Final IR built" || fail "Build final IR failed: $IR_RESP_V3"

echo "  🤖 AI: 'Generating final SQL with all your requirements...'"
echo "  Request:"
echo "  {\"ir\": $IR_JSON_V3, \"datasource_id\": \"$DSID\"}"
echo "  Response:"
SQL_REQ_V3=$(jq -n --argjson ir "$IR_JSON_V3" --arg ds "$DSID" '{ir: $ir, datasource_id: $ds}')
SQL_RESP_V3=$(curl_json -X POST "$BASE_URL/v1/sql" -d "$SQL_REQ_V3")
echo "$SQL_RESP_V3" | jq .
SQL_TEXT_V3=$(echo "$SQL_RESP_V3" | jq -r .sql || true)
[[ -n "$SQL_TEXT_V3" && "$SQL_TEXT_V3" != "null" ]] && pass "Final SQL generated" || echo "⚠️ Final SQL generation may require SQLCoder model in Ollama"

echo "  📋 Final SQL:"
echo "  $(echo "$SQL_TEXT_V3" | sed 's/^/    /')"
echo ""
echo "  👤 User: 'Perfect! This looks exactly what I need. Let's save this as a report.'"
echo "  🤖 AI: 'Great! I'll create the report for you...'"

echo "9) Create DB report and version with final SQL (if SQL available)"
if [[ -n "$SQL_TEXT_V3" && "$SQL_TEXT_V3" != "null" ]]; then
  TIMESTAMP=$(date +%s)
  echo "  Request: {\"key\":\"energy_file_demo_$TIMESTAMP\",\"title\":\"Energy File Demo $TIMESTAMP\"}"
  echo "  Response:"
  REP=$(curl_json -X POST "$BASE_URL/v1/reports" -d "{\"key\":\"energy_file_demo_$TIMESTAMP\",\"title\":\"Energy File Demo $TIMESTAMP\"}")
  echo "$REP" | jq .
  REP_ID=$(echo "$REP" | jq -r .id)
  [[ "$REP_ID" != "null" && -n "$REP_ID" ]] && pass "Report created id=$REP_ID" || fail "Create report failed: $REP"

  echo "  Creating report version:"
  DEF=$(jq -n --arg sql "$SQL_TEXT_V3" '{sql: $sql}')
  echo "  Request: {\"scope_version_id\": $SCOPE_VER_V3_ID, \"datasource_id\": \"$DSID\", \"def_json\": $DEF}"
  echo "  Response:"
  VER=$(curl_json -X POST "$BASE_URL/v1/reports/$REP_ID/versions" -d "$(jq -n --argjson sid $SCOPE_VER_V3_ID --arg ds "$DSID" --arg def "$DEF" '{scope_version_id: $sid, datasource_id: $ds, def_json: $def}')")
  echo "$VER" | jq .
  VER_ID=$(echo "$VER" | jq -r .id)
  [[ "$VER_ID" != "null" && -n "$VER_ID" ]] && pass "Report version created id=$VER_ID" || fail "Create report version failed: $VER"

  echo "10) Execute report with user parameters"
  echo "  👤 User: 'Let me test this with some real data from January 2024'"
  echo "  🤖 AI: 'Executing your report with the specified parameters...'"
  echo "  Request: {\"params\": {\"date_from\": \"2024-01-01\"}, \"datasource_id\": \"sqlite-dev\"}"
  echo "  Response:"
  RUN=$(curl_json -X POST "$BASE_URL/v1/reports/$REP_ID/execute" -d '{"params": {"date_from": "2024-01-01"}, "datasource_id": "sqlite-dev"}')
  echo "$RUN" | jq .
  RUN_STATUS=$(echo "$RUN" | jq -r .status)
  ROWS=$(echo "$RUN" | jq -r .row_count)
  if [[ "$RUN_STATUS" != "failed" ]]; then
    pass "Report executed successfully rows=$ROWS"
    echo "  👤 User: 'Excellent! I can see $ROWS rows of data. This is exactly what I needed.'"
    echo "  🤖 AI: 'Perfect! Your report is now ready and can be executed anytime with different date ranges.'"
  else
    echo "⚠️ Report execution failed: $RUN"
    echo "  👤 User: 'Hmm, there seems to be an issue with the execution.'"
    echo "  🤖 AI: 'Let me help you troubleshoot this. The error suggests there might be a data format issue.'"
  fi
else
  echo "Skipping DB report steps due to missing SQL (SQLCoder model not ready)"
fi

echo "11) Generated Reports (file-based API) smoke"
echo "  Response:"
GR_RESP=$(curl -sS "$BASE_URL/v1/generated/reports")
echo "$GR_RESP" | jq .
GR_LIST=$(echo "$GR_RESP" | jq -r 'type' || true)
[[ "$GR_LIST" == "array" ]] && pass "Generated reports list ok" || echo "ℹ️ Generated reports yet to be created"

echo ""
echo "🎉 WORKFLOW COMPLETE! 🎉"
echo ""
echo "📊 Summary of User-AI Interaction:"
echo "  • User provided initial requirements for energy analysis"
echo "  • AI generated initial scope and IR"
echo "  • User requested refinements (device type, daily totals)"
echo "  • AI updated scope and regenerated IR"
echo "  • User requested additional filter (device status = active)"
echo "  • AI created final scope with all requirements"
echo "  • User approved final SQL and saved as report"
echo "  • User tested report execution with real parameters"
echo "  • AI confirmed successful execution and report readiness"
echo ""
echo "✅ This demonstrates a realistic iterative scope refinement process!"
echo "   All steps attempted. Review warnings if any."

