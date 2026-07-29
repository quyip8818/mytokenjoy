#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"
# shellcheck source=../lib/test-common.sh
source "${ROOT}/scripts/lib/test-common.sh"

for mode in "${modes[@]}"; do
  echo ""
  echo "════════════════════════════════════════════"
  echo "  TEST_MODE=${mode} (integration)"
  echo "════════════════════════════════════════════"
  echo ""
  TEST_MODE="${mode}" pnpm -F @tokenjoy/backend test:integration
done
