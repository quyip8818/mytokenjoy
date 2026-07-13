#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
verify_load_backend_dotenv

MODEL="${DEV_MOCK_MODEL:-local-test-model}"
BASE_URL="${DEV_MOCK_BASE_URL:-http://host.docker.internal:8765/v1}"
GROUP="${DEV_MOCK_CHANNEL_GROUP:-dept-dept-3}"
NAME="${DEV_MOCK_CHANNEL_NAME:-local-test-model}"

verify_info "local-test-model channel → ${BASE_URL} (group=${GROUP})"

if [[ -z "${NEW_API_ADMIN_TOKEN}" ]]; then
  cat <<EOF

NEW_API_ADMIN_TOKEN unset — create channel manually:

1. pnpm start:dev-mock
2. NewAPI Admin → Channels → Add
   - Type: OpenAI · Name: ${NAME} · Models: ${MODEL}
   - Base URL: ${BASE_URL} · Key: sk-local-test · Group: ${GROUP}
3. Abilities → Sync

EOF
  exit 0
fi

verify_wait_newapi

verify_ensure_newapi_group "${GROUP}" "后端组"

list_resp="$(mktemp)"
resp="$(mktemp)"
trap 'rm -f "${list_resp}" "${resp}"' EXIT

list_code=$(curl -s -o "${list_resp}" -w "%{http_code}" \
  -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
  -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
  "${NEWAPI_URL}/api/channel/?p=0&size=200")
[[ "${list_code}" == "200" ]] || verify_fail "list channels HTTP ${list_code}: $(cat "${list_resp}")"

existing=$(python3 - "${list_resp}" "${NAME}" "${MODEL}" <<'PY'
import json
import sys

path, name, model = sys.argv[1], sys.argv[2], sys.argv[3]
data = json.load(open(path, encoding="utf-8"))
items = data.get("data", {}).get("items") or data.get("data") or []
if isinstance(items, dict):
    items = items.get("items") or []
for item in items:
    if item.get("name") == name and model in (item.get("models") or ""):
        print("yes")
        break
PY
)

if [[ "${existing}" == "yes" ]]; then
  verify_info "channel already exists (${NAME}) — skipping create"
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
    "key": "sk-local-test",
    "base_url": "${BASE_URL}",
    "models": "${MODEL}",
    "group": "${GROUP}",
    "status": 1
  }
}
EOF
)")
  if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
    verify_fail "create channel HTTP ${code}: $(cat "${resp}")"
  fi
fi

sync_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -X GET "${NEWAPI_URL}/api/channel/sync" \
  -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
  -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}")
[[ "${sync_code}" == "200" ]] || verify_fail "channel sync HTTP ${sync_code}"

verify_info "local-test-model channel ready"
