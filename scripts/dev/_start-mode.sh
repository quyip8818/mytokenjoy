#!/usr/bin/env bash
# Shared start logic — called by start-local.sh / start-saas.sh after mode-env.sh is sourced.
# Expects: ROOT, MODE, COMPOSE_FILE, BACKEND_PORT, FRONTEND_PORT, MOCK_PORT + all exports from mode-env.sh
set -euo pipefail

: "${ROOT:?}" "${MODE:?}" "${COMPOSE_FILE:?}" "${COMPOSE_PROJECT:?}" "${BACKEND_PORT:?}" "${FRONTEND_PORT:?}" "${MOCK_PORT:?}"

cleanup() {
  pkill -P $$ 2>/dev/null || true
  sleep 0.3
  pkill -9 -P $$ 2>/dev/null || true
  lsof -ti :"${BACKEND_PORT}" | xargs kill -9 2>/dev/null || true
  lsof -ti :"${FRONTEND_PORT}" | xargs kill -9 2>/dev/null || true
  lsof -ti :"${MOCK_PORT}" | xargs kill -9 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# Start infra
docker compose -p "${COMPOSE_PROJECT}" -f "${COMPOSE_FILE}" up -d --wait --no-build || {
  echo "Infra not ready. Run: docker compose -p ${COMPOSE_PROJECT} -f ${COMPOSE_FILE} up -d --build" >&2
  exit 1
}

# Locate air binary
AIR="$(go env GOPATH)/bin/air"
if [[ ! -x "${AIR}" ]]; then
  echo "air not found at ${AIR}; run: pnpm install (in apps/backend)" >&2
  exit 1
fi

concurrently --kill-others-on-fail --kill-signal SIGTERM \
  -n "${MODE}:be,${MODE}:fe,${MODE}:mock" \
  -c blue,green,magenta \
  "cd \"${ROOT}/apps/backend\" && \"${AIR}\"" \
  "bash \"${ROOT}/scripts/dev/frontend-wait.sh\"" \
  "PORT=${MOCK_PORT} node \"${ROOT}/apps/dev-mock-llm/src/server.mjs\""
