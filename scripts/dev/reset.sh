#!/usr/bin/env bash
# Reset: wipe volumes, rebuild infra, seed data.
# Usage: pnpm reset [local|saas] [--empty|--minimal|--full]
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

MODE=local
SEED=full
for arg in "$@"; do
  case "${arg}" in
    local|saas) MODE="${arg}" ;;
    --empty) SEED=empty ;;
    --minimal) SEED=minimal ;;
    --full) SEED=full ;;
    -h|--help)
      echo "usage: pnpm reset [local|saas] [--empty|--minimal|--full]"
      echo "  local (default): single company, seed controlled by --empty/--minimal/--full"
      echo "  saas: multi-tenant platform + demo company (always full seed)"
      exit 0
      ;;
    *) echo "unknown arg: ${arg}" >&2; exit 1 ;;
  esac
done

# saas mode: always full seed (platform admin + demo company)
if [[ "${MODE}" == "saas" ]]; then
  SEED=full
fi

SUPPORT_SAAS=$([[ "${MODE}" == "saas" ]] && echo true || echo false)

# --- Update .env.development (only modify existing keys) ---
ENV_FILE="${ROOT}/apps/backend/.env.development"

set_env() {
  local key="$1" value="$2"
  if grep -q "^${key}=" "${ENV_FILE}"; then
    sed -i '' "s|^${key}=.*|${key}=${value}|" "${ENV_FILE}"
  fi
}

set_env SUPPORT_SAAS "${SUPPORT_SAAS}"

# --- Wipe & rebuild ---
"${COMPOSE[@]}" down -v
"${COMPOSE[@]}" up postgres -d --wait
"${NEWAPI_SCRIPTS}/bootstrap-local-after-reset.sh"
"${COMPOSE[@]}" exec -T redis redis-cli FLUSHALL

# Seed data (SEED env is read by Go config, not persisted)
if [[ "${MODE}" == "saas" ]]; then
  SEED="${SEED}" \
  PLATFORM_BOOTSTRAP_EMAIL="${PLATFORM_BOOTSTRAP_EMAIL:-ops@tokenjoy.local}" \
  PLATFORM_BOOTSTRAP_PASSWORD="${PLATFORM_BOOTSTRAP_PASSWORD:-platform-dev-123}" \
    pnpm -F @tokenjoy/backend dev-bootstrap
else
  SEED="${SEED}" pnpm -F @tokenjoy/backend dev-bootstrap
fi

# Re-sync NewAPI channel abilities after dev-bootstrap to ensure test-model is routable.
"${NEWAPI_SCRIPTS}/setup-dev-mock-channel.sh" || true

echo ""
echo "Reset complete (mode=${MODE}, seed=${SEED})."
echo "Next: pnpm start"
