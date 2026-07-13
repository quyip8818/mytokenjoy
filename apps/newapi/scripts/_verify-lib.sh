# Shared helpers for verify:gate and verify:integration.
# shellcheck shell=bash

VERIFY_SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFY_ROOT="$(cd "${VERIFY_SCRIPTS_DIR}/../../.." && pwd)"
VERIFY_COMPOSE_FILE="${VERIFY_ROOT}/apps/newapi/docker-compose.yml"

API_URL="${API_URL:-http://localhost:8080}"
NEWAPI_URL="${NEWAPI_URL:-http://localhost:3000}"
WEBHOOK_SECRET="${NEW_API_WEBHOOK_SECRET:-tokenjoy-webhook-secret}"
NEW_API_ADMIN_TOKEN="${NEW_API_ADMIN_TOKEN:-}"
BACKEND_ENV_FILE="${BACKEND_ENV_FILE:-${VERIFY_ROOT}/apps/backend/.env}"
NEW_API_ROOT_USERNAME="${NEW_API_ROOT_USERNAME:-root}"
NEW_API_ROOT_PASSWORD="${NEW_API_ROOT_PASSWORD:-tokenjoy123}"
NEW_API_ADMIN_USER_ID="${NEW_API_ADMIN_USER_ID:-1}"
export NEW_API_ROOT_USERNAME NEW_API_ROOT_PASSWORD
DATABASE_URL="${DATABASE_URL:-postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable}"
LOG_DATABASE_URL="${LOG_DATABASE_URL:-postgres://tokenjoy:tokenjoy@127.0.0.1:5432/logs?sslmode=disable}"
WORKER_WAIT_SEC="${WORKER_WAIT_SEC:-8}"
VERIFY_DEFAULT_COMPANY_ID="${VERIFY_DEFAULT_COMPANY_ID:-2}"

VERIFY_RUN_TS="${VERIFY_RUN_TS:-$(date +%s)}"
VERIFY_COOKIE_JAR="${VERIFY_COOKIE_JAR:-$(mktemp)}"
VERIFY_TMPDIR="${VERIFY_TMPDIR:-$(mktemp -d)}"
VERIFY_PLATFORM_KEY_ID=""
VERIFY_PLATFORM_KEY_BEARER=""
VERIFY_NEWAPI_KEY_ID=""

verify_cleanup() {
  rm -f "${VERIFY_COOKIE_JAR}"
  rm -rf "${VERIFY_TMPDIR}"
}

verify_fail() {
  echo "FAIL: $*" >&2
  exit 1
}

verify_info() {
  echo "$*"
}

verify_json_field() {
  local file="$1"
  local field="$2"
  python3 -c "import json,sys; d=json.load(open(sys.argv[1])); v=d;
for part in sys.argv[2].split('.'):
  v = v.get(part) if isinstance(v, dict) else None
  if v is None: break
print('' if v is None else v)" "${file}" "${field}"
}

verify_json_success() {
  local file="$1"
  python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); print("yes" if d.get("success") else "no")' "${file}"
}

verify_require_tools() {
  if ! docker info >/dev/null 2>&1; then
    verify_fail "Docker unavailable"
  fi
  if ! command -v python3 >/dev/null 2>&1; then
    verify_fail "python3 not found (required for JSON parsing)"
  fi
}

verify_start_stack() {
  local no_build="${1:-false}"
  verify_info "Starting postgres + newapi..."
  if [[ "${no_build}" == "true" ]]; then
    docker compose -f "${VERIFY_COMPOSE_FILE}" up -d postgres redis new-api
  else
    docker compose -f "${VERIFY_COMPOSE_FILE}" up -d --build postgres redis new-api
  fi
}

verify_wait_newapi() {
  verify_info "Waiting for NewAPI..."
  for _ in $(seq 1 60); do
    if curl -fsS "${NEWAPI_URL}/api/status" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  curl -fsS "${NEWAPI_URL}/api/status" >/dev/null || verify_fail "NewAPI /api/status unreachable"
}

