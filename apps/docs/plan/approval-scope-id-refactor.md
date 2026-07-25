# approval_requests 表重构：scope_id 替代 department_id

> 2026-07-23 · 前置重构（为 project_budget / project_member_budget 审批铺路） · ✅ 已实现
> 项目未上线，破坏性更新，无需 migration，直接改 schema

---

## 动机

`approval_requests` 表当前有 `department_id` + `department_name` 两列，语义绑定"申请人所属部门"。随着新增项目相关审批类型（`project_budget`、`project_member_budget`），需要按 `project_id` 过滤审批列表（Owner 查看自己项目下的待审批）。

问题：
- 在通用审批表上加业务语义列（department_id、project_id...）会随类型增长无限膨胀
- `department_name` 是展示字段，不应在关系表上做列

方案：用一个通用的 `scope_id UUID` 列替代 `department_id`，`department_name` 下沉到 metadata。

---

## 设计

### 表结构变更

直接改 schema（项目未上线，无需 migration）：

```sql
CREATE TABLE approval_requests (
    -- ...existing columns...
    -- 删除: department_id UUID, department_name TEXT
    -- 新增:
    scope_id UUID NOT NULL,
    -- ...
);

-- 索引：覆盖按 type + scope 过滤的高频查询
CREATE INDEX idx_approval_type_scope ON approval_requests(company_id, type, scope_id, status);
-- 保留通用查询索引
CREATE INDEX idx_approval_company_status ON approval_requests(company_id, status);
```

`scope_id` 为 NOT NULL —— 所有审批类型都有明确的 scope 归属。

### scope_id 语义约定

| approval type | scope_id 存什么 | 用途 |
|---------------|----------------|------|
| `key` | applicant 的 department_id | 按部门过滤 key 审批 |
| `member_budget` | applicant 的 department_id | 按部门过滤额度审批 |
| `project_budget` | project_id | 管理员按项目过滤 |
| `project_member_budget` | project_id | Owner 查看自己项目下的待审批 |

规则：`scope_id` 是"这条审批归属的作用域实体 ID"，具体含义由 `type` 决定。

### metadata 变更

各类型的 metadata struct 统一增加 applicant 上下文字段（展示用）：

```go
// 所有 metadata 公共部分（约定，非强制 struct）
// departmentId:   申请人所属部门 ID（快照）
// departmentName: 申请人所属部门名称（快照）
```

具体：

```go
type KeyApprovalMeta struct {
    Reason          string      `json:"reason"`
    RequestedBudget float64     `json:"requestedBudget"`
    RequestedModels []uuid.UUID `json:"requestedModels"`
    DepartmentID    uuid.UUID   `json:"departmentId"`    // 新增
    DepartmentName  string      `json:"departmentName"`  // 新增
}

type MemberBudgetApprovalMeta struct {
    Amount         int64     `json:"amount"`
    Reason         string    `json:"reason"`
    DepartmentID   uuid.UUID `json:"departmentId"`   // 新增
    DepartmentName string    `json:"departmentName"` // 新增
}

// 新类型（后续 PR）
type ProjectBudgetApprovalMeta struct {
    ProjectID   uuid.UUID `json:"projectId"`
    ProjectName string    `json:"projectName"`
    Amount      int64     `json:"amount"`
    Reason      string    `json:"reason"`
}

type ProjectMemberBudgetApprovalMeta struct {
    ProjectID   uuid.UUID `json:"projectId"`
    ProjectName string    `json:"projectName"`
    Amount      int64     `json:"amount"`
    Reason      string    `json:"reason"`
}
```

### Go struct 变更

```go
// domain/types/approval.go
type ApprovalRequest struct {
    ID            uuid.UUID       `json:"id"`
    CompanyID     uuid.UUID       `json:"-"`
    Type          ApprovalType    `json:"type"`
    Status        ApprovalStatus  `json:"status"`
    ApplicantID   uuid.UUID       `json:"applicantId"`
    ApplicantName string          `json:"applicantName"`
    ScopeID       uuid.UUID       `json:"scopeId"`  // NOT NULL，替代 DepartmentID
    Metadata      json.RawMessage `json:"metadata"`
    ApproverID    *uuid.UUID      `json:"approverId,omitempty"`
    ApproverName  *string         `json:"approverName,omitempty"`
    RejectReason  *string         `json:"rejectReason,omitempty"`
    CreatedAt     time.Time       `json:"createdAt"`
    ResolvedAt    *time.Time      `json:"resolvedAt,omitempty"`
}

// domain/approval/types.go
type CreateInput struct {
    Type           types.ApprovalType
    ApplicantID    uuid.UUID
    ApplicantName  string
    DepartmentID   uuid.UUID       // 用于 engine 计算 scopeID + 注入 metadata
    DepartmentName string          // 用于 engine 注入 metadata
    Metadata       json.RawMessage // 前端只传业务字段（reason, amount 等）
}
```

