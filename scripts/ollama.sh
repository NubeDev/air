#!/usr/bin/env bash
set -euo pipefail

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: '$1' not found in PATH."
    exit 1
  fi
}
require ollama

MODELS=()

load_models() {
  MODELS=()
  # Parse the tabular output of `ollama list`
  # Header looks like: NAME  ID  SIZE  MODIFIED
  mapfile -t MODELS < <(ollama list 2>/dev/null \
    | awk 'NR==1 && $1=="NAME"{next} {print $1}' \
    | sed '/^$/d')
}

print_models() {
  echo
  echo "Installed models:"
  echo "-----------------"
  if [[ ${#MODELS[@]} -eq 0 ]]; then
    echo "  (none found)  → try: pull llama3"
  else
    local i=1
    for m in "${MODELS[@]}"; do
      printf "  %2d) %s\n" "$i" "$m"
      ((i++))
    done
  fi
  echo
}

select_model() {
  local sel="${1:-}"
  [[ -z "$sel" ]] && { echo "Provide a number or name."; return 1; }

  local model=""
  if [[ "$sel" =~ ^[0-9]+$ ]]; then
    local idx=$((sel-1))
    if (( idx >= 0 && idx < ${#MODELS[@]} )); then
      model="${MODELS[$idx]}"
    else
      echo "Invalid number: $sel"; return 1
    fi
  else
    # exact match, else prefix match
    for m in "${MODELS[@]}"; do [[ "$m" == "$sel" ]] && model="$m" && break; done
    if [[ -z "$model" ]]; then
      for m in "${MODELS[@]}"; do [[ "$m" == "$sel"* ]] && model="$m" && break; done
    fi
    [[ -z "$model" ]] && { echo "Model not found: $sel"; return 1; }
  fi
  MODEL_NAME="$model"
  return 0
}

run_model() {
  local model="$1"
  echo
  echo "Starting interactive session for: $model"
  echo "----------------------------------------"
  echo "Type prompts; Ctrl+C to exit."
  echo
  ollama run "$model"
}

stop_model() {
  local model="$1"
  echo "Stopping/unloading: $model"
  ollama stop "$model" || true
}

show_running() {
  echo
  echo "Running / recent sessions:"
  echo "--------------------------"
  ollama ps || true
  echo
}

pull_model() {
  local name="${1:-}"
  [[ -z "$name" ]] && { echo "Usage: pull <model>"; return 1; }
  echo "Pulling model: $name"
  ollama pull "$name"
}

rm_model() {
  local name="${1:-}"
  [[ -z "$name" ]] && { echo "Usage: rm <#|name>"; return 1; }
  if [[ "$name" =~ ^[0-9]+$ ]]; then
    if select_model "$name"; then name="$MODEL_NAME"; else return 1; fi
  fi
  echo "Removing model: $name"
  ollama rm "$name"
}

ensure_server() {
  # Lightweight hint if daemon not running
  if ! pgrep -x "ollama" >/dev/null 2>&1; then
    echo "Note: 'ollama serve' might not be running. If commands fail, start it in another terminal."
  fi
}

main_menu() {
  ensure_server
  while true; do
    load_models
    print_models
    echo "Options:"
    echo "  r <#|name>   Run a model (interactive)"
    echo "  s <#|name>   Stop/unload a model"
    echo "  p            Show running sessions (ollama ps)"
    echo "  pull <name>  Download a model (e.g., 'pull llama3')"
    echo "  rm <#|name>  Remove a model"
    echo "  l            Reload model list"
    echo "  q            Quit"
    echo
    read -rp "> " cmd arg || exit 0
    case "${cmd:-}" in
      r|R)   if select_model "${arg:-}"; then run_model "$MODEL_NAME"; fi ;;
      s|S)   if select_model "${arg:-}"; then stop_model "$MODEL_NAME"; fi ;;
      p|P)   show_running ;;
      pull)  pull_model "${arg:-}" ;;
      rm|RM) rm_model "${arg:-}" ;;
      l|L)   : ;;  # reload happens next loop
      q|Q|exit) exit 0 ;;
      *)
        # bare number/name ⇒ run
        if [[ -n "${cmd:-}" && -z "${arg:-}" ]]; then
          if select_model "$cmd"; then run_model "$MODEL_NAME"; fi
        else
          echo "Unknown. Examples: 'r 1', 's llama3', 'p', 'pull llama3', 'rm 2', 'q'"
        fi
        ;;
    esac
  done
}

main_menu
