#!/usr/bin/env bash
set -euo pipefail

# Post docker:reset: start infra, ensure NewAPI root account, optional dev-mock channel.
# Admin token is no longer written to .env — Backend reads it directly from NewAPI's database.

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
trap verify_cleanup EXIT

SKIP_CHANNEL=false
for arg in "$@"; do
  case "${arg}" in
    --skip-channel) SKIP_CHANNEL=true ;;
    --token-only)
      echo "DEPRECATED: --token-only is no longer needed. Backend reads token directly from NewAPI DB."
      exit 0
      ;;
    -h|--help)
      cat <<EOF
Usage: bootstrap-local-after-reset.sh [--skip-channel]

Run after pnpm docker:reset (or whenever NewAPI Postgres volume was wiped):
  default          start-infra + ensure root account + dev-mock channel (best-effort)
  --skip-channel   Skip local-test-model channel setup

Then: pnpm start
EOF
      exit 0
      ;;
  esac
done

echo "== TokenJoy bootstrap-local-after-reset =="

verify_require_tools

"${VERIFY_SCRIPTS_DIR}/start-infra.sh"

# Ensure root account exists (needed for NewAPI admin panel access).
verify_wait_newapi
verify_newapi_ensure_root

# Mint token into env for verify/channel scripts that still need it.
verify_newapi_mint_admin_token

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
echo "  Backend reads admin token directly from NewAPI database — no .env needed."
echo "Next: pnpm start"
