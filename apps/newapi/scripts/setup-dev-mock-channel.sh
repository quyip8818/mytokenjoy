#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
verify_load_backend_dotenv

# ─── test-model channel (dev-mock upstream) ───────────────────────────────────

MODEL="${DEV_MOCK_MODEL:-test-model}"
BASE_URL="$(verify_http_origin "${DEV_MOCK_BASE_URL:-http://host.docker.internal:8765}")"
GROUP="${DEV_MOCK_CHANNEL_GROUP:-platform_shared}"
NAME="${DEV_MOCK_CHANNEL_NAME:-test-model}"
CHANNEL_KEY="${DEV_MOCK_CHANNEL_KEY:-sk-local-test}"

# NewAPI ChannelSettings (`setting` column). Pass-through keeps TokenJoy
# `dev_usage` in the upstream body so dev-mock-llm can echo usage.
CHANNEL_SETTING_JSON='{"pass_through_body_enabled":true}'

verify_info "test-model channel → ${BASE_URL} (group=${GROUP}, pass_through_body=on)"

if [[ -z "${NEW_API_ADMIN_TOKEN}" ]]; then
  cat <<EOF

NEW_API_ADMIN_TOKEN unset — create channel manually:

1. pnpm start:dev-mock
2. NewAPI Admin → Channels → Add
   - Type: OpenAI · Name: ${NAME} · Models: ${MODEL}
   - Base URL: ${BASE_URL} · Key: ${CHANNEL_KEY} · Group: ${GROUP}
   - Extra: enable Pass Through Body (pass_through_body_enabled)
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

channel_payload() {
  local mode="$1"
  python3 - "${mode}" "${existing_id:-}" "${NAME}" "${CHANNEL_KEY}" "${BASE_URL}" "${MODEL}" "${GROUP}" "${CHANNEL_SETTING_JSON}" <<'PY'
import json
import sys

mode, existing_id, name, key, base_url, model, group, setting = sys.argv[1:9]
channel = {
    "type": 1,
    "name": name,
    "key": key,
    "base_url": base_url,
    "models": model,
    "group": group,
    "weight": 1,
    "priority": 0,
    "setting": setting,
}
if mode == "update":
    channel["id"] = int(existing_id)
    print(json.dumps(channel))
else:
    channel["status"] = 1
    print(json.dumps({"mode": "single", "channel": channel}))
PY
}

if [[ -n "${existing_id}" ]]; then
  # NewAPI UpdateChannel rejects bodies that include "status".
  code=$(curl -s -o "${resp}" -w "%{http_code}" \
    -X PUT "${NEWAPI_URL}/api/channel/" \
    -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
    -H "Content-Type: application/json" \
    -d "$(channel_payload update)")
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
    -d "$(channel_payload create)")
  if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
    verify_fail "create channel HTTP ${code}: $(cat "${resp}")"
  fi
  verify_info "created channel ${NAME}"
fi

# ─── DeepSeek channel (production upstream) ───────────────────────────────────

DS_CHANNEL_NAME="Deepseek"
DS_CHANNEL_TYPE=43  # DeepSeek official
DS_CHANNEL_MODELS="deepseek-v4-flash,deepseek-v4-pro"
DS_CHANNEL_GROUP="default"
DS_CHANNEL_KEY="${DEEPSEEK_API_KEY:-sk-f0463e3791b741aca89144cf78106da4}"

verify_info "DeepSeek channel → models=${DS_CHANNEL_MODELS} (group=${DS_CHANNEL_GROUP})"

# Ensure pricing ratios for deepseek models
verify_ensure_local_test_model_pricing "deepseek-v4-pro" "0.2175" "1"
verify_ensure_local_test_model_pricing "deepseek-v4-flash" "0.0725" "1"

ds_existing_id=$(python3 - "${list_resp}" "${DS_CHANNEL_NAME}" "deepseek-v4" <<'PY'
import json
import sys

path, name, model_prefix = sys.argv[1], sys.argv[2], sys.argv[3]
data = json.load(open(path, encoding="utf-8"))
items = data.get("data", {}).get("items") or data.get("data") or []
if isinstance(items, dict):
    items = items.get("items") or []
for item in items:
    if item.get("name") == name and model_prefix in (item.get("models") or ""):
        print(item.get("id") or "")
        break
PY
)

ds_channel_payload() {
  local mode="$1"
  python3 - "${mode}" "${ds_existing_id:-}" "${DS_CHANNEL_NAME}" "${DS_CHANNEL_KEY}" "" "${DS_CHANNEL_MODELS}" "${DS_CHANNEL_GROUP}" "${DS_CHANNEL_TYPE}" <<'PY'
import json
import sys

mode, existing_id, name, key, base_url, models, group, ch_type = sys.argv[1:9]
channel = {
    "type": int(ch_type),
    "name": name,
    "key": key,
    "base_url": base_url,
    "models": models,
    "group": group,
    "weight": 1,
    "priority": 0,
}
if mode == "update":
    channel["id"] = int(existing_id)
    print(json.dumps(channel))
else:
    channel["status"] = 1
    print(json.dumps({"mode": "single", "channel": channel}))
PY
}

