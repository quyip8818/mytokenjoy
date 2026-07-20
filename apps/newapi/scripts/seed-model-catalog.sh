#!/usr/bin/env bash
set -euo pipefail

# Seed model pricing catalog into NewAPI (idempotent).
# Can be run standalone or as part of bootstrap-local-after-reset.sh.

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
verify_load_backend_dotenv

if [[ -z "${NEW_API_ADMIN_TOKEN}" ]]; then
  verify_info "SKIP: NEW_API_ADMIN_TOKEN unset — model catalog not seeded"
  exit 0
fi

verify_wait_newapi

CATALOG="${VERIFY_SCRIPTS_DIR}/lib/model-catalog.json"
if [[ ! -f "${CATALOG}" ]]; then
  verify_fail "model-catalog.json not found: ${CATALOG}"
fi

verify_info "Seeding model pricing catalog..."
python3 "${VERIFY_SCRIPTS_DIR}/lib/newapi_admin.py" seed-model-catalog "${CATALOG}"
verify_info "Model catalog seed done."
