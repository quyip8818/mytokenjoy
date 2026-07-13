#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

sub="${1:-}"
case "${sub}" in
  ""|all) "${NEWAPI_SCRIPTS}/start-infra.sh" ;;
  postgres) "${COMPOSE[@]}" up postgres -d --wait ;;
  redis) "${COMPOSE[@]}" up redis -d --wait ;;
  attach|newapi) "${COMPOSE[@]}" up ;;
  *)
    echo "usage: pnpm infra [postgres|redis|attach]" >&2
    exit 1
    ;;
esac
