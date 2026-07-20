#!/usr/bin/env bash
# Dev orchestration dispatcher.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEV="${ROOT}/scripts/dev"

cmd="${1:-}"
shift || true

case "${cmd}" in
  start) exec bash "${DEV}/start.sh" ;;
  lite) exec bash "${DEV}/start-lite.sh" ;;
  reset) exec bash "${DEV}/reset.sh" "$@" ;;
  infra) exec bash "${DEV}/infra.sh" "$@" ;;
  test) exec bash "${DEV}/test.sh" "$@" ;;
  frontend-wait) exec bash "${DEV}/frontend-wait.sh" "${1:-full}" ;;
  "")
    cat <<EOF >&2
usage: scripts/dev.sh <command> [args...]

commands:
  start              Start backend + frontend + mock (run reset first)
  lite               Lightweight: postgres + backend + frontend only
  reset [mode]       Wipe & seed: pnpm reset [local|saas] [--empty|--minimal|--full]
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
