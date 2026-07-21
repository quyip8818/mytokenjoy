#!/usr/bin/env bash
# Start backend + frontend + mock. No parameters — mode is in .env.development.
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

# Ensure child processes are killed on exit (prevents orphaned port listeners).
cleanup() {
  pkill -P $$ 2>/dev/null || true
  # Give children a moment to exit, then force-kill stragglers.
  sleep 0.5
  pkill -9 -P $$ 2>/dev/null || true
}
trap cleanup EXIT INT TERM

ENV_FILE="${ROOT}/apps/backend/.env.development"
SUPPORT_SAAS="$(grep '^SUPPORT_SAAS=' "${ENV_FILE}" 2>/dev/null | cut -d= -f2 || echo false)"

"${NEWAPI_SCRIPTS}/ensure-infra.sh"
concurrently --kill-others-on-fail --kill-signal SIGTERM -n backend,frontend,mock -c blue,green,magenta \
  "pnpm -F @tokenjoy/backend start" \
  "VITE_SUPPORT_SAAS=${SUPPORT_SAAS} bash \"${ROOT}/scripts/dev/frontend-wait.sh\" full" \
  "pnpm -F @tokenjoy/dev-mock-llm start"
