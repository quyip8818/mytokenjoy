#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCAFFOLD="$ROOT/scaffold"

usage() {
  echo "usage: make scaffold-domain DOMAIN=<name>" >&2
  echo "  <name> must be lowercase letters only (e.g. notification)" >&2
  exit 1
}

DOMAIN="${1:-}"
if [[ -z "$DOMAIN" ]]; then
  usage
fi

if [[ ! "$DOMAIN" =~ ^[a-z][a-z0-9]*$ ]]; then
  echo "error: DOMAIN must match ^[a-z][a-z0-9]*$ (got: $DOMAIN)" >&2
  exit 1
fi

DOMAIN_TITLE="$(python3 -c "print(''.join(p.capitalize() for p in '$DOMAIN'.split('_')))")"
DOMAIN_IMPORT="domain${DOMAIN}"

render() {
  local src="$1"
  local dest="$2"
  sed \
    -e "s/__DOMAIN__/${DOMAIN}/g" \
    -e "s/__DOMAIN_TITLE__/${DOMAIN_TITLE}/g" \
    -e "s/__DOMAIN_IMPORT__/${DOMAIN_IMPORT}/g" \
    "$src" > "$dest"
}

DOMAIN_DIR="$ROOT/internal/domain/$DOMAIN"
HANDLER_FILE="$ROOT/internal/http/handler/${DOMAIN}.go"
DOMAIN_TEST_DIR="$ROOT/tests/domain/$DOMAIN"
HANDLER_TEST_FILE="$ROOT/tests/handler/${DOMAIN}_test.go"

if [[ -e "$DOMAIN_DIR" || -e "$HANDLER_FILE" ]]; then
  echo "error: domain '$DOMAIN' already exists" >&2
  exit 1
fi

mkdir -p "$DOMAIN_DIR" "$DOMAIN_TEST_DIR"

render "$SCAFFOLD/domain/service.go.tmpl" "$DOMAIN_DIR/service.go"
render "$SCAFFOLD/domain/service_test.go.tmpl" "$DOMAIN_TEST_DIR/service_test.go"
render "$SCAFFOLD/handler/handler.go.tmpl" "$HANDLER_FILE"
render "$SCAFFOLD/handler/handler_test.go.tmpl" "$HANDLER_TEST_FILE"

echo "Created:"
echo "  internal/domain/$DOMAIN/service.go"
echo "  internal/http/handler/${DOMAIN}.go"
echo "  tests/domain/$DOMAIN/service_test.go"
echo "  tests/handler/${DOMAIN}_test.go"
echo ""
echo "Manual registration required:"
echo ""
cat "$SCAFFOLD/snippets/permission_key.go.snippet" | sed -e "s/__DOMAIN__/${DOMAIN}/g" -e "s/__DOMAIN_TITLE__/${DOMAIN_TITLE}/g"
echo ""
cat "$SCAFFOLD/snippets/app_register.go.snippet" | sed -e "s/__DOMAIN__/${DOMAIN}/g" -e "s/__DOMAIN_TITLE__/${DOMAIN_TITLE}/g" -e "s/__DOMAIN_IMPORT__/${DOMAIN_IMPORT}/g"
echo ""
cat "$SCAFFOLD/snippets/router_register.go.snippet" | sed -e "s/__DOMAIN__/${DOMAIN}/g" -e "s/__DOMAIN_TITLE__/${DOMAIN_TITLE}/g" -e "s/__DOMAIN_IMPORT__/${DOMAIN_IMPORT}/g"
echo ""
echo "Also add types to internal/domain/types/ and update docs/Frontend-API契约.md if needed."
