#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
verify_load_backend_dotenv

MODEL="${DEV_MOCK_MODEL:-local-test-model}"
BASE_URL="$(verify_http_origin "${DEV_MOCK_BASE_URL:-http://host.docker.internal:8765}")"
GROUP="${DEV_MOCK_CHANNEL_GROUP:-dept-dept-3}"
NAME="${DEV_MOCK_CHANNEL_NAME:-local-test-model}"
CHANNEL_KEY="${DEV_MOCK_CHANNEL_KEY:-sk-local-test}"

verify_info "local-test-model channel → ${BASE_URL} (group=${GROUP})"

if [[ -z "${NEW_API_ADMIN_TOKEN}" ]]; then
  cat <<EOF

NEW_API_ADMIN_TOKEN unset — create channel manually:

1. pnpm start:dev-mock
2. NewAPI Admin → Channels → Add
   - Type: OpenAI · Name: ${NAME} · Models: ${MODEL}
   - Base URL: ${BASE_URL} · Key: ${CHANNEL_KEY} · Group: ${GROUP}
3. System Settings → Group & Model Pricing → set ModelRatio for ${MODEL}
4. Abilities → Sync

EOF
  exit 0
fi

verify_wait_newapi

verify_ensure_newapi_group "${GROUP}" "后端组"
verify_ensure_local_test_model_pricing "${MODEL}"

list_resp="$(mktemp)"
resp="$(mktemp)"
trap 'rm -f "${list_resp}" "${resp}"' EXIT

list_code=$(curl -s -o "${list_resp}" -w "%{http_code}" \
  -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
  -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
  "${NEWAPI_URL}/api/channel/?p=0&size=200")
[[ "${list_code}" == "200" ]] || verify_fail "list channels HTTP ${list_code}: $(cat "${list_resp}")"

existing_id=$(python3 - "${list_resp}" "${NAME}" "${MODEL}" <<'PY'
import json
import sys

path, name, model = sys.argv[1], sys.argv[2], sys.argv[3]
data = json.load(open(path, encoding="utf-8"))
items = data.get("data", {}).get("items") or data.get("data") or []
if isinstance(items, dict):
    items = items.get("items") or []
for item in items:
    if item.get("name") == name and model in (item.get("models") or ""):
        print(item.get("id") or "")
        break
PY
)

if [[ -n "${existing_id}" ]]; then
  # NewAPI UpdateChannel rejects bodies that include "status".
  code=$(curl -s -o "${resp}" -w "%{http_code}" \
    -X PUT "${NEWAPI_URL}/api/channel/" \
    -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
    -H "Content-Type: application/json" \
    -d "$(cat <<EOF
{
  "id": ${existing_id},
  "type": 1,
  "name": "${NAME}",
  "key": "${CHANNEL_KEY}",
  "base_url": "${BASE_URL}",
  "models": "${MODEL}",
  "group": "${GROUP}",
  "weight": 1,
  "priority": 0
}
EOF
)")
  if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
    verify_fail "update channel HTTP ${code}: $(cat "${resp}")"
  fi
  verify_info "updated channel ${NAME} (id=${existing_id})"
else
  code=$(curl -s -o "${resp}" -w "%{http_code}" \
    -X POST "${NEWAPI_URL}/api/channel/" \
    -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
    -H "Content-Type: application/json" \
    -d "$(cat <<EOF
{
  "mode": "single",
  "channel": {
    "type": 1,
    "name": "${NAME}",
    "key": "${CHANNEL_KEY}",
    "base_url": "${BASE_URL}",
    "models": "${MODEL}",
    "group": "${GROUP}",
    "status": 1,
    "weight": 1,
    "priority": 0
  }
}
EOF
)")
  if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
    verify_fail "create channel HTTP ${code}: $(cat "${resp}")"
  fi
  verify_info "created channel ${NAME}"
fi

sync_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -X GET "${NEWAPI_URL}/api/channel/sync" \
  -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
  -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}")
[[ "${sync_code}" == "200" ]] || verify_fail "channel sync HTTP ${sync_code}"

verify_info "local-test-model channel ready"
