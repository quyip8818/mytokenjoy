# 项目负责人 + 项目额度审批

> 2026-07-23 · 产品设计 + 实施方案 · ✅ 已实现

---

## 需求概述

1. 项目增加 **OwnerID**（负责人），从项目成员中指定
2. Owner 不可被移除出项目成员，除非先更换 Owner
3. 项目 Owner 可以发起「项目额度追加」审批 → 管理员批 → 从部门可分配余额划拨到项目 budget
4. 项目成员可以发起「项目成员额度」审批 → Owner 批 → 从项目未分配余额中切分 member_budget

---

## 现状

### 数据模型

```sql
-- projects 表（无 owner_id 列）
id, company_id, name, budget, owner_department_id, updated_at

-- project_members 表
company_id, project_id, member_id, member_budget
```

```go
type Project struct {
    ID                uuid.UUID
    Name              string
    Budget            int64               // 项目总额度
    Consumed          int64               // 计算值
    MemberIDs         []uuid.UUID         // 成员列表
    MemberBudgets     map[uuid.UUID]int64 // 成员子额度
    OwnerDepartmentID uuid.UUID           // 所属部门
    // 无 OwnerID
}
```

### 前端结构

- 项目通过 `/budget` 页面的树形面板选中后，右侧展示 `ProjectDetail`
- `ProjectDetail` 组成：`ProjectHeader` + `ProjectSummary` + `ProjectMembersSection` + `ProjectSettingsForm`
- 审批提交入口：`openWithRefresh('approval-submit', { defaultType: '...' })` 打开 workflow panel
- 现有 widget 可复用：`BudgetMemberPicker`（平面列表单选/多选）、`BudgetOrgMemberPicker`（组织树多选）
- Session 提供：`memberId`、`permissions[]` → 可判断当前用户是否是 Owner

---

## 设计决策

- **无需 migration**：项目未上线，直接修改 `schema.sql` DDL，重建数据库即可
- **OwnerID 用单指针 `*uuid.UUID`**：不需要三态（nil/清空/设置），nil=不改，non-nil=设置。如需清空 owner 改用专门的 API 或传零值 UUID
- **`decorateCanResolve` owner 判断**：list 时提取所有 `project_member_budget` 类型的 scopeIDs，批量查一次 projects 表拿 ownerID，内存比对 caller
- **前端 `ProjectView` 和 `Project` 类型都加 `ownerId`**：保持接口一致

---

## 设计方案

### 1. Project 加 OwnerID

```go
type Project struct {
    // ...existing...
    OwnerID *uuid.UUID `json:"ownerId,omitempty"`
}

type UpdateProjectInput struct {
    // ...existing...
    OwnerID *uuid.UUID `json:"ownerId"` // nil=不改, non-nil=设置
}
```

```sql
-- 直接在 schema.sql 中加列（项目未上线，无需 migration）
-- projects 表加 owner_id UUID
owner_id UUID,
FOREIGN KEY (company_id, owner_id) REFERENCES members(company_id, id)
```

**约束**：
- OwnerID 必须在 MemberIDs 中（或为 NULL）
- **Owner 不能从 MemberIDs 中移除**，除非同时更换/清空 OwnerID
- 管理员通过 UpdateProject 设置/更换 Owner

### 2. 新增审批类型（✅ 类型已定义，handler 待实现）

| 类型 | 申请人 | 审批人 | 资金流向 |
|------|--------|--------|---------|
| `project_budget` | 项目 Owner | 管理员 (`budget:approve`) | 部门可分配余额 → project.Budget |
| `project_member_budget` | 项目成员 | 项目 Owner (`caller==ownerID`) | project 未分配余额 → member_budget |

> `ApprovalTypeProjectBudget`、`ApprovalTypeProjectMemberBudget` 常量及 `ProjectBudgetApprovalMeta`、`ProjectMemberBudgetApprovalMeta` struct 已在 scope_id 重构中定义于 `domain/types/approval.go`。

### 3. `project_budget` 审批

Meta struct 已定义（见 `domain/types/approval.go`），scope_id = project_id。

**流程**：
- Validate: applicant == project.OwnerID, amount > 0
- PreApprove: 部门可分配余额 >= amount
- OnApprovedTx: lock → 验证余额 → project.Budget += amount → persist
- Compensate: project.Budget -= amount

