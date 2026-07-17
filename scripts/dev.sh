#!/usr/bin/env bash
# Dev orchestration dispatcher. See docs/本地开发-启动优化.md
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEV="${ROOT}/scripts/dev"

cmd="${1:-}"
shift || true

case "${cmd}" in
  local) exec bash "${DEV}/start-local.sh" ;;
  saas) exec bash "${DEV}/start-saas.sh" ;;
  lite) exec bash "${DEV}/start-lite.sh" ;;
  reset|docker-reset) exec bash "${DEV}/reset.sh" ;;
  infra) exec bash "${DEV}/infra.sh" "$@" ;;
  test) exec bash "${DEV}/test.sh" "$@" ;;
  frontend-wait) exec bash "${DEV}/frontend-wait.sh" "${1:-full}" ;;
  "")
    echo "usage: scripts/dev.sh <local|saas|lite|reset|infra|test|frontend-wait> [args...]" >&2
    exit 1
    ;;
  *)
    echo "unknown command: ${cmd}" >&2
    exit 1
    ;;
esac
