#!/usr/bin/env bash
# Start in local/private-deploy mode with BOOTSTRAP_MODE=prod (empty: no demo data).
# Creates company + admin but no members, budgets, or demo fixtures.
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

export SUPPORT_SAAS=false
export BOOTSTRAP_MODE=prod
export BOOTSTRAP_CONFIG_PATH="${ROOT}/apps/backend/config/bootstrap-local.yaml"

"${NEWAPI_SCRIPTS}/ensure-infra.sh"
concurrently --kill-others-on-fail -n backend,frontend,mock -c blue,green,magenta \
  "pnpm -F @tokenjoy/backend start" \
  "VITE_SUPPORT_SAAS=false bash \"${ROOT}/scripts/dev/frontend-wait.sh\" full" \
  "pnpm -F @tokenjoy/dev-mock-llm start"
