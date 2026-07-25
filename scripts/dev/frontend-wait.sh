#!/usr/bin/env bash
# Wait for backend healthz then start frontend dev server.
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

wait-on -t 60000 http://127.0.0.1:8010/healthz
pnpm -F @tokenjoy/frontend start
