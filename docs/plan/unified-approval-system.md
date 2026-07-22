# 统一审批系统

## 目标

合并 `domain/keys/approval.go` 和 `domain/budget/approvals.go` 为统一审批引擎。单表存储、Handler 接口扩展、两阶段副作用（事务内 DB + 事务外 IO + 补偿）。项目未上线，直接替换。

## 1. 数据库

```sql
CREATE TABLE approval_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID NOT NULL REFERENCES companies(id),
    type            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    applicant_id    UUID NOT NULL,
    applicant_name  TEXT NOT NULL,
    department_id   UUID,
    department_name TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    approver_id     UUID,
    approver_name   TEXT,
    reject_reason   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ,
    CONSTRAINT valid_status CHECK (status IN ('pending','approved','rejected','cancelled','failed'))
);

CREATE INDEX idx_approval_company_status ON approval_requests(company_id, status);
CREATE INDEX idx_approval_company_type   ON approval_requests(company_id, type);
CREATE INDEX idx_approval_applicant      ON approval_requests(company_id, applicant_id);
CREATE INDEX idx_approval_created_at     ON approval_requests(company_id, created_at DESC);
```

状态：`pending` → `approved`/`rejected`/`cancelled`/`failed`。`failed` = 审批通过但副作用失败，可 Retry。

## 2. Go 类型

```go
// domain/types/approval.go
package types

type ApprovalStatus string
const (
    ApprovalPending   ApprovalStatus = "pending"
    ApprovalApproved  ApprovalStatus = "approved"
    ApprovalRejected  ApprovalStatus = "rejected"
    ApprovalCancelled ApprovalStatus = "cancelled"
    ApprovalFailed    ApprovalStatus = "failed"
)

type ApprovalType string
const (
    ApprovalTypeKey          ApprovalType = "key"
    ApprovalTypeBudget       ApprovalType = "budget"
    ApprovalTypeMemberBudget ApprovalType = "member_budget"
)

type ApprovalRequest struct {
    ID             uuid.UUID       `json:"id"`
    Type           ApprovalType    `json:"type"`
    Status         ApprovalStatus  `json:"status"`
    ApplicantID    uuid.UUID       `json:"applicantId"`
    ApplicantName  string          `json:"applicantName"`
    DepartmentID   uuid.UUID       `json:"departmentId,omitempty"`
    DepartmentName string          `json:"departmentName,omitempty"`
    Metadata       json.RawMessage `json:"metadata"`
    ApproverID     *uuid.UUID      `json:"approverId,omitempty"`
    ApproverName   *string         `json:"approverName,omitempty"`
    RejectReason   *string         `json:"rejectReason,omitempty"`
    CreatedAt      time.Time       `json:"createdAt"`
    ResolvedAt     *time.Time      `json:"resolvedAt,omitempty"`
}

// 各类型 Metadata
type KeyApprovalMeta struct {
    Reason          string      `json:"reason"`
    RequestedBudget float64     `json:"requestedBudget"`
    RequestedModels []uuid.UUID `json:"requestedModels"`
}

type BudgetApprovalMeta struct {
    Reason          string      `json:"reason"`
    RequestedBudget float64     `json:"requestedBudget"`
    RequestedModels []uuid.UUID `json:"requestedModels,omitempty"`
}

type MemberBudgetApprovalMeta struct {
    Amount int64  `json:"amount"`
    Reason string `json:"reason"`
}
```

## 3. Handler 接口

```go
// domain/approval/handler.go
package approval

// ApproveResult 是 OnApprovedTx 的产出物，Engine 透传给 PostApprove/Compensate。
// 各 Handler 内部断言为具体 struct。无事务外副作用的 Handler 返回 nil。
type ApproveResult interface{}

type Handler interface {
    Type() types.ApprovalType
    Validate(ctx context.Context, input CreateInput) error
    PreApprove(ctx context.Context, req types.ApprovalRequest) error
    OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (ApproveResult, error)
    PostApprove(ctx context.Context, req types.ApprovalRequest, result ApproveResult) error
    // Compensate 必须幂等——多次调用结果等价于一次。
    // Retry 流程会先调 Compensate 清理残留再重走全链。
    Compensate(ctx context.Context, req types.ApprovalRequest, result ApproveResult) error
    OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error
    PreCheck(ctx context.Context, req types.ApprovalRequest) (json.RawMessage, error)
}
```