if [[ -n "${ds_existing_id}" ]]; then
  code=$(curl -s -o "${resp}" -w "%{http_code}" \
    -X PUT "${NEWAPI_URL}/api/channel/" \
    -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
    -H "Content-Type: application/json" \
    -d "$(ds_channel_payload update)")
  if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
    verify_fail "update DeepSeek channel HTTP ${code}: $(cat "${resp}")"
  fi
  verify_info "updated channel ${DS_CHANNEL_NAME} (id=${ds_existing_id})"
else
  code=$(curl -s -o "${resp}" -w "%{http_code}" \
    -X POST "${NEWAPI_URL}/api/channel/" \
    -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
    -H "Content-Type: application/json" \
    -d "$(ds_channel_payload create)")
  if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
    verify_fail "create DeepSeek channel HTTP ${code}: $(cat "${resp}")"
  fi
  verify_info "created channel ${DS_CHANNEL_NAME}"
fi

# ─── SiliconFlow channel ─────────────────────────────────────────────────────

SF_CHANNEL_NAME="siliconFlow"
SF_CHANNEL_TYPE=40  # SiliconFlow
SF_CHANNEL_MODELS="zai-org/GLM-5.2,moonshotai/Kimi-K2.7-Code"
SF_CHANNEL_GROUP="default"
SF_CHANNEL_KEY="${SILICONFLOW_API_KEY:-}"

if [[ -z "${SF_CHANNEL_KEY}" ]]; then
  verify_info "SKIP: SILICONFLOW_API_KEY unset — SiliconFlow channel not created"
else
  verify_info "SiliconFlow channel → models=${SF_CHANNEL_MODELS} (group=${SF_CHANNEL_GROUP})"

  # Pricing: GLM-5.2 $1.302/$4.092, Kimi-K2.7-Code $0.85916/$3.80 per 1M tokens
  verify_ensure_local_test_model_pricing "zai-org/GLM-5.2" "0.651" "3.143"
  verify_ensure_local_test_model_pricing "moonshotai/Kimi-K2.7-Code" "0.42958" "4.423"

  sf_existing_id=$(python3 - "${list_resp}" "${SF_CHANNEL_NAME}" "zai-org/GLM-5.2" <<'PY'
import json
import sys

path, name, model_prefix = sys.argv[1], sys.argv[2], sys.argv[3]
data = json.load(open(path, encoding="utf-8"))
items = data.get("data", {}).get("items") or data.get("data") or []
if isinstance(items, dict):
    items = items.get("items") or []
for item in items:
    if item.get("name") == name and model_prefix in (item.get("models") or ""):
        print(item.get("id") or "")
        break
PY
  )

  sf_channel_payload() {
    local mode="$1"
    python3 - "${mode}" "${sf_existing_id:-}" "${SF_CHANNEL_NAME}" "${SF_CHANNEL_KEY}" "" "${SF_CHANNEL_MODELS}" "${SF_CHANNEL_GROUP}" "${SF_CHANNEL_TYPE}" <<'PY'
import json
import sys

mode, existing_id, name, key, base_url, models, group, ch_type = sys.argv[1:9]
channel = {
    "type": int(ch_type),
    "name": name,
    "key": key,
    "base_url": base_url,
    "models": models,
    "group": group,
    "weight": 1,
    "priority": 0,
}
if mode == "update":
    channel["id"] = int(existing_id)
    print(json.dumps(channel))
else:
    channel["status"] = 1
    print(json.dumps({"mode": "single", "channel": channel}))
PY
  }

  if [[ -n "${sf_existing_id}" ]]; then
    code=$(curl -s -o "${resp}" -w "%{http_code}" \
      -X PUT "${NEWAPI_URL}/api/channel/" \
      -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
      -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
      -H "Content-Type: application/json" \
      -d "$(sf_channel_payload update)")
    if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
      verify_fail "update SiliconFlow channel HTTP ${code}: $(cat "${resp}")"
    fi
    verify_info "updated channel ${SF_CHANNEL_NAME} (id=${sf_existing_id})"
  else
    code=$(curl -s -o "${resp}" -w "%{http_code}" \
      -X POST "${NEWAPI_URL}/api/channel/" \
      -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
      -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
      -H "Content-Type: application/json" \
      -d "$(sf_channel_payload create)")
    if [[ "${code}" != "200" ]] || [[ "$(verify_json_success "${resp}")" != "yes" ]]; then
      verify_fail "create SiliconFlow channel HTTP ${code}: $(cat "${resp}")"
    fi
    verify_info "created channel ${SF_CHANNEL_NAME}"
  fi
fi

# ─── Sync abilities ───────────────────────────────────────────────────────────

sync_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -X GET "${NEWAPI_URL}/api/channel/sync" \
  -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
  -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}")
[[ "${sync_code}" == "200" ]] || verify_fail "channel sync HTTP ${sync_code}"

verify_info "all channels ready"