### 4. `project_member_budget` 审批

Meta struct 已定义（见 `domain/types/approval.go`），scope_id = project_id。

**流程**：
- Validate: applicant 在 project.MemberIDs 中, amount > 0
- PreApprove: project.Budget - Σ已分配 member_budget >= amount
- OnApprovedTx: lock → 验证余额 → project.MemberBudgets[applicant] += amount → persist
- Compensate: project.MemberBudgets[applicant] -= amount

**审批人**：不走 `budget:approve` 权限，改为校验 caller == project.OwnerID。

---

## 前端交互（复用现有 widget）

### 项目详情页 `ProjectDetail`

**管理员视角**（不变 + 增加 Owner 设置）：
- `ProjectSettingsForm` 增加 Owner 选择 → 复用 `BudgetMemberPicker`（从项目 members 中单选）
- 编辑成员时，如果试图移除 Owner → toast 提示"请先更换负责人"

**Owner 视角**（新增按钮）：
- `ProjectDetail` actions 区域增加「申请追加额度」Button
- 点击 → `openWithRefresh('approval-submit', { defaultType: 'project_budget', projectId, projectName })`
- 模式与 `member-keys-page-shell.tsx` 的「申请额度」按钮完全一致

**成员视角**（项目成员列表中新增按钮）：
- `ProjectMembersSection` 表格增加「额度」列（显示 member_budget）
- 当前用户行增加「申请额度」按钮
- 点击 → `openWithRefresh('approval-submit', { defaultType: 'project_member_budget', projectId, projectName })`

### 审批提交表单 `ApprovalSubmitWorkflow`

Select 增加两个选项（根据 payload 中 projectId 是否存在控制显隐）：

```tsx
// 当 payload.projectId 存在时才显示项目相关选项
{projectId && <>
  <SelectItem value="project_budget">项目额度追加</SelectItem>
  <SelectItem value="project_member_budget">项目成员额度</SelectItem>
</>}
```

metadata 构造增加分支：
```tsx
type === 'project_budget' || type === 'project_member_budget'
  ? { projectId, projectName, amount: displayToQuota(...), reason }
  : // existing key/member_budget logic
```

### 审批列表 `useApprovalPage`

- 管理员（`budget:approve`）看到 `project_budget` 待审批 — 已有逻辑
- **Owner 看到 `project_member_budget` 待审批** — 后端 List 已支持 `scopeIds` 过滤，前端调用 List 时传入 `type=project_member_budget` + `scopeIds=[owned project IDs]`
- 成员看到自己提交的所有审批 — 已有逻辑（`applicantId` filter）

### 审批操作权限

- `project_budget` 的 approve/reject：仍需 `budget:approve` 权限（现有逻辑不变）
- `project_member_budget` 的 approve/reject：后端检查 caller == project.OwnerID（不需要 `budget:approve`）
- `canResolve` 由后端 List 逐条返回（scope_id 重构已实现基础逻辑），本 PR 需要在 `decorateCanResolve` 中为 `project_member_budget` 类型增加 owner 判断

---

## 实施清单

### 前置依赖（✅ 已完成）

scope_id 重构已合入，以下改动已就绪：
- `approval_requests` 表用 `scope_id` 替代 `department_id`/`department_name`
- `ApprovalTypeProjectBudget` / `ApprovalTypeProjectMemberBudget` 常量 + Meta struct 已定义
- 前端 `ApprovalType` union 已扩展、`TYPE_LABELS` 已更新
- 后端 List 返回逐条 `canResolve: boolean`（当前统一走 `budget:approve`，待本 PR 扩展为 owner 判断）
- `scope_id` 对 project 类型存 `project_id`，List 支持 `scopeIds` 过滤参数

### 后端

