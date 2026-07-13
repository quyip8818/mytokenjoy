#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

"${COMPOSE[@]}" up postgres -d --wait
concurrently --kill-others-on-fail -n backend,frontend -c blue,green \
  "pnpm -F @tokenjoy/backend start" \
  "bash \"${ROOT}/scripts/dev/frontend-wait.sh\" lite"
