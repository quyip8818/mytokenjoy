# 统一审批系统 — 待办

统一审批引擎已落地（`domain/approval/` Engine + 3 个 Handler + `approval_requests` 表 + 前端 `features/approval/` + workflow 面板）。以下为残留死代码清理和建议优化。

---

## 1. 死代码清理

### 后端

| 位置 | 问题 | 处理 |
|------|------|------|
| `tests/seed/tables_test.go` L31 | `TRUNCATE TABLE ... key_approvals ...` — 旧表已删除，schema.sql 中不存在 | 从 TRUNCATE 列表移除 `key_approvals` |

### 前端

| 位置 | 问题 | 处理 |
|------|------|------|
| `features/keys/components/status-badges.tsx` → `ApprovalStatusBadge` | export 但零消费者 | 删除该函数 |
| `features/budget/hooks/use-budget-selection.ts` → `approvalOpen` / `setApprovalOpen` | useState + 返回值，无组件读取 | 删除状态及返回字段 |
| `tests/features/keys/use-approval-page.test.tsx` | 路径在 `tests/features/keys/`，实际测试 `features/approval` | 移动到 `tests/features/approval/use-approval-page.test.tsx` |

### Workflow 归属

| 位置 | 问题 | 处理 |
|------|------|------|
| `features/workflow/definitions/keys.ts` | `ApprovalSubmitWorkflow` / `ApprovalReviewWorkflow` 注册在 keys 定义文件 | 提取至 `definitions/approval.ts` |
| `features/workflow/definitions/workflow-meta.ts` | `'approval-submit': 'keys'`, `'approval-review': 'keys'` | 改为 `'approval'` 命名空间 |

---

## 2. 建议优化

### 2.1 `useApprovalPendingCountQuery` 归属

当前位于 `features/org/hooks/`，消费 `approvalApi.list({ status: 'pending' })`。

**做法：** 移至 `features/approval/hooks/use-approval-pending-count-query.ts`，从 `features/approval/index.ts` 导出。sidebar 改为 `import { useApprovalPendingCountQuery } from '@/features/approval'`。`features/org/query-keys.ts` 中 `approvalPendingCount` 键迁至 `features/approval/lib/query-keys.ts`。

### 2.2 `pendingCount` 计算冗余

`use-approval-page.ts` 的 `pendingCount = approvals.filter(a => a.status === 'pending').length` 是客户端二次过滤。tab='pending' 时等于 `total`；tab≠'pending' 时无意义。

**做法：** 直接取 `tab === 'pending' ? total : 0`，或复用 `useApprovalPendingCountQuery`。移除 `useMemo` 过滤。

### 2.3 Key/Budget Handler 事务内缺 reservedPool 校验

`KeyApprovalHandler.OnApprovedTx` 和 `BudgetApprovalHandler.OnApprovedTx` 未加事务内校验 + 锁。并发审批两笔同时 PreApprove 通过，第二笔 OnApprovedTx 可能超支。

**做法：** 两者 `OnApprovedTx` 开头加 `tx.Budget().AcquireBudgetLock(ctx)` + reservedPool >= requested 校验（对齐 `MemberBudgetApprovalHandler`）。PreApprove 保留作快速失败。

### 2.4 文档 §8 目录结构更新

规划中列出 `approval-detail.tsx` / `approval-form.tsx`，实际使用 workflow 面板替代。

**做法：** 如保留本文档作参考，更新 §8 目录树以反映 workflow 面板方案（`workflow/workflows/approval-submit.tsx` + `approval-review.tsx`）。若不保留本文档则无需操作。

### 2.5 确认 `approval_repo.go` Get 租户隔离

Engine `Get(ctx, id)` 需确认 repo 实现为 `WHERE id = $1 AND company_id = $2`。如未带 `company_id` 条件，需补上以防跨租户泄漏。
