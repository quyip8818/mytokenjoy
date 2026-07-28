#!/usr/bin/env bash
# Shared paths for scripts/dev.sh and scripts/dev/*.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export ROOT
export PATH="${ROOT}/node_modules/.bin:${PATH}"
NEWAPI_SCRIPTS="${ROOT}/apps/newapi/scripts"
export NEWAPI_SCRIPTS

# Default compose file (used by SMS and verify scripts).
# Apps scripts override via mode-env.sh.
COMPOSE=(docker compose -p tokenjoy-saas -f "${ROOT}/docker-compose.yml")
