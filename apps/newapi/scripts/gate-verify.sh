#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
RELAY_URL="${RELAY_URL:-http://localhost:3000}"
API_URL="${API_URL:-http://localhost:8080}"
WEBHOOK_SECRET="${NEW_API_WEBHOOK_SECRET:-tokenjoy-webhook-secret}"

echo "== TokenJoy P0 Gate verification =="

if ! docker info >/dev/null 2>&1; then
  echo "Docker unavailable; start Docker Desktop then re-run."
  exit 1
fi

echo "[1/6] Starting relay stack..."
docker compose -f "${ROOT}/apps/newapi/docker-compose.yml" up -d --build

echo "[2/6] Waiting for relay..."
for i in $(seq 1 60); do
  if curl -fsS "${RELAY_URL}/api/status" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "[3/6] Checking /v1 route..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${RELAY_URL}/v1/models" || true)
echo "GET /v1/models => ${HTTP_CODE}"

echo "[4/6] Checking management health..."
curl -fsS "${API_URL}/healthz" >/dev/null
echo "GET /healthz => 200"

echo "[5/6] Budget node oversell (expect 422)..."
CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  -X PUT "${API_URL}/api/budget/nodes/dept-3" \
  -H "Content-Type: application/json" \
  -H "Cookie: tokenjoy_session_member=m-admin" \
  -d '{"budget":90000,"reservedPool":1500}')
echo "PUT budget node => ${CODE}"

echo "[6/7] Simulated NewAPI settle webhook..."
PAYLOAD='{"id":999888777,"token_id":1,"quota":1000,"model":"gpt-4o","created_at":1717200000}'
WH_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "${API_URL}/api/internal/webhooks/newapi-log" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: ${WEBHOOK_SECRET}" \
  -d "${PAYLOAD}" || true)
echo "POST newapi-log webhook => ${WH_CODE}"
if [[ -x "${ROOT}/apps/newapi/patches/webhook/settle_webhook.sh" ]]; then
  MANAGEMENT_WEBHOOK_URL="${API_URL}/api/internal/webhooks/newapi-log" \
  MANAGEMENT_WEBHOOK_SECRET="${WEBHOOK_SECRET}" \
  "${ROOT}/apps/newapi/patches/webhook/settle_webhook.sh" "${PAYLOAD}" || true
fi

echo "[7/7] 502 spike (manual log review)..."
"${ROOT}/apps/newapi/scripts/spike-502.sh" || true

echo "Gate script finished. Fill docs/tokenjoy-architecture.md section 5.9 after reviewing spike output."