约定：
- Engine 管状态机 + 编排，Handler 只管业务数据
- `OnApprovedTx` 内涉及余额扣减时，Handler 自行加行锁（`SELECT ... FOR UPDATE`）
- `Compensate` 必须幂等（删不存在的 key = no-op，回滚已回滚的 budget = no-op）
- `tx store.Store` 完整暴露，Handler 只操作自己 domain 的 repo（靠约定，不加 adapter）
- repo 实现层所有查询强制带 `company_id`（从 context 取）

## 4. Engine

```go
// domain/approval/engine.go
package approval

type Engine struct {
    repo     Repository
    txRunner func(ctx context.Context, fn func(store.Store) error) error
    handlers map[types.ApprovalType]Handler
    logger   *slog.Logger
}

func NewEngine(repo Repository, txRunner TxRunner, logger *slog.Logger, handlers ...Handler) *Engine

func (e *Engine) Create(ctx context.Context, input CreateInput) (types.ApprovalRequest, error)
func (e *Engine) Approve(ctx context.Context, id uuid.UUID, approver ApproverInfo) error
func (e *Engine) Reject(ctx context.Context, id uuid.UUID, approver ApproverInfo, reason string) error
func (e *Engine) Cancel(ctx context.Context, id uuid.UUID, applicantID uuid.UUID) error
func (e *Engine) Retry(ctx context.Context, id uuid.UUID) error
func (e *Engine) List(ctx context.Context, filter ListFilter) ([]types.ApprovalRequest, int, error)
func (e *Engine) Get(ctx context.Context, id uuid.UUID) (types.ApprovalRequest, error)
func (e *Engine) PreCheck(ctx context.Context, id uuid.UUID) (json.RawMessage, error)
```

### Approve 流程

```
PreApprove ──fail──► return error（仍 pending）
    │ ok
    ▼
TX{ Update(approved) + OnApprovedTx } ──fail──► 自动回滚（仍 pending）
    │ commit
    ▼
PostApprove(result) ──ok──► 完成 ✓
    │ fail
    ▼
Compensate(result)（尽力而为，幂等）
    │ 无论成功失败
    ▼
标记 failed → 可 Retry
```

**failed 不回滚到 pending**：审批决策是业务事实，已发生不可抹除。failed 状态支持 `Retry`。

### Retry 流程

对 `failed` 状态的审批单：
1. 先调 `Compensate(req, nil)` 清理残留数据（result 为 nil，Handler 从 DB 状态推断需要清理什么）
2. 再走 `PreApprove → TX{OnApprovedTx} → PostApprove` 全链
3. 成功后状态变为 `approved`
4. 再次失败则保持 `failed`（可再次 Retry）

因为 Compensate 幂等 + 能处理 nil result，无论上次失败在哪一步，Retry 都安全。

**Compensate 的两种调用场景：**
- Approve 流程内 PostApprove 失败：`Compensate(req, result)`，result 有值
- Retry 流程开头清理残留：`Compensate(req, nil)`，Handler 从 DB 现状推断

### 辅助类型

```go
// domain/approval/types.go
type CreateInput struct {
    Type           types.ApprovalType
    ApplicantID    uuid.UUID
    ApplicantName  string
    DepartmentID   uuid.UUID
    DepartmentName string
    Metadata       json.RawMessage
}

type ApproverInfo struct {
    ID   uuid.UUID
    Name string
}

type ListFilter struct {
    Status      *types.ApprovalStatus
    Type        *types.ApprovalType
    ApplicantID *uuid.UUID
    Limit       int
    Offset      int
}
```

### Repository

```go
// domain/approval/repository.go
type Repository interface {
    Create(ctx context.Context, req types.ApprovalRequest) error
    Get(ctx context.Context, id uuid.UUID) (types.ApprovalRequest, error)
    Update(ctx context.Context, req types.ApprovalRequest) error
    List(ctx context.Context, filter ListFilter) ([]types.ApprovalRequest, int, error)
}
```

## 5. Handler 实现

### KeyApprovalHandler（domain/keys/approval_handler.go）

