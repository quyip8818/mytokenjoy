#!/usr/bin/env bash
# Start backend + frontend + mock. No parameters — mode is in .env.development.
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

ENV_FILE="${ROOT}/apps/backend/.env.development"
SUPPORT_SAAS="$(grep '^SUPPORT_SAAS=' "${ENV_FILE}" 2>/dev/null | cut -d= -f2 || echo false)"

"${NEWAPI_SCRIPTS}/ensure-infra.sh"
concurrently --kill-others-on-fail -n backend,frontend,mock -c blue,green,magenta \
  "pnpm -F @tokenjoy/backend start" \
  "VITE_SUPPORT_SAAS=${SUPPORT_SAAS} bash \"${ROOT}/scripts/dev/frontend-wait.sh\" full" \
  "pnpm -F @tokenjoy/dev-mock-llm start"
