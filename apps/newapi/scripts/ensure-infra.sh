#!/usr/bin/env bash
# Daily dev: attach infra without building images. First-time setup: pnpm docker:reset or pnpm infra.
set -euo pipefail

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"

COMPOSE=(docker compose -f "${VERIFY_COMPOSE_FILE}")

verify_info "Ensuring postgres + redis + new-api (no build)..."
if ! "${COMPOSE[@]}" up postgres redis new-api -d --wait --no-build; then
  verify_fail "Infra not ready (missing image or containers). Run: pnpm docker:reset   or: pnpm infra"
fi

verify_info "NewAPI quick check (${NEWAPI_URL}/api/status)..."
for _ in $(seq 1 15); do
  if curl -fsS "${NEWAPI_URL}/api/status" >/dev/null 2>&1; then
    verify_info "Infra ready"
    exit 0
  fi
  sleep 1
done

verify_fail "NewAPI /api/status unreachable after attach. Try: pnpm infra   or: pnpm docker:reset"
