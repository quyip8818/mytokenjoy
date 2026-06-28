#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
WEBHOOK_URL="${MANAGEMENT_WEBHOOK_URL:-http://127.0.0.1:8080/api/internal/webhooks/newapi-log}"
WEBHOOK_SECRET="${MANAGEMENT_WEBHOOK_SECRET:-tokenjoy-webhook-secret}"

export MANAGEMENT_WEBHOOK_URL="${WEBHOOK_URL}"
export MANAGEMENT_WEBHOOK_SECRET="${WEBHOOK_SECRET}"

PAYLOAD='{"id":900001,"token_id":1,"quota":1000,"model":"gpt-4o","created_at":1717200000}'
"${SCRIPT_DIR}/settle_webhook.sh" "${PAYLOAD}"

if [[ "${1:-}" == "--expect-failure" ]]; then
  exit 0
fi

echo "settle_webhook payload posted"
