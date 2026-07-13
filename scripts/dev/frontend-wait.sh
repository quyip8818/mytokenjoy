#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"

mode="${1:-full}"
timeout_ms=30000
if [[ "${mode}" == "full" ]]; then
  timeout_ms=60000
fi

wait-on -t "${timeout_ms}" http://127.0.0.1:8080/healthz
pnpm -F @tokenjoy/frontend start