| # | 文件 | 改动 |
|---|------|------|
| 1 | `store/postgres/schema.sql` | `projects` 表加 `owner_id UUID`（直接改 DDL，不做 migration） |
| 2 | `domain/types/budget.go` | Project 加 `OwnerID *uuid.UUID`；UpdateProjectInput 加 `OwnerID *uuid.UUID` |
| 3 | `store/postgres/budget_repo_projects.go` | SELECT/INSERT/UPDATE 加 owner_id 列 |
| 4 | `domain/budget/projects.go` | CreateProject/UpdateProject 处理 OwnerID |
| 5 | `domain/budget/projects.go` | 验证：OwnerID ∈ MemberIDs；移除成员时禁止移除 Owner |
| 6 | `domain/budget/project_budget_approval.go` | 新文件：ProjectBudgetApprovalHandler（owner→管理员） |
| 7 | `domain/budget/project_member_budget_approval.go` | 新文件：ProjectMemberBudgetApprovalHandler（成员→owner） |
| 8 | `app/compose_domain_wire.go` | 注册两个新 handler |
| 9 | `http/handler/approval/handler.go` | `decorateCanResolve`：`project_member_budget` 类型改为 caller==ownerID（批量查一次涉及的 project scopeIDs） |
| 10 | `http/handler/approval/handler.go` | approve/reject 端点：`project_member_budget` 类型改为校验 caller==ownerID（不要求 budget:approve） |

### 前端

| # | 文件 | 改动 |
|---|------|------|
| 1 | `api/types/budget.ts` | Project/ProjectView 加 `ownerId?: string` |
| 2 | `features/workflow/payloads/keys.ts` | `'approval-submit'` payload 加 `projectId?`, `projectName?` |
| 3 | `features/workflow/workflows/approval-submit.tsx` | Select 加两个选项（projectId 存在时显示）；metadata 构造加分支 |
| 4 | `features/workflow/workflows/approval-review.tsx` | 展示项目名（从 meta.projectName 读取） |
| 5 | `features/budget/components/project-detail.tsx` | 当前用户是 Owner 时显示「申请追加额度」Button |
| 6 | `features/budget/components/project-members-section.tsx` | 表格加「额度」列；当前用户行加「申请额度」Button |
| 7 | `features/budget/components/project-settings-form.tsx` | 加 Owner 选择（复用 `BudgetMemberPicker` 做单选） |

### 测试

| # | 文件 | 覆盖 |
|---|------|------|
| 1 | `tests/domain/budget/project_budget_approval_test.go` | Owner 发起、管理员批、部门余额扣减 |
| 2 | `tests/domain/budget/project_member_budget_approval_test.go` | 成员发起、Owner 批、member_budget 增加 |
| 3 | `tests/domain/budget/service_test.go` | Owner ∈ MemberIDs 约束、禁止移除 Owner |

---

## 审批权限模型总结

```
┌──────────────────────────┬────────────────────┬──────────────────────────┐
│ 审批类型                  │ 谁可以审批          │ 权限/校验逻辑             │
├──────────────────────────┼────────────────────┼──────────────────────────┤
│ key                      │ 管理员             │ budget:approve           │
│ member_budget            │ 管理员             │ budget:approve           │
│ project_budget           │ 管理员             │ budget:approve           │
│ project_member_budget    │ 项目 Owner         │ caller == project.OwnerID│
└──────────────────────────┴────────────────────┴──────────────────────────┘
```

---

## 验证清单

1. 管理员创建项目指定 Owner → Owner 正确展示
2. 管理员编辑成员列表试图移除 Owner → 报错"请先更换负责人"
3. 管理员更换 Owner → 原 Owner 可移除、新 Owner 生效
4. Owner 在项目详情看到「申请追加额度」按钮
5. Owner 发起 project_budget 审批 → 管理员审批列表可见
6. 管理员通过 → project.Budget 增加，部门可分配余额减少
7. 部门余额不足 → PreApprove 拦截
8. 非 Owner 发起 project_budget → Validate 拒绝
9. 成员在项目详情看到「申请额度」按钮
10. 成员发起 project_member_budget → Owner 审批列表可见
11. Owner 通过 → member_budget 增加，项目未分配余额减少
12. 项目未分配余额不足 → PreApprove 拦截
13. 非项目成员发起 → Validate 拒绝
14. Owner 拒绝成员申请 → 记录状态正确
15. 管理员直接改 project budget / member_budget → 正常（不走审批）

---

## 不做

- **不做 Owner 审批自己的 project_budget（自己审自己）** — Owner 申请，管理员批
- **不做多 Owner** — 一个项目一个负责人，够用
- **不做项目 budget 缩减审批** — 管理员直接操作
- **不做 project_member_budget 缩减审批** — Owner 直接调整
