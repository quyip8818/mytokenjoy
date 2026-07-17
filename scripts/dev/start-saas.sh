#!/usr/bin/env bash
# Start in SaaS mode (SUPPORT_SAAS=true, multi-tenant, platform operator, trial registration).
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

export SUPPORT_SAAS=true
export PLATFORM_BOOTSTRAP_EMAIL="${PLATFORM_BOOTSTRAP_EMAIL:-ops@tokenjoy.local}"
export PLATFORM_BOOTSTRAP_PASSWORD="${PLATFORM_BOOTSTRAP_PASSWORD:-platform-dev-123}"
export PLATFORM_SESSION_SECRET="${PLATFORM_SESSION_SECRET:-tokenjoy-platform-dev-secret}"
export REGISTRATION_ENABLED="${REGISTRATION_ENABLED:-true}"

"${NEWAPI_SCRIPTS}/ensure-infra.sh"
concurrently --kill-others-on-fail -n backend,frontend,mock -c blue,green,magenta \
  "pnpm -F @tokenjoy/backend start" \
  "bash \"${ROOT}/scripts/dev/frontend-wait.sh\" full" \
  "pnpm -F @tokenjoy/dev-mock-llm start"
