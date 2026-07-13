#!/usr/bin/env bash
# Heavy infra bootstrap: build/wait NewAPI. Used by pnpm docker:reset, pnpm bootstrap, pnpm infra.
# Daily attach: ensure-infra.sh (--no-build).
set -euo pipefail

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"

COMPOSE=(docker compose -f "${VERIFY_COMPOSE_FILE}")

verify_info "Starting postgres + redis..."
"${COMPOSE[@]}" up postgres redis -d --wait

verify_info "Ensuring logs.newapi schema..."
"${COMPOSE[@]}" exec -T postgres psql -U tokenjoy -d logs -v ON_ERROR_STOP=1 \
  -c "CREATE SCHEMA IF NOT EXISTS newapi;"

NEWAPI_IMAGE="$("${COMPOSE[@]}" config --images | awk '/new-api/ { print; exit }')"
if [[ -z "${NEWAPI_IMAGE}" ]]; then
  verify_fail "could not resolve new-api image from compose config"
fi

if docker image inspect "${NEWAPI_IMAGE}" >/dev/null 2>&1; then
  verify_info "Starting new-api (existing image ${NEWAPI_IMAGE})..."
  "${COMPOSE[@]}" up new-api -d --wait --no-build
else
  verify_info "Building new-api image (first run)..."
  "${COMPOSE[@]}" up new-api -d --wait --build
fi

verify_info "Waiting for NewAPI /api/status..."
for _ in $(seq 1 60); do
  if curl -fsS "${NEWAPI_URL}/api/status" >/dev/null 2>&1; then
    verify_info "NewAPI ready at ${NEWAPI_URL}"
    exit 0
  fi
  sleep 2
done

curl -fsS "${NEWAPI_URL}/api/status" >/dev/null || verify_fail "NewAPI /api/status unreachable"
