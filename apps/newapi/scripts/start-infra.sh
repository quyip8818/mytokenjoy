#!/usr/bin/env bash
# Heavy infra bootstrap: build/wait NewAPI. Used by pnpm reset, pnpm infra.
# Daily attach: ensure-infra.sh (--no-build).
set -euo pipefail

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"

COMPOSE=(docker compose -f "${VERIFY_COMPOSE_FILE}")
if [[ -n "${COMPOSE_PROJECT:-}" ]]; then
  COMPOSE=(docker compose -p "${COMPOSE_PROJECT}" -f "${VERIFY_COMPOSE_FILE}")
fi

verify_info "Starting postgres + redis..."
"${COMPOSE[@]}" up postgres redis -d --wait

verify_info "Ensuring logs.newapi schema..."
"${COMPOSE[@]}" exec -T postgres psql -U tokenjoy -d logs -v ON_ERROR_STOP=1 \
  -c "CREATE SCHEMA IF NOT EXISTS newapi;"

NEWAPI_IMAGE="$("${COMPOSE[@]}" config --images | awk '/newapi-apps/ { print; exit }')"
if [[ -z "${NEWAPI_IMAGE}" ]]; then
  verify_fail "could not resolve newapi-apps image from compose config"
fi

if docker image inspect "${NEWAPI_IMAGE}" >/dev/null 2>&1; then
  verify_info "Starting newapi-apps (existing image ${NEWAPI_IMAGE})..."
  "${COMPOSE[@]}" up newapi-apps -d --wait --no-build
else
  verify_info "Building newapi-apps image (first run)..."
  "${COMPOSE[@]}" up newapi-apps -d --wait --build
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
