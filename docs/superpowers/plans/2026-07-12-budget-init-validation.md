# Budget Initialization & Validation Enforcement Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce budget allocation rules: total budget must be set first; department budgets cannot exceed parent; member + project budgets cannot exceed department budget.

**Architecture:** Backend enforces all constraints via validation in existing endpoints. Frontend shows initialization prompt when budget=0, and displays "available" amounts to guide users. Toast shows backend error messages on rejection.

**Tech Stack:** Go 1.24 (chi, pgx), React 19, TypeScript, shadcn/ui, sonner

## Global Constraints

- Storage unit: 1 元 = 1000 points. All budget fields stored in points.
- Budget formula per department: `总额度 >= 子部门合计 + 项目合计 + 成员额度总和`
- Available for projects = `部门额度 - 子部门合计 - 成员额度总和`
- Available for members = `部门额度 - 子部门合计 - 项目合计`
- Department budget=0 means "未设置", block downstream operations.

---

### Task 1: Frontend - Budget Initialization Prompt

**Files:**
- Create: `apps/frontend/src/features/budget/components/budget-init-prompt.tsx`
- Modify: `apps/frontend/src/features/budget/components/budget-detail-team.tsx`

**Interfaces:**
- Consumes: `node.budget`, `onUpdateDepartment`
- Produces: `<BudgetInitPrompt>` component shown when `node.budget === 0`

- [ ] **Step 1: Create BudgetInitPrompt component**

A card with message "请先设置总额度" + button that opens a Dialog with amount input.

```tsx
// budget-init-prompt.tsx
import { useState } from 'react'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { displayToPoints, formatDisplayCurrency } from '@/lib/points'
import { Wallet } from 'lucide-react'

interface BudgetInitPromptProps {
  departmentId: string
  departmentName: string
  onUpdateDepartment: (id: string, data: { budget: number }) => Promise<void>
}

export function BudgetInitPrompt({ departmentId, departmentName, onUpdateDepartment }: BudgetInitPromptProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)

  async function handleSave() {
    const value = parseFloat(draft)
    if (Number.isNaN(value) || value <= 0) {
      toast.error('请输入有效的额度')
      return
    }
    setSaving(true)
    try {
      await onUpdateDepartment(departmentId, { budget: displayToPoints(value) })
      setDialogOpen(false)
      toast.success(`已设置${departmentName}总额度为 ${formatDisplayCurrency(displayToPoints(value))}`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '设置失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <div className="flex flex-col items-center gap-4 rounded-lg border border-dashed border-border p-8 text-center">
        <Wallet className="size-8 text-muted-foreground" />
        <div>
          <p className="text-sm font-medium text-foreground">尚未设置预算额度</p>
          <p className="mt-1 text-xs text-muted-foreground">请先设置总额度，然后再分配部门和成员额度</p>
        </div>
        <Button onClick={() => setDialogOpen(true)}>设置总额度</Button>
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>设置{departmentName}总额度</DialogTitle>
          </DialogHeader>
          <div className="py-2">
            <Input
              type="number"
              min={0}
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') void handleSave() }}
              placeholder="输入总额度（元）"
              className="tabular-nums"
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)} disabled={saving}>取消</Button>
            <Button onClick={handleSave} disabled={saving || !draft.trim()}>
              {saving ? '设置中…' : '确定'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
```

- [ ] **Step 2: Integrate into budget-detail-team.tsx**

When `node.budget === 0`, show `BudgetInitPrompt` instead of the normal detail content.

```tsx
// In BudgetDetailTeam, after computing values:
if (node.budget === 0) {
  return (
    <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-5">
      <h3 className="text-sm font-semibold text-foreground">{node.name}</h3>
      <BudgetInitPrompt
        departmentId={node.id}
        departmentName={node.name}
        onUpdateDepartment={onUpdateDepartment}
      />
    </div>
  )
}
```

- [ ] **Step 3: Build and verify**

```bash
pnpm -F @tokenjoy/frontend exec tsc --noEmit
```

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/src/features/budget/components/budget-init-prompt.tsx \
  apps/frontend/src/features/budget/components/budget-detail-team.tsx
