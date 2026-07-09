# Budget Module Full Fix Plan

## Overview

Fix all 20 identified issues (P0-P3) across the budget management module. Work is organized into 5 parallel tracks, each targeting a specific layer.

---

## Track 1: P0 — ResolveApproval 执行预算变更 (#2)

**File:** `apps/backend/internal/domain/budget/service.go` (`ResolveApproval` method)

**Problem:** Approval resolves only updates status in DB, does not execute the actual budget mutation (deduct from reserved pool, add to member personal quota).

**Fix:**
1. After `UpdateBudgetApproval` succeeds with status "approved":
   - Load the approval record to get `applicant_id` and `amount`
   - Load the member's department to find the org node
   - Read the node's `reserved_pool`; if `reserved_pool < amount`, return validation error "预留池余额不足"
   - Deduct `amount` from `reserved_pool` via `UpdateNode`-like logic
   - Add `amount` to member's `personal_quota` via `SetMembers`
   - Wrap the entire operation in `s.store.WithTx()` for atomicity
2. Add `BudgetApproval.ApplicantID` and `DepartmentID` fields to support the lookup (the DB already stores `applicant_id`)

**Validation:** The PRD requires "预留池剩余不足时阻止审批通过"—we enforce this before approving.

---

## Track 2: P0 — Concurrency Safety (#6)

**Files:** `apps/backend/internal/domain/budget/service.go`, `apps/backend/internal/store/postgres/budget_repo.go`

**Problem:** `UpdateNode`, `CreateGroup`, `UpdateGroup`, `DeleteGroup`, `CreateAlert`, `UpdateAlert`, `DeleteAlert` all use read-modify-write without locking.

**Fix:** Use PostgreSQL advisory locks scoped per company to serialize budget write operations:

1. Add a helper `acquireBudgetLock(ctx, store, lockKey)` that runs `SELECT pg_advisory_xact_lock($1)` inside a transaction
2. Wrap each budget-mutating service method in `s.store.WithTx(ctx, func(txStore store.Store) error { ... })`:
   - At the start of the tx, acquire advisory lock (`company_id` as lock key)
   - Re-read current state inside the tx
   - Validate and apply
   - Persist via txStore

This ensures:
- Two concurrent `UpdateNode` calls on the same company serialize
- `SetGroups` runs atomically within the tx (already uses `dbQuerier`)
- No stale-read overwrites

**Store changes:**
- Add `AcquireBudgetLock(ctx context.Context) error` to the `BudgetRepository` interface
- Implement: `SELECT pg_advisory_xact_lock(company_id)` (using company_id from context)

---

## Track 3: P1 — Month Reset Rebalance (#1), Overrun Policy UI (#17), Budget Group Key Management (#3)

### 3a. Monthly Rebalance Trigger (#1)

**File:** `apps/backend/internal/infra/worker/runner.go`

**Problem:** When a new month begins, the budget snapshots naturally reset (new period_key has no consumed), but relay token `RemainQuota` is not rebalanced.

**Fix:** Add a monthly reconciliation to the worker:
1. In `relayLoop`, track `lastRebalanceMonth` (initialized to current month)
2. On each tick, check if the current month differs from `lastRebalanceMonth`
3. If it does, enqueue a company-wide rebalance: `EnqueueRebalanceAxis(ctx, "company", companyID)`
4. This triggers `ProcessAxis("company", ...)` which iterates all active mappings and recomputes their `RemainQuota` based on the new (zero consumed) period

### 3b. Overrun Policy Frontend (#17)

**New file:** `apps/frontend/src/features/budget/components/budget-overrun-policy-dialog.tsx`

Add a dialog accessible from the budget alerts page that allows configuring:
- Thresholds (multi-value input for percentage values like 80, 90)
- Notification channels (email/phone/IM checkboxes)
- Block message (text input for custom 429 message)

Wire to existing `budgetApi.updateOverrunPolicy()`.

### 3c. Budget Group Key Management (#3)

The `platform_keys` table already has `budget_group_id` FK. The relay integration already handles budget group keys. The gap is frontend-only.

**Change:** In `budget-detail-project.tsx` (the project/budget-group detail view), add a "API Keys" section that:
- Lists platform keys belonging to this budget group (filter by `budget_group_id`)
- Shows create key button that pre-selects this budget group

This requires a new API call: `GET /api/keys?budgetGroupId={id}` — check if this already exists in the keys handler. If not, add query parameter filtering.

---

## Track 4: P2 — ID Generation (#7), Budget=0 (#8), Member Consumed (#16), DB Constraints (#11-14)

### 4a. ID Generation (#7)

**Files:** `apps/backend/internal/domain/budget/service.go`

