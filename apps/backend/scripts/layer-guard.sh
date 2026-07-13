#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
rg -q 'internal/infra/' internal/domain/ && { echo "domain must not import infra"; exit 1; }
rg -q 'integration/newapi|integration/datasource/feishu' internal/domain/ && { echo "domain must not import integration impl"; exit 1; }
rg -q '\.Store\b' internal/http/handler/ && { echo "handler must not use store.Store"; exit 1; }
rg -q 'fanout' internal/infra/river/periodic/ && { echo "fanout periodic forbidden"; exit 1; }
echo "layer-guard ok"
