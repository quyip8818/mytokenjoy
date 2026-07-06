#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MGMT="${ROOT}/patches/management"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "${WORKDIR}"' EXIT

mkdir "${WORKDIR}/management"
cp "${MGMT}/notify.go" "${MGMT}/notify_test.go" "${WORKDIR}/management/"
cd "${WORKDIR}"
go mod init github.com/tokenjoy/newapi-management-test
go test -count=1 ./management/...