# Register a TokenJoy department group in NewAPI options so tokens/channels can use it.
verify_ensure_newapi_group() {
  local group="$1"
  local label="${2:-$group}"
  if [[ -z "${NEW_API_ADMIN_TOKEN}" || -z "${group}" ]]; then
    return 0
  fi
  python3 - "${group}" "${label}" <<'PY'
import json
import os
import subprocess
import sys
import urllib.request

group, label = sys.argv[1], sys.argv[2]
base = os.environ.get("NEWAPI_URL", "http://localhost:3000").rstrip("/")
token = os.environ["NEW_API_ADMIN_TOKEN"]
user_id = os.environ.get("NEW_API_ADMIN_USER_ID", "1")

def request(method: str, path: str, body: dict | None = None) -> dict:
    data = None
    headers = {
        "Authorization": f"Bearer {token}",
        "New-Api-User": user_id,
    }
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(f"{base}{path}", data=data, headers=headers, method=method)
    with urllib.request.urlopen(req) as resp:
        return json.load(resp)

options = {item["key"]: item.get("value", "") for item in request("GET", "/api/option/").get("data", [])}

def merge_map(key: str, group: str, label: str) -> str | None:
    raw = options.get(key, "") or "{}"
    try:
        data = json.loads(raw) if raw else {}
    except json.JSONDecodeError:
        data = {}
    if not isinstance(data, dict) or group in data:
        return None
    data[group] = label if key == "UserUsableGroups" else 1
    return json.dumps(data, ensure_ascii=False)

updates = []
for key in ("UserUsableGroups", "GroupRatio"):
    merged = merge_map(key, group, label)
    if merged is not None:
        updates.append({"key": key, "value": merged})

if not updates:
    print(f"NewAPI group already registered: {group}")
    sys.exit(0)

for item in updates:
    request("PUT", "/api/option/", item)
print(f"Registered NewAPI group: {group}")
PY
}

verify_require_backend_health() {
  verify_info "Checking Backend health..."
  curl -fsS "${API_URL}/healthz" >/dev/null || verify_fail "Backend /healthz unreachable — start Backend with NEW_API_ENABLED + Gateway + LOG_DATABASE_URL"
}

verify_admin_login() {
  local resp="${VERIFY_TMPDIR}/login.json"
  local code
  code=$(curl -s -o "${resp}" -w "%{http_code}" -c "${VERIFY_COOKIE_JAR}" \
    -X POST "${API_URL}/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@example.com","password":"demo1234"}')
  if [[ "${code}" != "200" ]]; then
    verify_fail "admin login failed HTTP ${code}: $(cat "${resp}")"
  fi
  verify_info "Admin login OK"
}

verify_api_call() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local out="$4"
  local args=(-s -o "${out}" -w "%{http_code}" -b "${VERIFY_COOKIE_JAR}" -c "${VERIFY_COOKIE_JAR}" \
    -X "${method}" "${API_URL}${path}" -H "Content-Type: application/json")
  if [[ -n "${body}" ]]; then
    args+=(-d "${body}")
  fi
  curl "${args[@]}"
}

verify_create_platform_key() {
  local name="$1"
  local body resp code
  body=$(printf '{"name":"%s","scope":"member","memberId":"m-1","budget":100,"modelWhitelist":[100]}' "${name}")
  resp="${VERIFY_TMPDIR}/create-key-${name}.json"
  code=$(verify_api_call POST "/api/keys/platform" "${body}" "${resp}")
  if [[ "${code}" != "200" ]]; then
    verify_fail "create platform key HTTP ${code}: $(cat "${resp}")"
  fi
  VERIFY_PLATFORM_KEY_ID=$(verify_json_field "${resp}" "id")
  VERIFY_PLATFORM_KEY_BEARER=$(verify_json_field "${resp}" "fullKey")
  if [[ -z "${VERIFY_PLATFORM_KEY_ID}" || -z "${VERIFY_PLATFORM_KEY_BEARER}" ]]; then
    verify_fail "create key missing id/fullKey: $(cat "${resp}")"
  fi
  verify_platform_key_soft_remain "${VERIFY_PLATFORM_KEY_ID}"
  VERIFY_NEWAPI_KEY_ID=$(verify_mapping_newapi_key_id "${VERIFY_PLATFORM_KEY_ID}")
  if [[ -z "${VERIFY_NEWAPI_KEY_ID}" ]]; then
    verify_fail "no newapi_key_id mapping for ${VERIFY_PLATFORM_KEY_ID}"
  fi
  verify_info "Created key ${VERIFY_PLATFORM_KEY_ID} (newapi_key_id=${VERIFY_NEWAPI_KEY_ID})"
}

