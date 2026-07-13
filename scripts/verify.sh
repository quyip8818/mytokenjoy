#!/usr/bin/env bash
# CI and stack verification. Invoked via: pnpm verify [ci|gate|integration|newapi]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NEWAPI_SCRIPTS="${ROOT}/apps/newapi/scripts"

target="${1:-ci}"
shift || true

case "${target}" in
  ci)
    pnpm lint
    pnpm test "$@"
    pnpm build
    pnpm -F @tokenjoy/backend build:check
    ;;
  gate)
    exec "${NEWAPI_SCRIPTS}/gate-verify.sh" "$@"
    ;;
  integration)
    exec "${NEWAPI_SCRIPTS}/integration-verify.sh" "$@"
    ;;
  newapi)
    exec "${NEWAPI_SCRIPTS}/test-management.sh" "$@"
    ;;
  *)
    echo "usage: pnpm verify [ci|gate|integration|newapi]" >&2
    exit 1
    ;;
esac
