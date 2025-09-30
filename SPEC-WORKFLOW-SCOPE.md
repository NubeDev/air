# AIR Workflow: Scope → IR → SQL/Plan → Execute → API

## Model responsibilities
- Scope discovery/refinement: Llama/ChatGPT
- SQL generation: SQLCoder (only)
- File reads/processing: Python FastAPI (only)
- DB SQL execution: Go against registered datasources

## Steps
1) Start session (file or DB)
   - Files: POST `/v1/sessions/start` (Python handles file reads)
   - DBs: Register via `/v1/datasources`; then learn via DB learn endpoints
   - Artifact: session row in control plane (SQLite `air.db`)

2) Learn (overview of structure)
   - Files: Python infer schema, profiling, sample (see SPEC-PY.md)
   - DBs: `POST /v1/learn?datasource_id=...` → `SchemaNote`
   - Artifact: schema notes + minimal DDL for prompts

3) Scope build (interactive)
   - Llama/ChatGPT: `POST /v1/sessions/{id}/scope/build` (planned) and `/scope/refine`
   - Artifact: `scopes`, `scope_versions.scope_md`

4) AI generates/refines scope
   - Llama/ChatGPT: refinement endpoint saves a new scope version
   - Artifact: updated `scope_versions`

5) Approve scope
   - User approves a scope version for IR build
   - Artifact: approved `scope_versions`

6) Generate query (SQL or file plan)
   - Build IR from scope (LLM-assisted) → `scope_versions.ir_json`
   - DBs: SQLCoder converts IR + minimal DDL → SQL (no heuristic fallback)
     - Endpoint: `POST /v1/sql` (IR + datasource_id)
   - Files: IR → file execution plan; Python executes for a sample
   - Artifact: SQL text or file plan + sample results

7) Approve and save as API
   - Files: `POST /v1/generated/reports` → `/v1/generated/reports/{id}/execute`
   - DBs: `POST /v1/reports` + versions → `/v1/reports/{id}/execute`
   - Artifacts: `generated_reports` (files) or `reports` + `report_versions` (DBs)

8) Execute at scale
   - Files: Python executes plan over file(s)
   - DBs: Go executes SQL on selected datasource (SQLite/Postgres/MySQL/Timescale)
   - Artifact: run records (`report_runs`, `report_executions`) with results metadata

## Notes
- No fake data; failures return actionable errors.
- Separate analytics SQLite DB (not `air.db`) is used for DB queries.
- Supports SQLite, Postgres, MySQL, Timescale, and Files.