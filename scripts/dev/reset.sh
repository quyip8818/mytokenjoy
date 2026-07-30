#!/usr/bin/env bash
# Reset apps databases and re-seed data.
# Usage: pnpm reset [local|saas]
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"
# shellcheck source=../lib/db-reset.sh
source "${ROOT}/scripts/lib/db-reset.sh"

MODE=local
for arg in "$@"; do
  case "${arg}" in
    local|saas) MODE="${arg}" ;;
    -h|--help)
      echo "usage: pnpm reset [local|saas]"
      echo "  local (default): empty DB, triggers setup flow on next start"
      echo "  saas: multi-tenant platform + demo company (full seed)"
      exit 0
      ;;
    *) echo "unknown arg: ${arg}" >&2; exit 1 ;;
  esac
done

# Load mode-specific env (sets COMPOSE, DATABASE_URL, NEWAPI_URL, etc.)
# shellcheck source=../lib/mode-env.sh
source "${ROOT}/scripts/lib/mode-env.sh"

# --- Wipe & rebuild (SQL drop/create, preserves sms databases) ---
"${COMPOSE[@]}" stop newapi-apps 2>/dev/null || true
"${COMPOSE[@]}" up postgres redis -d --wait
reset_apps_databases

"${NEWAPI_SCRIPTS}/bootstrap-local-after-reset.sh"
"${COMPOSE[@]}" exec -T redis redis-cli -n 0 FLUSHDB
"${COMPOSE[@]}" exec -T redis redis-cli -n 2 FLUSHDB

# Seed data
pnpm -F @tokenjoy/backend dev-bootstrap

# Re-sync NewAPI channel abilities
"${NEWAPI_SCRIPTS}/setup-dev-mock-channel.sh" || true

echo ""
echo "Reset complete (mode=${MODE})."
echo "Next: pnpm start ${MODE}"
