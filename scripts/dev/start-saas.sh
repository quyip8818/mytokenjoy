#!/usr/bin/env bash
# Start SaaS mode.
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"
export MODE=saas
source "${ROOT}/scripts/lib/mode-env.sh"
exec bash "${ROOT}/scripts/dev/_start-mode.sh"
