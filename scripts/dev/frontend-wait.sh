#!/usr/bin/env bash
# Wait for backend (normal or setup mode) then start frontend dev server.
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

BACKEND_PORT="${BACKEND_PORT:-8011}"

# Backend may be in normal mode (/healthz) or setup mode (/api/setup/status).
# Poll until either endpoint responds with HTTP 200.
echo "Waiting for backend on port ${BACKEND_PORT}..."
deadline=$((SECONDS + 60))
while (( SECONDS < deadline )); do
  if curl -sf "http://127.0.0.1:${BACKEND_PORT}/healthz" >/dev/null 2>&1; then
    echo "Backend ready (normal mode)"
    break
  fi
  if curl -sf "http://127.0.0.1:${BACKEND_PORT}/api/setup/status" >/dev/null 2>&1; then
    echo "Backend ready (setup mode) — open http://localhost:${VITE_DEV_PORT:-${FRONTEND_PORT:-9192}}/setup"
    break
  fi
  sleep 0.5
done

if (( SECONDS >= deadline )); then
  echo "Timeout waiting for backend on port ${BACKEND_PORT}" >&2
  exit 1
fi

exec pnpm -F @tokenjoy/frontend start
