#!/usr/bin/env bash
set -euo pipefail

NEWAPI_URL="${NEWAPI_URL:-http://localhost:3000}"
API_KEY="${API_KEY:-}"

if [[ -z "${API_KEY}" ]]; then
  echo "API_KEY is required"
  exit 1
fi

echo "== 502 spike: upstream failure =="
STATUS=$(curl -s -o /tmp/spike-502-body.json -w "%{http_code}" \
  -X POST "${NEWAPI_URL}/v1/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"model":"invalid-upstream-model","messages":[{"role":"user","content":"ping"}]}')
echo "HTTP status: ${STATUS}"
echo "Body:"
cat /tmp/spike-502-body.json
echo
echo "Record whether NewAPI produced a log and whether RemainQuota changed in docs/tokenjoy-architecture.md section 5.9"
