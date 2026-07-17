#!/usr/bin/env bash
# Start in local/private-deploy mode (SUPPORT_SAAS=false, single company).
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

export SUPPORT_SAAS=false

"${NEWAPI_SCRIPTS}/ensure-infra.sh"
concurrently --kill-others-on-fail -n backend,frontend,mock -c blue,green,magenta \
  "pnpm -F @tokenjoy/backend start" \
  "VITE_SUPPORT_SAAS=false bash \"${ROOT}/scripts/dev/frontend-wait.sh\" full" \
  "pnpm -F @tokenjoy/dev-mock-llm start"
