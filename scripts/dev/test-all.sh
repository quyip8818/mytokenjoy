#!/usr/bin/env bash
# test:all — drop test template DBs to force schema rebuild, then run all tests.
set -euo pipefail

# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"
# shellcheck source=../lib/test-common.sh
source "${ROOT}/scripts/lib/test-common.sh"

echo "Dropping test template DBs to force rebuild..."
export DATABASE_URL="${DATABASE_URL:-postgres://tokenjoy:tokenjoy@127.0.0.1:5530/tokenjoy?sslmode=disable}"
psql "${DATABASE_URL}" -c "DROP DATABASE IF EXISTS template_saas;" 2>/dev/null || true
psql "${DATABASE_URL}" -c "DROP DATABASE IF EXISTS template_local;" 2>/dev/null || true
echo "Done. Templates will be rebuilt on first test run."
echo ""

# 前端
pnpm -F @tokenjoy/frontend test

# 后端按 mode 循环
for mode in "${modes[@]}"; do
  echo ""
  echo "════════════════════════════════════════════"
  echo "  TEST_MODE=${mode} (fresh template)"
  echo "════════════════════════════════════════════"
  echo ""
  TEST_MODE="${mode}" pnpm -F @tokenjoy/backend test
done