git commit -m "feat(budget): 总额度未设置时显示初始化提示"
```

---

### Task 2: Backend - Fix UpdateNode Validation to Include Projects + Members

**Files:**
- Modify: `apps/backend/internal/pkg/budget/validate.go`
- Modify: `apps/backend/internal/domain/budget/tree.go` (pass groups + members to validate)

**Interfaces:**
- Consumes: `store.Budget().Groups()`, `store.Org().Members()`
- Produces: Updated `ValidateBudgetNodeUpdate` that rejects if `childrenSum + projectSum + memberSum + reserved > newBudget`

- [ ] **Step 1: Update ValidateBudgetNodeUpdate signature**

Add `groups []types.BudgetGroup` and `members []types.Member` parameters.

```go
func ValidateBudgetNodeUpdate(
    tree []types.BudgetNode,
    nodeID string,
    newBudget float64,
    newReservedPool float64,
    groups []types.BudgetGroup,
    members []types.Member,
) *string {
```

- [ ] **Step 2: Add project + member sum to floor constraint**

```go
childrenSum := SumChildrenBudget(*node)
projectSum := GroupsBudgetForDept(groups, nodeID)
memberSum := MemberBudgetSumForDept(members, nodeID)
totalAllocated := childrenSum + newReservedPool + projectSum + memberSum
if newBudget < totalAllocated {
    msg := fmt.Sprintf("部门预算不能低于已分配总额（子部门¥%.0f + 项目¥%.0f + 成员¥%.0f + 预留池¥%.0f = ¥%.0f）",
        childrenSum/1000, projectSum/1000, memberSum/1000, newReservedPool/1000, totalAllocated/1000)
    return &msg
}
```

- [ ] **Step 3: Add helper functions**

```go
func GroupsBudgetForDept(groups []types.BudgetGroup, deptID string) float64 {
    sum := 0.0
    for _, g := range groups {
        for _, d := range g.DepartmentIDs {
            if d == deptID { sum += g.Budget; break }
        }
    }
    return sum
}

func MemberBudgetSumForDept(members []types.Member, deptID string) float64 {
    sum := 0.0
    for _, m := range members {
        if m.DepartmentID == deptID { sum += m.PersonalBudget }
    }
    return sum
}
```

- [ ] **Step 4: Update service call site to pass groups + members**

In `tree.go` UpdateNode, load groups and members inside the transaction and pass to validate.

- [ ] **Step 5: Build**

```bash
cd apps/backend && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add apps/backend/internal/pkg/budget/validate.go apps/backend/internal/domain/budget/tree.go
git commit -m "fix(budget): UpdateNode 校验加入项目和成员额度"
```

---

### Task 3: Frontend - Project Dialog Include Member Sum in Available

**Files:**
- Modify: `apps/frontend/src/features/budget/components/budget-project-dialog.tsx`

**Interfaces:**
- Consumes: `getMemberBudgets(departmentId)` (needs to be added as prop)
- Produces: `available = dept.budget - childrenSum - existingProjects - memberBudgetSum`

- [ ] **Step 1: Add member budget loading to available calculation**

The dialog should show: `可用额度 = 部门额度 - 子部门合计 - 已有项目合计 - 成员额度总和`

Pass `memberBudgetSum` as a prop (computed by parent), or load inside dialog.

- [ ] **Step 2: Update validation error message**

When budget exceeds available, the message should explain what's consuming the space.

- [ ] **Step 3: Build and commit**

```bash
pnpm -F @tokenjoy/frontend exec tsc --noEmit
git commit -m "fix(budget): 创建项目可用额度扣减成员额度总和"
```

---

### Task 4: Backend - UpdateMemberBudget Single-Member Validation

**Files:**
- Modify: `apps/backend/internal/pkg/budget/memberbudgetquota.go` (`ValidateMemberBudgetUpdate`)

**Interfaces:**
- Consumes: existing `ValidateMemberBudgetUpdate` function
- Produces: validation that `memberBudgetSum + projectSum <= dept.budget - childrenSum`

- [ ] **Step 1: Update ValidateMemberBudgetUpdate**

Current validation only checks `personalBudget >= allocated key budget`. Add:
```go
// After existing check, add capacity check:
projectSum := GroupsBudgetForDept(groups, member.DepartmentID)  // need groups param
childrenSum := SumChildrenBudget(*deptNode)
otherMemberSum := ... // sum of all other members' personalBudget in same dept
totalAfter := childrenSum + projectSum + otherMemberSum + personalBudget
if totalAfter > deptNode.Budget {
    msg := "成员额度设置后将超出部门预算上限"
    return &msg
}
```

- [ ] **Step 2: Pass groups to ValidateMemberBudgetUpdate call site**

- [ ] **Step 3: Build and commit**

```bash
cd apps/backend && go build ./...
git commit -m "fix(budget): 单个成员额度更新增加部门预算上限校验"
```

---

## Status: PENDING

Ready for implementation.
