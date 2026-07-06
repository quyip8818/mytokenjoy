#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REF="$(tr -d '[:space:]' < "${ROOT}/patches/new-api/UPSTREAM_REF")"

docker build \
  --build-arg "UPSTREAM_REF=${REF}" \
  -f "${ROOT}/Dockerfile" \
  -t tokenjoy-new-api:local \
  "${ROOT}"
