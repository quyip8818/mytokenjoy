#!/usr/bin/env bash
# Fail if open-budget / occurrence period keys are resolved from wall clock.
# time.Now for lease/retry/ID/timestamps is fine; feeding it to SnapshotKey is not.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

violations=0

fail() {
  echo "lint-clock: $1"
  violations=$((violations + 1))
}

# SnapshotKey must not be fed time.Now.
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  fail "SnapshotKey must not use time.Now: $hit"
done < <(rg -n --glob '*.go' 'SnapshotKey\([^)]*time\.Now' \
  internal/domain \
  internal/pkg/budget \
  internal/store \
  internal/infra/worker \
  seed || true)

# Domain open-budget paths must use Open*/Occurrence* factories, not SnapshotKey.
# seed/apply may call SnapshotKey with Snapshot.SeedAt (already from Clock).
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  fail "use Open*/Occurrence* factories, not SnapshotKey: $hit"
done < <(rg -n --glob '*.go' '\bSnapshotKey\(' \
  internal/domain/budget \
  internal/domain/gateway internal/domain/newapisync \
  internal/domain/usage || true)

# Domain/seed must not call the old exported DepartmentPeriodKey name if reintroduced.
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  fail "use OpenDepartmentPeriod / OccurrenceDepartmentPeriod: $hit"
done < <(rg -n --glob '*.go' '\bDepartmentPeriodKey\(' \
  internal/domain \
  seed || true)

if [[ "$violations" -gt 0 ]]; then
  echo "lint-clock: $violations violation(s)."
  exit 1
fi
echo "lint-clock: ok"