注意：
- `ScopeID` 为 `uuid.UUID`（非指针），对应 DB NOT NULL
- `CreateInput` 保留 DepartmentID/DepartmentName——handler 透传 session 信息，由 engine 层决定如何使用（计算 scope_id + 构造完整 metadata）

### 前端 TypeScript 类型变更

```ts
// api/types/approval.ts
export type ApprovalType = 'key' | 'member_budget' | 'project_budget' | 'project_member_budget'

export interface ApprovalRequest {
  id: string
  type: ApprovalType
  status: ApprovalStatus
  applicantId: string
  applicantName: string
  scopeId: string               // 替代 departmentId，NOT NULL
  metadata: Record<string, unknown>
  approverId?: string
  approverName?: string
  rejectReason?: string
  canResolve: boolean           // 后端计算，逐条返回
  createdAt: string
  resolvedAt?: string
}
```

`TYPE_LABELS` 同步扩展：
```ts
const TYPE_LABELS: Record<string, string> = {
  key: 'Key 申请',
  member_budget: '额度追加',
  project_budget: '项目预算',
  project_member_budget: '项目成员额度',
}
```

---

## 现有 Use Case 兼容分析

### Use Case 1：MemberBudgetApprovalHandler.resolveDeptID

**现状**：
```go
func (h *MemberBudgetApprovalHandler) resolveDeptID(ctx context.Context, req types.ApprovalRequest) uuid.UUID {
    if req.DepartmentID != uuid.Nil {
        return req.DepartmentID
    }
    // fallback: 查 member 当前部门
    member, err := h.svc.store.Org().MemberByID(ctx, req.ApplicantID)
    ...
}
```

**迁移后**：
- `scope_id` 存的就是 department_id（对 `member_budget` 类型），直接用 `req.ScopeID`
- 或者从 metadata 中取 `departmentId`（更语义化）
- fallback 逻辑不变

**推荐**：改为从 metadata 取 departmentId，因为 scope_id 是"过滤用途"，业务逻辑应该读 metadata 中的快照。

```go
func (h *MemberBudgetApprovalHandler) resolveDeptID(ctx context.Context, req types.ApprovalRequest) uuid.UUID {
    var meta types.MemberBudgetApprovalMeta
    json.Unmarshal(req.Metadata, &meta)
    if meta.DepartmentID != uuid.Nil {
        return meta.DepartmentID
    }
    member, err := h.svc.store.Org().MemberByID(ctx, req.ApplicantID)
    ...
}
```

**兼容**：✅ 无风险，逻辑等价。

---

### Use Case 2：KeyApprovalHandler 中使用 req.DepartmentID

**现状**（`domain/keys/approval_handler.go`）：
```go
deptID := req.DepartmentID
if deptID == uuid.Nil {
    if applicant, ok := org.FindMemberByID(members, req.ApplicantID); ok {
        deptID = applicant.DepartmentID
    }
}
// 用 deptID 扣减 ReservedPool
```

**迁移后**：改为从 metadata 取。

```go
var meta types.KeyApprovalMeta
json.Unmarshal(req.Metadata, &meta)
deptID := meta.DepartmentID
if deptID == uuid.Nil {
    if applicant, ok := org.FindMemberByID(members, req.ApplicantID); ok {
        deptID = applicant.DepartmentID
    }
}
```

**兼容**：✅ 等价，有 fallback 兜底。

---

### Use Case 3：HTTP Create handler 构造 CreateInput

**现状**：
```go
input := domainapproval.CreateInput{
    Type:           body.Type,
    ApplicantID:    sessionCtx.Member.ID,
    ApplicantName:  sessionCtx.Member.Alias,
    DepartmentID:   sessionCtx.Member.DepartmentID,
    DepartmentName: sessionCtx.Member.DepartmentName,
    Metadata:       body.Metadata,
}
```

