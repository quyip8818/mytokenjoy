#!/usr/bin/env bash
set -euo pipefail

# Post docker:reset: start infra, mint NewAPI admin token → apps/backend/.env, optional dev-mock channel.

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
trap verify_cleanup EXIT

SKIP_CHANNEL=false
TOKEN_ONLY=false
for arg in "$@"; do
  case "${arg}" in
    --skip-channel) SKIP_CHANNEL=true ;;
    --token-only) TOKEN_ONLY=true ;;
    -h|--help)
      cat <<EOF
Usage: bootstrap-local-after-reset.sh [--token-only] [--skip-channel]

Run after pnpm docker:reset (or whenever NewAPI Postgres volume was wiped):
  --token-only     Mint admin token → apps/backend/.env only (NewAPI must already be up)
  default          start-infra + token + dev-mock channel (best-effort)
  --skip-channel   Skip local-test-model channel setup

Then: pnpm start
EOF
      exit 0
      ;;
  esac
done

if [[ "${TOKEN_ONLY}" == "true" ]]; then
  echo "== TokenJoy bootstrap:token =="
  verify_require_tools
  verify_bootstrap_newapi_admin_token
  echo ""
  echo "OK: NEW_API_ADMIN_TOKEN → ${BACKEND_ENV_FILE}"
  echo "Restart Backend if it is already running."
  exit 0
fi

echo "== TokenJoy bootstrap-local-after-reset =="

verify_require_tools

"${VERIFY_SCRIPTS_DIR}/start-infra.sh"

verify_bootstrap_newapi_admin_token

if [[ "${SKIP_CHANNEL}" == "true" ]]; then
  verify_info "Skipping dev-mock channel (--skip-channel)"
else
  verify_info "Configuring local-test-model channel (best-effort)..."
  if ! "${VERIFY_SCRIPTS_DIR}/setup-dev-mock-channel.sh"; then
    verify_info "WARN: dev-mock channel setup failed — configure manually in NewAPI Admin if needed"
  fi
fi

echo ""
echo "Bootstrap complete."
echo "  NEW_API_ADMIN_TOKEN → ${BACKEND_ENV_FILE}"
echo "Next: pnpm start   (L1 platform key sync already done by pnpm docker:reset)"
echo "      If Backend is already running, restart it to load the new token."
