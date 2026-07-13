#!/usr/bin/env bash
# Heavy init: wipe PG, full infra + token + channel, flush redis, sync demo platform keys.
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

"${COMPOSE[@]}" down -v
"${COMPOSE[@]}" up postgres -d --wait
"${NEWAPI_SCRIPTS}/bootstrap-local-after-reset.sh"
"${COMPOSE[@]}" exec -T redis redis-cli FLUSHALL
pnpm -F @tokenjoy/backend dev-bootstrap