verify_platform_key_soft_remain() {
  local platform_key_id="$1"
  local remain
  remain=$(verify_psql_main "SELECT gateway_soft_remain FROM platform_keys WHERE id='${platform_key_id}' AND company_id=${VERIFY_DEFAULT_COMPANY_ID} LIMIT 1")
  if [[ -z "${remain}" ]]; then
    verify_fail "gateway_soft_remain NULL for ${platform_key_id}"
  fi
  verify_info "gateway_soft_remain=${remain} for ${platform_key_id}"
}

verify_three_axis_consumed_after_ingest() {
  local platform_key_id="$1"
  local pk_consumed member_consumed org_node_count
  pk_consumed=$(verify_psql_main "SELECT COALESCE(SUM(consumed),0) FROM budget_consumed WHERE axis_kind='platform_key' AND axis_id='${platform_key_id}' AND company_id=${VERIFY_DEFAULT_COMPANY_ID}")
  member_consumed=$(verify_psql_main "SELECT COALESCE(SUM(consumed),0) FROM budget_consumed WHERE axis_kind='member' AND axis_id='m-1' AND company_id=${VERIFY_DEFAULT_COMPANY_ID}")
  org_node_count=$(verify_psql_main "SELECT COUNT(*) FROM budget_consumed WHERE axis_kind='org_node' AND company_id=${VERIFY_DEFAULT_COMPANY_ID}")
  if [[ "${pk_consumed}" == "0" ]]; then
    verify_fail "expected platform_key consumed > 0 after ingest, got ${pk_consumed}"
  fi
  if [[ "${member_consumed}" == "0" ]]; then
    verify_fail "expected member consumed > 0 after ingest, got ${member_consumed}"
  fi
  verify_info "three-axis OK: platform_key=${pk_consumed} member=${member_consumed} (org_node rows=${org_node_count}, not written on ingest)"
}

verify_gateway_code() {
  local bearer="$1"
  curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${API_URL}/v1/chat/completions" \
    -H "Authorization: Bearer ${bearer}" \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"verify ping"}]}' || true
}

verify_assert_gateway_ok() {
  local label="$1"
  local code="$2"
  case "${code}" in
    200|403|422) verify_info "${label}: Gateway HTTP ${code} (OK)" ;;
    401) verify_fail "${label}: Unauthorized (401) — check bearer / mapping / NewAPI sync" ;;
    502|504) verify_fail "${label}: Gateway proxy error (${code})" ;;
    *) verify_fail "${label}: unexpected Gateway HTTP ${code}" ;;
  esac
}

verify_assert_gateway_forbidden() {
  local label="$1"
  local code="$2"
  if [[ "${code}" == "403" ]]; then
    verify_info "${label}: Gateway HTTP 403 (expected)"
  else
    verify_fail "${label}: expected 403, got ${code}"
  fi
}

verify_psql_main() {
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -tAc "$1"
}

verify_psql_logs() {
  psql "${LOG_DATABASE_URL}" -v ON_ERROR_STOP=1 -tAc "$1"
}

verify_mapping_newapi_key_id() {
  local platform_key_id="$1"
  verify_psql_main "SELECT newapi_key_id FROM platform_key_mappings WHERE platform_key_id='${platform_key_id}' AND company_id=${VERIFY_DEFAULT_COMPANY_ID} LIMIT 1"
}

