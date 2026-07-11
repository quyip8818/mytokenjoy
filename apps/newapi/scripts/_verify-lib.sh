# Shared helpers for verify:gate and verify:integration.
# shellcheck shell=bash

VERIFY_SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFY_ROOT="$(cd "${VERIFY_SCRIPTS_DIR}/../../.." && pwd)"
VERIFY_COMPOSE_FILE="${VERIFY_ROOT}/apps/newapi/docker-compose.yml"

API_URL="${API_URL:-http://localhost:8080}"
NEWAPI_URL="${NEWAPI_URL:-http://localhost:3000}"
WEBHOOK_SECRET="${NEW_API_WEBHOOK_SECRET:-tokenjoy-webhook-secret}"
NEW_API_ADMIN_TOKEN="${NEW_API_ADMIN_TOKEN:-}"
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
  body=$(printf '{"name":"%s","memberId":"m-1","budget":100,"modelWhitelist":[1]}' "${name}")
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
  VERIFY_NEWAPI_KEY_ID=$(verify_mapping_newapi_key_id "${VERIFY_PLATFORM_KEY_ID}")
  if [[ -z "${VERIFY_NEWAPI_KEY_ID}" ]]; then
    verify_fail "no newapi_key_id mapping for ${VERIFY_PLATFORM_KEY_ID}"
  fi
  verify_info "Created key ${VERIFY_PLATFORM_KEY_ID} (newapi_key_id=${VERIFY_NEWAPI_KEY_ID})"
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