Replace `fmt.Sprintf("bg-%d", time.Now().UnixMilli())` and `fmt.Sprintf("alert-%d", time.Now().UnixMilli())` with the `generateID` pattern already used in `structure/id.go`:
```go
func generateBudgetID(prefix string) string {
    b := make([]byte, 4)
    _, _ = rand.Read(b)
    return fmt.Sprintf("%s-%d-%x", prefix, time.Now().UnixMilli(), b)
}
```

### 4b. UpdateGroup Budget=0 (#8)

**File:** `apps/backend/internal/domain/budget/service.go` (`UpdateGroup`)

Change from `if patch.Budget != 0` to use a pointer or a separate "fields to update" approach. Since the HTTP handler receives the full BudgetGroup JSON, the simplest fix:
- Always apply `patch.Budget` (remove the `!= 0` guard)
- The frontend already sends the full budget value on update

### 4c. Member Consumed in Frontend (#16)

**Files:**
- `apps/frontend/src/api/budget.ts` — add `getGroupMemberConsumed(groupId, period)` API call
- Backend: add endpoint `GET /api/budget/groups/{id}/member-consumed` that returns `map[memberId]consumed` from `budget_snapshots` (axis_kind=member, filtered to group members)
- `apps/frontend/src/features/budget/components/budget-project-members-section.tsx` — fetch and display actual consumed values

### 4d. DB Constraints (#11-14)

**File:** `apps/backend/internal/store/postgres/schema.sql`

Add:
```sql
-- #11: FK from alert_rule_notify_roles to alert_rules
ALTER TABLE alert_rule_notify_roles ADD CONSTRAINT fk_alert_rule_notify_roles_rule
    FOREIGN KEY (company_id, rule_id) REFERENCES alert_rules(company_id, id) ON DELETE CASCADE;

-- #12: (informational only - applicant_id is nullable because deleted members should not cascade-delete approval history. No FK needed - this is by design.)

-- #13: Unique constraint on budget group name
CREATE UNIQUE INDEX IF NOT EXISTS idx_budget_groups_unique_name ON budget_groups(company_id, name);

-- #14: CHECK constraint on org_nodes.period
ALTER TABLE org_nodes ADD CONSTRAINT chk_org_nodes_period CHECK (period IN ('monthly', 'quarterly', 'yearly'));
```

Note: #12 is actually not a bug — keeping applicant_id without FK is intentional (audit trail survives member deletion). We'll skip this one.

---

## Track 5: P3 — N+1 Query (#10), Frontend Tests (#20), Error Handling (#18), Design Doc Update (#15)

### 5a. N+1 Query Fix (#10)

**File:** `apps/backend/internal/store/postgres/budget_repo.go` (`Groups` method)

Replace the loop with a batch query approach:
```sql
SELECT group_id, member_id FROM budget_group_members WHERE company_id = $1 ORDER BY group_id, member_id
SELECT group_id, department_id FROM budget_group_departments WHERE company_id = $1 ORDER BY group_id, department_id
```
Then join in-memory.

Same for `AlertRules`:
```sql
SELECT rule_id, role_id FROM alert_rule_notify_roles WHERE company_id = $1 ORDER BY rule_id, role_id
```

### 5b. Frontend Approval Error Handling (#18)

**File:** `apps/frontend/src/features/budget/components/budget-approval-drawer.tsx`

Add toast notifications on error in the catch blocks. Use the project's existing toast pattern.

### 5c. Frontend Tests (#20)

Add tests for:
- Budget allocation editing (validation that child sum <= parent)
- Approval workflow (approve/reject flows)
- Budget group CRUD

### 5d. Design Doc Update (#15)

Update `docs/Backend-预算.md` to reflect that consumed tracking uses `budget_snapshots` exclusively (no per-table `used`/`consumed` columns).

---

## Execution Order

All 5 tracks are independent and can run in parallel via workflow. Track 2 (concurrency) and Track 1 (approval logic) both modify `service.go` but different methods — they can be composed.

## Files Modified Summary

**Backend:**
- `internal/domain/budget/service.go` — approval logic, concurrency wrapping, ID generation, budget=0 fix
- `internal/store/budget_repo.go` — add `AcquireBudgetLock` interface method
- `internal/store/postgres/budget_repo.go` — implement advisory lock, N+1 fix
- `internal/store/postgres/schema.sql` — add FK, unique index, CHECK constraint
- `internal/infra/worker/runner.go` — monthly rebalance trigger
- `internal/http/handler/budget/handler.go` — new endpoint for group member consumed
- `internal/domain/types/budget.go` — add DepartmentID to BudgetApproval if needed

**Frontend:**
- `src/features/budget/components/budget-overrun-policy-dialog.tsx` — new
- `src/features/budget/components/budget-project-members-section.tsx` — use real consumed
- `src/features/budget/components/budget-approval-drawer.tsx` — error handling
- `src/api/budget.ts` — new API call
- `tests/features/budget/` — new test files

**Docs:**
- `docs/Backend-预算.md` — update consumed tracking description
