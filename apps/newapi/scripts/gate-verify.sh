#!/usr/bin/env bash
set -euo pipefail

VERIFY_NO_BUILD=false
for arg in "$@"; do
  case "${arg}" in
    --no-build) VERIFY_NO_BUILD=true ;;
    -h|--help)
      echo "Usage: gate-verify.sh [--no-build]"
      echo "Smoke: stack + health + create Platform Key + Gateway + webhook."
      echo "Requires Backend at API_URL with NEW_API_ENABLED, Gateway, and demo seed."
      exit 0
      ;;
  esac
done

# shellcheck source=_verify-lib.sh
source "$(cd "$(dirname "$0")" && pwd)/_verify-lib.sh"
trap verify_cleanup EXIT

echo "== TokenJoy verify:gate =="
echo "Backend: ${API_URL} | NewAPI: ${NEWAPI_URL}"

verify_run_gate_flow

echo ""
echo "GATE VERIFY PASSED"
echo "  Stack up"
echo "  Disposable Platform Key created"
echo "  Gateway reachable"
echo "  Webhook accepted"
