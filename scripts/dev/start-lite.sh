#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

cleanup() {
  pkill -P $$ 2>/dev/null || true
  sleep 0.5
  pkill -9 -P $$ 2>/dev/null || true
}
trap cleanup EXIT INT TERM

"${COMPOSE[@]}" up postgres -d --wait
concurrently --kill-others-on-fail --kill-signal SIGTERM -n backend,frontend -c blue,green \
  "pnpm -F @tokenjoy/backend start" \
  "bash \"${ROOT}/scripts/dev/frontend-wait.sh\" lite"
