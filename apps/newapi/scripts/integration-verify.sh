#!/usr/bin/env bash
set -euo pipefail

VERIFY_NO_BUILD=false
for arg in "$@"; do
  case "${arg}" in
    --no-build) VERIFY_NO_BUILD=true ;;
    -h|--help)
      echo "Usage: integration-verify.sh [--no-build]"
      echo "Full stack: gate flow + ledger + Toggle/Rotate/Revoke + ingest metrics."
      echo "Requires Backend full-stack .env and NEW_API_ADMIN_TOKEN."
      exit 0
      ;;
  esac
done

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
trap verify_cleanup EXIT
verify_load_backend_dotenv

echo "== TokenJoy verify:integration =="
echo "Backend: ${API_URL} | NewAPI: ${NEWAPI_URL}"

if [[ -z "${NEW_API_ADMIN_TOKEN}" ]]; then
  verify_fail "NEW_API_ADMIN_TOKEN is required (NewAPI Admin GET for Rotate/Revoke assertions)"
fi

if ! command -v psql >/dev/null 2>&1; then
  verify_fail "psql not found (required for ledger assertions)"
fi

verify_run_gate_flow

KEY_ID="${VERIFY_PLATFORM_KEY_ID}"
KEY_BEARER="${VERIFY_PLATFORM_KEY_BEARER}"
KEY_NAID="${VERIFY_NEWAPI_KEY_ID}"

verify_info "Ledger ingest..."
verify_inject_consume_log_and_ingest "${KEY_NAID}"
verify_three_axis_consumed_after_ingest "${KEY_ID}"

verify_info "Toggle off..."
resp="${VERIFY_TMPDIR}/toggle-off.json"
code=$(verify_api_call PUT "/api/keys/platform/${KEY_ID}/toggle" '{"enabled":false}' "${resp}")
[[ "${code}" == "200" ]] || verify_fail "toggle off HTTP ${code}: $(cat "${resp}")"
verify_assert_gateway_forbidden "toggle off" "$(verify_gateway_code "${KEY_BEARER}")"

verify_info "Toggle on..."
resp="${VERIFY_TMPDIR}/toggle-on.json"
code=$(verify_api_call PUT "/api/keys/platform/${KEY_ID}/toggle" '{"enabled":true}' "${resp}")
[[ "${code}" == "200" ]] || verify_fail "toggle on HTTP ${code}: $(cat "${resp}")"
verify_assert_gateway_ok "toggle on" "$(verify_gateway_code "${KEY_BEARER}")"

verify_info "Rotate..."
OLD_BEARER="${KEY_BEARER}"
resp="${VERIFY_TMPDIR}/rotate.json"
code=$(verify_api_call POST "/api/keys/platform/${KEY_ID}/rotate" "" "${resp}")
[[ "${code}" == "200" ]] || verify_fail "rotate HTTP ${code}: $(cat "${resp}")"
NEW_BEARER=$(verify_json_field "${resp}" "fullKey")
if [[ -z "${NEW_BEARER}" ]]; then
  verify_fail "rotate missing fullKey: $(cat "${resp}")"
fi
verify_assert_gateway_forbidden "rotate old bearer" "$(verify_gateway_code "${OLD_BEARER}")"
verify_assert_gateway_ok "rotate new bearer" "$(verify_gateway_code "${NEW_BEARER}")"
KEY_NAID_AFTER=$(verify_mapping_newapi_key_id "${KEY_ID}")
if [[ "${KEY_NAID}" != "${KEY_NAID_AFTER}" ]]; then
  verify_fail "newapi_key_id changed on rotate: ${KEY_NAID} -> ${KEY_NAID_AFTER}"
fi
na_code=$(verify_newapi_token_http_code "${KEY_NAID}")
[[ "${na_code}" == "200" ]] || verify_fail "NewAPI token ${KEY_NAID} not found after rotate (HTTP ${na_code})"
verify_info "Rotate OK: newapi_key_id unchanged (${KEY_NAID})"

verify_info "Revoke..."
verify_create_platform_key "revoke-${VERIFY_RUN_TS}"
REVOKE_ID="${VERIFY_PLATFORM_KEY_ID}"
REVOKE_BEARER="${VERIFY_PLATFORM_KEY_BEARER}"
REVOKE_NAID="${VERIFY_NEWAPI_KEY_ID}"
verify_assert_gateway_ok "revoke-target active" "$(verify_gateway_code "${REVOKE_BEARER}")"

resp="${VERIFY_TMPDIR}/revoke.json"
code=$(verify_api_call PUT "/api/keys/platform/${REVOKE_ID}/revoke" "" "${resp}")
[[ "${code}" == "204" || "${code}" == "200" ]] || verify_fail "revoke HTTP ${code}: $(cat "${resp}")"
verify_assert_gateway_forbidden "revoke" "$(verify_gateway_code "${REVOKE_BEARER}")"
status=$(verify_platform_key_status "${REVOKE_ID}")
[[ "${status}" == "revoked" ]] || verify_fail "expected platform key status revoked, got '${status}'"
na_code=$(verify_newapi_token_http_code "${REVOKE_NAID}")
if [[ "${na_code}" == "200" ]]; then
  verify_fail "NewAPI token ${REVOKE_NAID} still exists after revoke"
fi
verify_info "Revoke OK: DB revoked, NewAPI token gone (HTTP ${na_code})"

verify_info "Ingest metrics..."
metrics_resp="${VERIFY_TMPDIR}/metrics.json"
metrics_code=$(curl -s -o "${metrics_resp}" -w "%{http_code}" \
  -H "X-Webhook-Secret: ${WEBHOOK_SECRET}" \
  "${API_URL}/api/internal/metrics/ingest")
[[ "${metrics_code}" == "200" ]] || verify_fail "ingest metrics HTTP ${metrics_code}"
pending=$(verify_json_field "${metrics_resp}" "ingest_jobs_pending")
verify_info "ingest_jobs_pending=${pending}"
if [[ "${pending}" != "0" && "${pending}" != "" ]]; then
  verify_info "warn: ingest_jobs_pending=${pending} (non-zero; check dead jobs if persistent)"
fi

echo ""
echo "INTEGRATION VERIFY PASSED"
  echo "  Gate flow (disposable key + Gateway + webhook + soft remain)"
  echo "  Ledger ingest (three-axis consumed)"
echo "  Toggle off/on"
echo "  Rotate (old 403, new OK, newapi_key_id unchanged)"
echo "  Revoke (403, DB revoked, NewAPI token gone)"
echo "  Ingest metrics"
