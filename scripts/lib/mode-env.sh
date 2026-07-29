#!/usr/bin/env bash
# Mode-specific port/env configuration for local and saas.
# Usage: export MODE=local; source scripts/lib/mode-env.sh
# Exports all env vars needed by backend, frontend, and infra scripts.
# shellcheck shell=bash

: "${MODE:?MODE must be set to 'local' or 'saas'}"
: "${ROOT:?ROOT must be set}"

case "${MODE}" in
  local)
    PG_PORT=5520
    REDIS_PORT=6320
    NEWAPI_PORT=3011
    BACKEND_PORT=8011
    FRONTEND_PORT=9192
    MOCK_PORT=8766
    WEB_PORT=5176
    SUPPORT_SAAS=false
    SESSION_SECRET=tokenjoy-dev-session-secret-local
    COMPOSE_FILE="${ROOT}/docker-compose.local.yml"
    COMPOSE_PROJECT=tokenjoy-local
    ;;
  saas)
    PG_PORT=5510
    REDIS_PORT=6310
    NEWAPI_PORT=3010
    BACKEND_PORT=8010
    FRONTEND_PORT=9191
    MOCK_PORT=8765
    WEB_PORT=5175
    SUPPORT_SAAS=true
    SESSION_SECRET=tokenjoy-dev-session-secret
    COMPOSE_FILE="${ROOT}/docker-compose.yml"
    COMPOSE_PROJECT=tokenjoy-saas
    ;;
  *)
    echo "ERROR: MODE must be 'local' or 'saas', got '${MODE}'" >&2
    exit 1
    ;;
esac

# Infra / subprocess vars
export MODE COMPOSE_FILE COMPOSE_PROJECT
export VERIFY_COMPOSE_FILE="${COMPOSE_FILE}"
export PG_PORT REDIS_PORT NEWAPI_PORT BACKEND_PORT FRONTEND_PORT MOCK_PORT WEB_PORT

# Compose array (used by reset.sh, db-reset.sh helpers)
COMPOSE=(docker compose -p "${COMPOSE_PROJECT}" -f "${COMPOSE_FILE}")

# --- Backend ---
export DATABASE_URL="postgres://tokenjoy:tokenjoy@127.0.0.1:${PG_PORT}/tokenjoy?sslmode=disable"
export LOG_DATABASE_URL="postgres://tokenjoy:tokenjoy@127.0.0.1:${PG_PORT}/logs?sslmode=disable"
export PORT="${BACKEND_PORT}"
export CORS_ORIGINS="http://localhost:${FRONTEND_PORT},http://localhost:${WEB_PORT}"
export DEPLOY_ENV=local
export BOOTSTRAP_MODE=demo
export SUPPORT_SAAS
export CLOCK_ANCHOR=2026-06-19
export SECURE_COOKIE=false
export SIMULATE_DELAY=true
export COMPANY_NAME="Demo Company"
export SESSION_SECRET
export INVITE_SECRET=065d16aaaf7457999e5876d921263150c69f0177b772b701c7450a2510aa1cb0
export DATA_SOURCE_CREDENTIAL_KEY=dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=
export NEW_API_ENABLED=true
export NEW_API_GATEWAY_ENABLED=true
export NEW_API_BASE_URL="http://127.0.0.1:${NEWAPI_PORT}"
export NEW_API_ADMIN_USER_ID=1
export NEW_API_WEBHOOK_SECRET=tokenjoy-webhook-secret
export REDIS_URL="redis://127.0.0.1:${REDIS_PORT}/2"
export PLATFORM_BOOTSTRAP_EMAIL=admin@tokenjoy.me
export PLATFORM_BOOTSTRAP_PASSWORD=admin1234

# SaaS side: accept local instance registrations with this secret.
export LOCAL_REGISTRATION_SECRET="dev-local-registration-secret"

# --- Frontend ---
export VITE_API_PROXY_TARGET="http://localhost:${BACKEND_PORT}"
export VITE_DEV_PORT="${FRONTEND_PORT}"
export VITE_SUPPORT_SAAS="${SUPPORT_SAAS}"

# --- Shared URLs (for verify/newapi scripts) ---
export NEWAPI_URL="http://localhost:${NEWAPI_PORT}"
export API_URL="http://localhost:${BACKEND_PORT}"
