#!/usr/bin/env bash
# Start Local (私有化) mode.
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/../lib/common.sh"
export MODE=local
source "${ROOT}/scripts/lib/mode-env.sh"
exec bash "${ROOT}/scripts/dev/_start-mode.sh"
