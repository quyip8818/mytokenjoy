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
      exec concurrently -n apps,sms -c blue,cyan \
        "bash \"${DEV}/start.sh\"" \
        "bash \"${SMS}\" start"
    fi
    exec bash "${DEV}/start.sh"
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
  frontend-wait) exec bash "${DEV}/frontend-wait.sh" ;;
  "")
    cat <<EOF >&2
usage: scripts/dev.sh <command> [args...]

commands:
  start              Start apps backend + frontend + mock
  start sms          Start sms backend + frontend
  start all          Start both apps and sms in parallel
  reset [mode]       Reset apps: pnpm reset [local|saas] [--empty|--minimal|--full]
  reset sms          Reset sms databases and seed
  reset all          Reset both apps and sms
  infra [sub]        Manage docker infra
  test [args]        Run tests
EOF
    exit 1
    ;;
  *)
    echo "unknown command: ${cmd}" >&2
    exit 1
    ;;
esac
