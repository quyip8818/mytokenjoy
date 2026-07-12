# Budget 初始化与校验（已落地）

> **状态：** 已落地。详细规则见 [Backend-预算.md §3](./Backend-预算.md#3-分配层级)。

## 现状摘要

| 层 | 行为 |
| --- | --- |
| **Backend** | 非负与层级约束在 `domain/budget/tree.go`（`ValidateBudgetNodeUpdate` 等）；Handler 零业务规则 |
| **Frontend** | 总额度未设置时 `BudgetInitPrompt`；项目/成员可用额度扣减子部门、项目、成员合计 |
| **单位** | 存储 point（1 元 = 1000 point） |

## 约束公式

- 部门：`budget ≥ Σ子部门 + 预留池 + 项目合计 + 成员额度总和`
- 成员：`personal_budget ≥ 已分配 Key budget`；部门内成员合计 ≤ capacity
- 项目（预算组）：组 budget ≤ 部门可用（扣子部门与成员）

## 源码索引

- `apps/backend/internal/domain/budget/tree.go`
- `apps/backend/internal/pkg/budget/validate.go`
- `apps/frontend/src/features/budget/components/budget-init-prompt.tsx`
