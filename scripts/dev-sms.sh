#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib/db-reset.sh
source "${ROOT}/scripts/lib/db-reset.sh"

COMPOSE=(docker compose -f "${ROOT}/docker-compose.yml")

cmd_start() {
  cleanup() { pkill -P $$ 2>/dev/null || true; sleep 0.5; pkill -9 -P $$ 2>/dev/null || true; }
  trap cleanup EXIT INT TERM

  "${COMPOSE[@]}" up postgres redis newapi-sms -d --wait

  concurrently --kill-others-on-fail --kill-signal SIGTERM -n sms-be,sms-fe -c blue,green \
    "pnpm -F @sms/backend start" \
    "pnpm -F @sms/frontend dev"
}

cmd_reset() {
  # Stop newapi-sms first — its active connections would block DROP DATABASE.
  "${COMPOSE[@]}" stop newapi-sms 2>/dev/null || true
  "${COMPOSE[@]}" up postgres -d --wait
  reset_sms_databases
  pnpm -F @sms/backend seed
  echo -e "\nSMS reset complete. Next: pnpm start sms"
}

case "${1:-}" in
  start) cmd_start ;;
  reset) cmd_reset ;;
  *) echo "usage: scripts/dev-sms.sh <start|reset>" >&2; exit 1 ;;
esac