**迁移后**：
- handler 只传原始 session 数据，engine 在 Create 内部完成：
  1. 根据 type + input 计算 scope_id
  2. 构造完整 metadata（注入 departmentId/departmentName 快照）

```go
// handler 保持简单——只传 session 信息
input := domainapproval.CreateInput{
    Type:           body.Type,
    ApplicantID:    sessionCtx.Member.ID,
    ApplicantName:  sessionCtx.Member.Alias,
    DepartmentID:   sessionCtx.Member.DepartmentID,
    DepartmentName: sessionCtx.Member.DepartmentName,
    Metadata:       body.Metadata,
}
```

```go
// domain/approval/engine.go Create 内部
func (e *Engine) Create(ctx context.Context, input CreateInput) (types.ApprovalRequest, error) {
    // 1. 计算 scope_id
    scopeID := resolveScopeID(input)

    // 2. 构造完整 metadata（后端控制 schema，前端只传业务字段）
    enrichedMeta := buildMetadata(input)

    req := types.ApprovalRequest{
        // ...
        ScopeID:  scopeID,
        Metadata: enrichedMeta,
    }
    // ...
}

func resolveScopeID(input CreateInput) uuid.UUID {
    switch input.Type {
    case types.ApprovalTypeKey, types.ApprovalTypeMemberBudget:
        return input.DepartmentID
    case types.ApprovalTypeProjectBudget, types.ApprovalTypeProjectMemberBudget:
        var m struct{ ProjectID uuid.UUID `json:"projectId"` }
        json.Unmarshal(input.Metadata, &m)
        return m.ProjectID
    default:
        return input.DepartmentID
    }
}
```

**设计要点**：
- `resolveScopeID` 和 `buildMetadata` 放 engine 包级别函数（非方法），handler 不承担业务逻辑
- metadata 由后端完整构造，前端不需要传 departmentId/departmentName
- 前端只传：`{ reason, amount }` 或 `{ reason, requestedBudget, requestedModels }` 或 `{ projectId, amount, reason }`

**兼容**：✅ 对现有类型（key、member_budget），scope_id = 原来的 department_id，行为不变。

---

### Use Case 4：前端审批列表展示「部门」列

**现状**：`approval-table.tsx` 读 `approval.departmentName`

**迁移后**：改为从 metadata 读取：

```tsx
function getDepartmentName(approval: ApprovalRequest): string {
  const meta = approval.metadata
  return typeof meta.departmentName === 'string' ? meta.departmentName : ''
}
```

**兼容**：✅ 展示效果相同。

---

### Use Case 5：审批列表过滤

**现状**：List 按 `status`、`type`、`applicantId` 过滤，无按 department 过滤。

**迁移后**：List 新增 `scopeId` / `scopeIds` 过滤参数：
```go
if filter.ScopeIDs != nil && len(filter.ScopeIDs) > 0 {
    where += fmt.Sprintf(" AND scope_id = ANY($%d)", argIdx)
    args = append(args, filter.ScopeIDs)
    argIdx++
}
```

**用途**：Owner 查看自己项目下的 `project_member_budget` 待审批时，传入 `type=project_member_budget` + `scopeIds=[owned project IDs]` + `status=pending`。

**兼容**：✅ 新增功能，不影响现有查询。

---

### Use Case 6：canResolve 判定

**现状**：前端全局 `canResolve = has(BUDGET_APPROVE)`，传给 ApprovalTable 作为整体开关。

**迁移后**：
- 后端 List 返回中为每条审批增加 `canResolve: boolean` 字段
- **计算层**：handler List 做装饰（engine List 只返回数据，handler 根据 session 权限计算）
- 后端计算逻辑（当前简化版）：
  - 所有类型统一：caller 有 `budget:approve` → canResolve = true
  - `project_member_budget` 的 owner 判断留了扩展点，等 project-level RBAC 实现后再加
- 前端 ApprovalTable 的 `canResolve` prop 改为逐条从后端返回字段读取

