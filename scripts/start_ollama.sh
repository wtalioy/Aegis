#!/usr/bin/env bash
set -euo pipefail

MODEL="${1:-}"
ENDPOINT_ARG="${2:-}"
if [[ -z "$MODEL" ]]; then
  echo "Usage: $0 <model> [endpoint]" >&2
  exit 1
fi

if [[ -n "$ENDPOINT_ARG" ]]; then
  OLLAMA_HOST="$ENDPOINT_ARG"
else
  OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
fi
LOG_FILE="${OLLAMA_LOG:-/tmp/ollama.log}"

export OLLAMA_HOST

if [[ -z "$LOG_FILE" ]]; then
  LOG_FILE="/tmp/ollama.log"
fi

echo "[ollama] Target endpoint: $OLLAMA_HOST"
echo "[ollama] Target model: $MODEL"

echo "[ollama] Ensuring Ollama CLI is installed..."
if ! command -v ollama >/dev/null 2>&1; then
  if ! command -v curl >/dev/null 2>&1; then
    echo "[ollama] curl is required to install Ollama." >&2
    exit 1
  fi
  echo "[ollama] Installing Ollama..."
  curl -fsSL https://ollama.com/install.sh | sh
else
  echo "[ollama] Ollama already installed."
fi

start_ollama() {
  if pgrep -x ollama >/dev/null 2>&1; then
    echo "[ollama] Ollama daemon already running."
    return 0
  fi

  if command -v systemctl >/dev/null 2>&1 && systemctl list-unit-files --type=service 2>/dev/null | grep -q '^ollama\.service'; then
    echo "[ollama] Starting Ollama via systemd..."
    if command -v sudo >/dev/null 2>&1; then
      sudo systemctl start ollama
    else
      systemctl start ollama
    fi
  else
    echo "[ollama] Launching 'ollama serve' in background (log: $LOG_FILE)..."
    nohup ollama serve >"$LOG_FILE" 2>&1 &
    sleep 2
  fi
}

ensure_model() {
  local model="$1"
  if [[ -z "$model" ]]; then
    echo "[ollama] No model provided; skipping pull."
    return 0
  fi

  if ollama show "$model" >/dev/null 2>&1; then
    echo "[ollama] Model '$model' already present."
  else
    echo "[ollama] Pulling model '$model'..."
    ollama pull "$model"
  fi

  echo "[ollama] Pre-loading model '$model' into memory..."
  if curl -fsS -X POST "$OLLAMA_HOST/api/generate" \
    -H "Content-Type: application/json" \
    -d "{\"model\":\"$model\",\"prompt\":\"\",\"stream\":false,\"options\":{\"num_predict\":1}}" \
    >/dev/null 2>&1; then
    echo "[ollama] Model '$model' pre-loaded successfully."
  else
    echo "[ollama] Warning: Failed to pre-load model '$model', it will be loaded on first use."
  fi
}

start_ollama

echo "[ollama] Waiting for Ollama API at $OLLAMA_HOST ..."
for attempt in {1..30}; do
  if curl -fsS "$OLLAMA_HOST/api/tags" >/dev/null 2>&1; then
    echo "[ollama] Ollama is ready."
    READY=1
    break
  fi
  sleep 1
  echo "[ollama] Still waiting (attempt $attempt)..."
done

if [[ "${READY:-0}" != 1 ]]; then
  echo "[ollama] Timed out waiting for Ollama API at $OLLAMA_HOST" >&2
  exit 1
fi

ensure_model "$MODEL"