```go
type keyApproveResult struct {
    createdKeyID        uuid.UUID
    personalBudgetAdded int64
    departmentID        uuid.UUID
}

func (h *KeyApprovalHandler) Type() types.ApprovalType { return types.ApprovalTypeKey }

func (h *KeyApprovalHandler) Validate(ctx context.Context, input approval.CreateInput) error {
    var meta types.KeyApprovalMeta
    if err := json.Unmarshal(input.Metadata, &meta); err != nil {
        return domain.Validation("invalid metadata")
    }
    if meta.RequestedBudget <= 0 {
        return domain.Validation("requestedBudget must be positive")
    }
    // 校验 requestedModels 属于申请人部门
    return nil
}

func (h *KeyApprovalHandler) PreApprove(ctx context.Context, req types.ApprovalRequest) error {
    // 检查预留池余额 >= requestedBudget（快速失败，无锁）
    return nil
}

func (h *KeyApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
    var meta types.KeyApprovalMeta
    json.Unmarshal(req.Metadata, &meta)
    // 1. 计算是否需要增加 PersonalBudget
    // 2. 创建 PlatformKey 记录
    // 3. SetMembers（如有 budget 变更）
    return &keyApproveResult{createdKeyID: id, personalBudgetAdded: added, departmentID: deptID}, nil
}

func (h *KeyApprovalHandler) PostApprove(ctx context.Context, req types.ApprovalRequest, raw approval.ApproveResult) error {
    result := raw.(*keyApproveResult)
    if h.newAPISync == nil || !h.newAPISync.Enabled() {
        return nil
    }
    _, err := h.newAPISync.SyncPlatformKeyCreate(ctx, result.createdKeyID, result.departmentID)
    return err
}

func (h *KeyApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, raw approval.ApproveResult) error {
    if raw != nil {
        result := raw.(*keyApproveResult)
        // 精确清理：删除 result.createdKeyID + 回滚 result.personalBudgetAdded
        return nil
    }
    // Retry 场景（raw == nil）：从 DB 推断残留
    // 查找 applicant 名下最近创建的未同步 key，删除 + 回滚 budget
    return nil
}

func (h *KeyApprovalHandler) OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error {
    return nil
}
```

### BudgetApprovalHandler（domain/keys/approval_handler.go，同文件）

```go
func (h *BudgetApprovalHandler) Type() types.ApprovalType { return types.ApprovalTypeBudget }

func (h *BudgetApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
    var meta types.BudgetApprovalMeta
    json.Unmarshal(req.Metadata, &meta)
    // 只增加 PersonalBudget（不创建 key）
    members, _ := tx.Org().Members(ctx)
    members = budget.AddMemberPersonalBudget(members, req.ApplicantID, int64(meta.RequestedBudget))
    return nil, tx.Org().SetMembers(ctx, members)
}

func (h *BudgetApprovalHandler) PostApprove(ctx context.Context, _ types.ApprovalRequest, _ approval.ApproveResult) error {
    return nil // 纯 DB 操作，无外部 IO
}

func (h *BudgetApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
    // 幂等回滚 PersonalBudget
    return nil
}
```

### MemberBudgetApprovalHandler（domain/budget/approval_handler.go）

```go
func (h *MemberBudgetApprovalHandler) Type() types.ApprovalType { return types.ApprovalTypeMemberBudget }

func (h *MemberBudgetApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
    var meta types.MemberBudgetApprovalMeta
    json.Unmarshal(req.Metadata, &meta)
    // SELECT ... FOR UPDATE on org_node_budgets(department_id)
    // 检查 reserved_pool >= amount
    // 扣减 ReservedPool + 增加 PersonalBudget
    return nil, nil
}

func (h *MemberBudgetApprovalHandler) PostApprove(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
    // 入队 Rebalance job（幂等）
    return h.enqueuer.InsertRebalance(ctx, store.CompanyID(ctx), store.RebalanceAxisMember, req.ApplicantID)
}

func (h *MemberBudgetApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
    return nil // Rebalance 是幂等后台任务，失败不需要补偿
}
```

## 6. 目录结构

```
apps/backend/internal/
├── domain/
│   ├── approval/           # 引擎框架（零业务依赖）
│   │   ├── engine.go
│   │   ├── handler.go
│   │   ├── repository.go
│   │   └── types.go
│   ├── keys/
│   │   └── approval_handler.go   # KeyApprovalHandler + BudgetApprovalHandler
│   └── budget/
│       └── approval_handler.go   # MemberBudgetApprovalHandler
├── store/postgres/
│   └── approval_repo.go
└── http/handler/approval/
    └── handler.go
```

