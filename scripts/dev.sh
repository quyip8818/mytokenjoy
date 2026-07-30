#!/usr/bin/env bash
# Dev orchestration dispatcher.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEV="${ROOT}/scripts/dev"
SMS="${ROOT}/scripts/dev-sms.sh"

cmd="${1:-}"
shift || true

case "${cmd}" in
  start)
    target="${1:-}"
    if [[ "${target}" == "sms" ]]; then
      exec bash "${SMS}" start
    elif [[ "${target}" == "all" ]]; then
      exec concurrently -n saas,local -c blue,cyan \
        "bash \"${DEV}/start-saas.sh\"" \
        "bash \"${DEV}/start-local.sh\""
    elif [[ "${target}" == "saas" ]]; then
      exec bash "${DEV}/start-saas.sh"
    elif [[ "${target}" == "local" ]]; then
      exec bash "${DEV}/start-local.sh"
    fi
    # Default: Local mode
    exec bash "${DEV}/start-local.sh"
    ;;
  reset)
    target="${1:-}"
    if [[ "${target}" == "sms" ]]; then
      exec bash "${SMS}" reset
    elif [[ "${target}" == "all" ]]; then
      bash "${DEV}/reset.sh"
      exec bash "${SMS}" reset
    fi
    exec bash "${DEV}/reset.sh" "$@"
    ;;
  infra) exec bash "${DEV}/infra.sh" "$@" ;;
  test) exec bash "${DEV}/test.sh" "$@" ;;
  test:integration) exec bash "${DEV}/test-integration.sh" "$@" ;;
  test:e2e) exec bash "${DEV}/test-e2e.sh" "$@" ;;
  frontend-wait) exec bash "${DEV}/frontend-wait.sh" ;;
  "")
    cat <<EOF >&2
usage: scripts/dev.sh <command> [args...]

commands:
  start              Start Local mode (default)
  start local        Start Local mode (PG:5520 Backend:8011 Frontend:9192)
  start saas         Start SaaS mode (PG:5510 Backend:8010 Frontend:9191)
  start all          Start both SaaS and Local in parallel
  start sms          Start sms backend + frontend
  reset [mode]       Reset apps: pnpm reset [local|saas]
  reset sms          Reset sms databases and seed
  reset all          Reset both apps and sms
  infra [sub]        Manage docker infra
  test [args]        Run tests (--saas/--local to filter, default both)
  test:integration   Run integration tests (--saas/--local to filter)
  test:e2e           Run E2E tests (--saas/--local to filter)
EOF
    exit 1
    ;;
  *)
    echo "unknown command: ${cmd}" >&2
    exit 1
    ;;
esac
