#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

"${NEWAPI_SCRIPTS}/ensure-infra.sh"
concurrently --kill-others-on-fail -n backend,frontend,mock -c blue,green,magenta \
  "pnpm -F @tokenjoy/backend start" \
  "bash \"${ROOT}/scripts/dev/frontend-wait.sh\" full" \
  "pnpm -F @tokenjoy/dev-mock-llm start"
