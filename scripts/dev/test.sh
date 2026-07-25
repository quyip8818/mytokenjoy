#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

nocache=false
for arg in "$@"; do
  if [[ "${arg}" == "--nocache" ]]; then
    nocache=true
  fi
done

"${COMPOSE[@]}" up postgres -d --wait
if [[ "${nocache}" == "true" ]]; then
  pnpm -F @tokenjoy/frontend -F @tokenjoy/backend --parallel test:nocache
else
  pnpm -F @tokenjoy/frontend -F @tokenjoy/backend --parallel test
fi
