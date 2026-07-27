#!/usr/bin/env bash
# Shared paths for scripts/dev.sh and scripts/dev/*.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export PATH="${ROOT}/node_modules/.bin:${PATH}"
COMPOSE=(docker compose -f "${ROOT}/docker-compose.yml")
NEWAPI_SCRIPTS="${ROOT}/apps/newapi/scripts"