```go
// handler List 内部装饰 canResolve
func decorateCanResolve(items []types.ApprovalRequest, session types.SessionContext) []approvalResponse {
    hasBudgetApprove := slices.Contains(session.Permissions, permission.BudgetApprove)

    result := make([]approvalResponse, len(items))
    for i, item := range items {
        var canResolve bool
        switch item.Type {
        case types.ApprovalTypeKey, types.ApprovalTypeMemberBudget, types.ApprovalTypeProjectBudget:
            canResolve = hasBudgetApprove
        case types.ApprovalTypeProjectMemberBudget:
            // ponytail: 简化——暂时也用 budget:approve，等 project-level RBAC 实现后再改为 owner 判断
            canResolve = hasBudgetApprove
        default:
            canResolve = hasBudgetApprove
        }
        result[i] = approvalResponse{ApprovalRequest: item, CanResolve: canResolve}
    }
    return result
}
```

```ts
// ApprovalRequest 新增
export interface ApprovalRequest {
  // ...existing...
  canResolve: boolean  // 后端计算，非 optional
}
```

**兼容**：✅ 对现有类型，canResolve 计算结果与现在的全局判断一致（有权限的人看所有都是 true）。

---

## 实施清单

| # | 范围 | 文件 | 改动 |
|---|------|------|------|
| 1 | schema | `store/postgres/schema.sql` | 删 department_id/department_name 列，加 scope_id UUID NOT NULL + 双索引 |
| 2 | types | `domain/types/approval.go` | ApprovalRequest 去掉 DepartmentID/DepartmentName，加 ScopeID uuid.UUID |
| 3 | types | `domain/types/approval.go` | KeyApprovalMeta、MemberBudgetApprovalMeta 加 DepartmentID/DepartmentName；新增 ProjectBudgetApprovalMeta、ProjectMemberBudgetApprovalMeta |
| 4 | types | `domain/types/approval.go` | ApprovalType 新增 `project_budget`、`project_member_budget` 常量 |
| 5 | engine | `domain/approval/types.go` | CreateInput 保留 DepartmentID/DepartmentName（session 透传），去掉注释中的"替代"说法 |
| 6 | engine | `domain/approval/engine.go` | Create 内部新增 resolveScopeID + buildMetadata（metadata 由后端完整构造） |
| 7 | store | `store/postgres/approval_repo.go` | Create/Get/List/scan 改用 scope_id（NOT NULL，非指针） |
| 8 | store | `store/store.go` (ApprovalListFilter) | 加 `ScopeIDs []uuid.UUID` 过滤 |
| 9 | handler | `http/handler/approval/handler.go` | Create：handler 只传 session 信息，不做 enrichMetadata；List：加 scopeIds query param + decorateCanResolve |
| 10 | handler | `http/handler/approval/handler.go` | approve/reject：project_member_budget 暂时统一走 budget:approve（后续加 project owner 校验） |
| 11 | budget | `domain/budget/approval_handler.go` | resolveDeptID 改为从 metadata 取 DepartmentID |
| 12 | keys | `domain/keys/approval_handler.go` | PreApprove + **OnApprovedTx** 两处 deptID 均改为从 metadata 取 |
| 13 | frontend | `api/types/approval.ts` | 去掉 departmentId/departmentName 顶层字段，加 scopeId（非 optional）、canResolve；ApprovalType 扩展新类型 |
| 14 | frontend | `features/approval/components/approval-table.tsx` | departmentName 改为从 metadata 取；TYPE_LABELS 扩展；canResolve 改为逐条读取 |
| 15 | frontend | `features/approval/hooks/use-approval-page.ts` | 去掉全局 canResolve 计算（由后端逐条返回） |

---

## 风险

| 风险 | 缓解 |
|------|------|
| metadata 结构变更后旧数据不兼容 | 项目未上线，直接重建表 |
| scope_id 语义随 type 变化，查询时需配合 type 条件 | 文档约定 + List API 始终带 type filter；索引 (company_id, type, scope_id, status) 覆盖 |
| canResolve 当前简化为统一 budget:approve，后续需扩展 | 留了 switch case 扩展点 + 注释说明升级路径 |

---

## 不做

- 不做 migration —— 项目未上线，直接改 schema 重建
- 不加 GIN 索引 —— 审批量级小，B-tree 足够
- 不做 scope_type 列 —— type 字段已经隐含了 scope_id 的语义，不需要冗余
- 不做 metadata JSON schema 校验 —— 由 Engine.Validate（通过 handler 具体类型）保证
- 不做向后兼容 —— 无线上数据