依赖方向：`keys/budget → approval（接口）`，`approval → 不依赖业务 domain`。

## 7. API

```
GET    /api/approvals                    # ?status=&type=&applicantId=&limit=&offset=
POST   /api/approvals                    # 创建
GET    /api/approvals/{id}               # 详情
PUT    /api/approvals/{id}/approve       # 通过
PUT    /api/approvals/{id}/reject        # { "reason": "..." }
PUT    /api/approvals/{id}/cancel        # 撤回（仅申请人）
PUT    /api/approvals/{id}/retry         # 重试 failed
GET    /api/approvals/{id}/pre-check     # 前置检查
```

权限：`approval.submit`（提交/查看）、`approval.resolve`（通过/拒绝/retry）。cancel 由 Engine 校验申请人。

## 8. 前端

### 入口与权限

- 统一审批中心页面：`/approvals`
- Budget 页面保留"N 条待审批"入口，点击跳转 `/approvals?type=member_budget`
- 有 `approval.resolve` 权限：看到全部审批单，可操作通过/拒绝/retry
- 只有 `approval.submit` 权限：只看到自己的申请（后端 `applicantId` 过滤），可发起/撤回
- 不做双视角 Tab，一个列表根据权限展示不同内容和操作
- 交互模式保留现有 workflow panel（右侧滑出面板：提交、详情/审批、拒绝原因）

### 目录

```
features/approval/
├── index.ts
├── hooks/use-approval-page.ts
├── components/
│   ├── approval-table.tsx
│   ├── approval-detail.tsx      # 按 type switch 渲染 metadata
│   └── approval-form.tsx        # 按 type switch 渲染字段
└── lib/
    ├── types.ts
    └── query-keys.ts
```

```typescript
// api/approval.ts
export const approvalApi = {
  list: (params?) => request<{ items: ApprovalRequest[]; total: number }>(`/approvals${buildQuery(params)}`),
  get: (id: string) => request<ApprovalRequest>(`/approvals/${id}`),
  create: (data) => request<ApprovalRequest>('/approvals', { method: 'POST', body: JSON.stringify(data) }),
  approve: (id: string) => request<void>(`/approvals/${id}/approve`, { method: 'PUT' }),
  reject: (id: string, reason: string) => request<void>(`/approvals/${id}/reject`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  cancel: (id: string) => request<void>(`/approvals/${id}/cancel`, { method: 'PUT' }),
  retry: (id: string) => request<void>(`/approvals/${id}/retry`, { method: 'PUT' }),
  preCheck: (id: string) => request<Record<string, unknown>>(`/approvals/${id}/pre-check`),
}
```

前端类型放 `packages/contracts/src/approval.ts`（见第 2 节 Go 类型的 TS 对应）。

## 9. 删除清单

### 后端
- `domain/keys/approval.go`（整个文件）
- `domain/budget/approvals.go`（整个文件）
- `store/postgres/budget_repo_approvals.go`
- `budget_approvals` 表（DROP TABLE）
- `store.KeysRepository` 的 `Approvals()`/`SetApprovals()` 方法
- `store.BudgetRepository` 的 `BudgetApprovals()`/`SetBudgetApprovals()`/`UpdateBudgetApproval()` 方法
- `Snapshot` 的 `BudgetApprovals` + `Approvals` 字段
- 旧路由：`/api/keys/approvals/*`、`/api/budget/approvals/*`

### 前端
- `features/keys/hooks/use-approval-page.ts`
- `features/keys/components/approval-table.tsx`
- `features/budget/components/budget-approval-drawer.tsx`
- `api/keys.ts` 中 approvalApi
- `api/budget.ts` 中 getApprovals/resolveApproval
- `api/types/keys.ts` 中 KeyApproval
- `api/types/budget.ts` 中 BudgetApproval

## 10. 实施步骤

1. 建表 + `domain/approval/` 包（handler.go, engine.go, repository.go, types.go）
2. `store/postgres/approval_repo.go` + `store.Store` 加 `Approval()` 方法
3. 三个 Handler 实现 + 启动编排注册
4. `http/handler/approval/handler.go` + 路由注册
5. 删除旧后端代码
6. 前端 `features/approval/` + `api/approval.ts` + 页面
7. 删除旧前端代码
