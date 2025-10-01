# AI Models Refactor – Generic, Config‑Driven Registry

Purpose
- Remove all hardcoded model names from backend/UI.
- Centralize model discovery, defaults, and health in config + API.
- Support adding/removing models (e.g., qwen3) with only config changes.

## Definitions
- Model ID: `<provider>:<modelName>` (examples: `openai:gpt-4o-mini`, `ollama:llama3:latest`, `ollama:sqlcoder:7b`).
- Capability: `chat`, `sql`, `embeddings` (extensible).

## Config (v2) – Minimal Registry + Defaults

Add to `models`:
- `defaults`: map<string,string> – default model IDs per capability.
- Providers (openai, ollama) keep credentials/host and optional `registry` entries.

Example
```
models:
  defaults:
    chat: "openai:gpt-4o-mini"
    sql:  "ollama:sqlcoder:7b"
    embeddings: "openai:text-embedding-3-small"

  openai:
    api_key: ${OPENAI_API_KEY}
    endpoint: https://api.openai.com/v1
    registry:
      gpt-4o-mini:
        id: "openai:gpt-4o-mini"
        capabilities: ["chat"]

  ollama:
    host: http://localhost:11434
    registry:
      llama3:latest:
        id: "ollama:llama3:latest"
        capabilities: ["chat"]
      sqlcoder:7b:
        id: "ollama:sqlcoder:7b"
        capabilities: ["sql"]
      qwen3:7b:
        id: "ollama:qwen3:7b"
        capabilities: ["chat"]
```

Back‑compat loader
- If old fields (chat_primary/sql_primary, llama3_model/sqlcoder_model) are present, synthesize `defaults` and minimum `registry` entries.

## Backend Changes

New types (internal/config)
- `ModelsConfig` extends with `Defaults map[string]string` and `ProviderConfig{Endpoint/Host, Registry map[string]ModelRef}`.
- `ModelRef{ID string, Capabilities []string}` (provider inferred from prefix).

Resolver utilities (internal/llm)
- `ParseModelID(id string) (provider, name string)`.
- `NewClientFor(provider string, cfg *config.Config) (llm.LLMClient, error)`.

AIService API (internal/services/ai_service.go)
- `ResolveModel(capability string, overrideID string) (provider, name string, id string, error)`
  - If `overrideID` set → validate against registry; else use `models.defaults[capability]`.
  - If not in registry, still allow (advanced) but mark `registry=false` in response.
- `GetModels(ctx)`
  - Returns `{ defaults, models, health }`.
  - `models`: flatten all provider registries into `[{id, provider, name, capabilities}]`.
  - `health`: probe providers (OpenAI key present → health; Ollama host → health). No model‑specific hardcoding.
- `SetPrimary(capability, modelID)`
  - Update in‑memory defaults. No file write.

Handlers
- `POST /v1/ai/chat/completion` accepts `model?: string` (ID). Use `ResolveModel("chat", req.model)`.
- `POST /v1/ir/build` adds optional `model` (ID). Use `ResolveModel("chat", ...)`.
- `POST /v1/sql` adds optional `model` (ID). Use `ResolveModel("sql", ...)`.
- WebSocket chat and raw: remove alias maps; pass through `model` ID to resolver.

Model routes
- `GET /v1/ai/models` → `{ defaults, models }` from `GetModels` (without health if desired).
- `GET /v1/ai/models/status` → `{ defaults, models, health }`.
- `POST /v1/ai/models/primary` → body `{ capability, model }` uses `SetPrimary`.

Remove/replace all hardcoded references
- Delete switch/case mappings like `openai→gpt-4o-mini`, `llama→llama3:latest`, `sqlcoder→sqlcoder:7b`.
- Any default choice is via `models.defaults`.

## UI Changes
- Fetch `GET /v1/ai/models` (or `/status`) to populate selectors.
- Store selection per capability in localStorage; send `model` ID on requests.
- Chat header selector uses models with `chat` capability.
- Workflow → Generate SQL uses models with `sql` capability.
- No hardcoded label lists. Render `display` from ID/name; capability badges optional.

## Health & Discovery
- Providers’ health is tracked at provider level. Model presence isn’t enforced (e.g., Ollama tags may be enumerated later).
- UI can display provider health and allow selections regardless; backend returns clear errors if the provider/model is unavailable.

## Tests
- Unit: ParseModelID, ResolveModel precedence, back‑compat loader.
- Integration: `/v1/ai/models`, `/status`, `/primary` flows; chat/sql with and without overrides.
- Regression: websocket flows without alias mapping; ensure no hardcoded strings remain.

## Rollout Steps
1) Add config structs + back‑compat loader.
2) Implement resolver + client factory.
3) Replace handlers and websocket to use model IDs.
4) Add models APIs.
5) Update UI to dynamic lists and send model IDs.
6) Remove dead code (alias maps, hardcoded checks).

## Non‑Goals (phase 1)
- Persisting `SetPrimary` to disk (in‑memory only).
- Full dynamic model discovery from providers (beyond simple health signals/registry hints).


