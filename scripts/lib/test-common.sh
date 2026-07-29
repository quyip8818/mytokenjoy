#!/usr/bin/env bash
# Shared logic for test scripts: mode parsing + test infra startup.
# Source this after common.sh. Sets: modes array, nocache bool.

modes=()
nocache=false
for arg in "$@"; do
  case "${arg}" in
    --saas)    modes+=("saas") ;;
    --local)   modes+=("local") ;;
    --nocache) nocache=true ;;
  esac
done

# 无 flag 时：读 TEST_MODE env，仍无则两个都跑
if [[ ${#modes[@]} -eq 0 ]]; then
  if [[ -n "${TEST_MODE:-}" ]]; then
    modes=("${TEST_MODE}")
  else
    modes=("saas" "local")
  fi
fi

# 测试专用 compose（独立于 dev）
TEST_COMPOSE=(docker compose -p tokenjoy-test -f "${ROOT}/docker-compose.test.yml")
"${TEST_COMPOSE[@]}" up postgres redis -d --wait
