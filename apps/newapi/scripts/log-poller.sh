#!/usr/bin/env bash
set -euo pipefail

RELAY_URL="${RELAY_URL:-http://localhost:3000}"
MANAGEMENT_URL="${MANAGEMENT_URL:-http://localhost:8080/api/internal/webhooks/newapi-log}"
WEBHOOK_SECRET="${NEW_API_WEBHOOK_SECRET:-tokenjoy-webhook-secret}"
ADMIN_TOKEN="${NEW_API_ADMIN_TOKEN:-}"

if [[ -z "${ADMIN_TOKEN}" ]]; then
  echo "NEW_API_ADMIN_TOKEN is required for log polling"
  exit 1
fi

LAST_ID=0
while true; do
  curl -fsS "${RELAY_URL}/api/log/?p=1&page_size=50" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -o /tmp/tokenjoy-logs.json
  COUNT=$(jq '.data | length' /tmp/tokenjoy-logs.json)
  for ((i=0; i<COUNT; i++)); do
    ROW=$(jq -c ".data[$i]" /tmp/tokenjoy-logs.json)
    ID=$(echo "${ROW}" | jq -r '.id')
    if [[ "${ID}" -le "${LAST_ID}" ]]; then
      continue
    fi
    PAYLOAD=$(echo "${ROW}" | jq -c '{id:.id, token_id:.token_id, quota:.quota, model:(.model_name // .model // ""), created_at:(.created_at // 0)}')
    curl -fsS -X POST "${MANAGEMENT_URL}" \
      -H "Content-Type: application/json" \
      -H "X-Webhook-Secret: ${WEBHOOK_SECRET}" \
      -d "${PAYLOAD}" || true
    LAST_ID="${ID}"
  done
  sleep 5
done