verify_platform_key_status() {
  local platform_key_id="$1"
  verify_psql_main "SELECT status FROM platform_keys WHERE id='${platform_key_id}' AND company_id=${VERIFY_DEFAULT_COMPANY_ID} LIMIT 1"
}

verify_newapi_token_http_code() {
  local token_id="$1"
  curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID:-1}" \
    "${NEWAPI_URL}/api/token/${token_id}" || true
}

verify_post_webhook() {
  local log_id="$1"
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${API_URL}/api/internal/webhooks/newapi-log" \
    -H "Content-Type: application/json" \
    -H "X-Webhook-Secret: ${WEBHOOK_SECRET}" \
    -d "{\"log_id\": ${log_id}}" || true)
  if [[ "${code}" != "200" ]]; then
    verify_fail "webhook enqueue failed HTTP ${code} for log_id=${log_id}"
  fi
  verify_info "Webhook accepted log_id=${log_id}"
}

verify_inject_consume_log_and_ingest() {
  local token_id="$1"
  if ! command -v psql >/dev/null 2>&1; then
    verify_fail "psql not found (required for ledger assertions)"
  fi
  local log_id
  log_id=$(verify_psql_logs "INSERT INTO newapi.logs (user_id, created_at, type, token_id, model_name, quota, prompt_tokens, completion_tokens, use_time)
    VALUES (0, extract(epoch from now())::bigint, 2, ${token_id}, 'gpt-4o-mini', 50, 10, 5, 1)
    RETURNING id")
  log_id="$(echo "${log_id}" | tr -d '[:space:]')"
  verify_post_webhook "${log_id}"
  sleep "${WORKER_WAIT_SEC}"
  local exists
  exists=$(verify_psql_main "SELECT COUNT(*) FROM usage_ledger WHERE idempotency_key='newapi:${log_id}'")
  if [[ "${exists}" -lt 1 ]]; then
    verify_fail "ledger missing idempotency_key newapi:${log_id} after worker wait"
  fi
  verify_info "Ledger entry newapi:${log_id} verified"
}

verify_run_gate_flow() {
  verify_require_tools
  verify_start_stack "${VERIFY_NO_BUILD:-false}"
  verify_wait_newapi
  verify_require_backend_health
  verify_admin_login
  verify_create_platform_key "gate-${VERIFY_RUN_TS}"
  verify_assert_gateway_ok "gateway" "$(verify_gateway_code "${VERIFY_PLATFORM_KEY_BEARER}")"
  verify_post_webhook "900001"
}

# Load apps/backend/.env when NEW_API_ADMIN_TOKEN is unset (verify / channel scripts).
verify_load_backend_dotenv() {
  if [[ -n "${NEW_API_ADMIN_TOKEN}" || ! -f "${BACKEND_ENV_FILE}" ]]; then
    return 0
  fi
  # shellcheck disable=SC1090
  set -a && source "${BACKEND_ENV_FILE}" && set +a
}

verify_write_env_var() {
  local file="$1"
  local key="$2"
  local value="$3"
  python3 - "${file}" "${key}" "${value}" <<'PY'
import os
import sys

path, key, value = sys.argv[1], sys.argv[2], sys.argv[3]
lines: list[str] = []
found = False
if os.path.exists(path):
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            if line.startswith(f"{key}="):
                lines.append(f"{key}={value}\n")
                found = True
            else:
                lines.append(line)
if not found:
    if lines and not lines[-1].endswith("\n"):
        lines[-1] += "\n"
    lines.append(f"{key}={value}\n")
os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
with open(path, "w", encoding="utf-8") as fh:
    fh.writelines(lines)
PY
}

verify_newapi_root_login() {
  local resp code
  resp="${VERIFY_TMPDIR}/newapi-login.json"
  code=$(curl -s -o "${resp}" -w "%{http_code}" -c "${VERIFY_COOKIE_JAR}" -b "${VERIFY_COOKIE_JAR}" \
    -X POST "${NEWAPI_URL}/api/user/login" \
    -H "Content-Type: application/json" \
    -d "$(python3 -c 'import json,os; print(json.dumps({"username":os.environ["NEW_API_ROOT_USERNAME"],"password":os.environ["NEW_API_ROOT_PASSWORD"]}))')")
  if [[ "${code}" == "200" && "$(verify_json_success "${resp}")" == "yes" ]]; then
    return 0
  fi
  verify_info "NewAPI root login failed HTTP ${code}: $(cat "${resp}")"
  return 1
}

verify_newapi_run_setup() {
  local resp code
  resp="${VERIFY_TMPDIR}/newapi-setup.json"
  code=$(curl -s -o "${resp}" -w "%{http_code}" \
    -X POST "${NEWAPI_URL}/api/setup" \
    -H "Content-Type: application/json" \
    -d "$(python3 -c 'import json,os; u=os.environ["NEW_API_ROOT_USERNAME"]; p=os.environ["NEW_API_ROOT_PASSWORD"]; print(json.dumps({"username":u,"password":p,"confirmPassword":p}))')")
  if [[ "${code}" == "200" && "$(verify_json_success "${resp}")" == "yes" ]]; then
    verify_info "NewAPI root account created (${NEW_API_ROOT_USERNAME})"
    return 0
  fi
  verify_info "NewAPI setup HTTP ${code}: $(cat "${resp}")"
  return 0
}

verify_newapi_ensure_root() {
  if verify_newapi_root_login; then
    verify_info "NewAPI root login OK (${NEW_API_ROOT_USERNAME})"
    return 0
  fi

  verify_newapi_run_setup

  verify_newapi_root_login || verify_fail "NewAPI root login failed after setup — set NEW_API_ROOT_USERNAME/NEW_API_ROOT_PASSWORD to match your NewAPI root account"
  verify_info "NewAPI root login OK (${NEW_API_ROOT_USERNAME})"
}

verify_newapi_mint_admin_token() {
  local resp
  resp="${VERIFY_TMPDIR}/newapi-admin-token.json"
  local code
  code=$(curl -s -o "${resp}" -w "%{http_code}" -b "${VERIFY_COOKIE_JAR}" -c "${VERIFY_COOKIE_JAR}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID}" \
    "${NEWAPI_URL}/api/user/token")
  if [[ "${code}" != "200" ]]; then
    verify_fail "fetch NewAPI admin token HTTP ${code}: $(cat "${resp}")"
  fi
  NEW_API_ADMIN_TOKEN="$(verify_json_field "${resp}" "data")"
  if [[ -z "${NEW_API_ADMIN_TOKEN}" ]]; then
    verify_fail "empty admin token in response: $(cat "${resp}")"
  fi
  export NEW_API_ADMIN_TOKEN
}

verify_assert_newapi_admin_token() {
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer ${NEW_API_ADMIN_TOKEN}" \
    -H "New-Api-User: ${NEW_API_ADMIN_USER_ID}" \
    "${NEWAPI_URL}/api/token/?p=0&size=1")
  if [[ "${code}" != "200" ]]; then
    verify_fail "NEW_API_ADMIN_TOKEN rejected by NewAPI (HTTP ${code})"
  fi
  verify_info "NEW_API_ADMIN_TOKEN verified against NewAPI"
}

verify_bootstrap_newapi_admin_token() {
  verify_require_tools
  verify_wait_newapi
  verify_newapi_ensure_root
  verify_newapi_mint_admin_token
  verify_assert_newapi_admin_token
  verify_write_env_var "${BACKEND_ENV_FILE}" "NEW_API_ADMIN_TOKEN" "${NEW_API_ADMIN_TOKEN}"
  verify_info "Wrote NEW_API_ADMIN_TOKEN to ${BACKEND_ENV_FILE}"
}
