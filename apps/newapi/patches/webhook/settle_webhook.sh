#!/usr/bin/env bash
set -euo pipefail

WEBHOOK_URL="${MANAGEMENT_WEBHOOK_URL:-}"
WEBHOOK_SECRET="${MANAGEMENT_WEBHOOK_SECRET:-}"

if [[ -z "${WEBHOOK_URL}" ]]; then
  exit 0
fi

if [[ $# -ge 1 && -n "${1}" ]]; then
  PAYLOAD="${1}"
elif [[ ! -t 0 ]]; then
  PAYLOAD="$(cat)"
else
  exit 0
fi

if [[ -z "${PAYLOAD}" ]]; then
  exit 0
fi

curl -fsS -X POST "${WEBHOOK_URL}" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: ${WEBHOOK_SECRET}" \
  -d "${PAYLOAD}" \
  --max-time 10 \
  >/dev/null
