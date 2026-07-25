# 统一审批系统

## 概览

审批系统基于 **Engine + Handler** 模式实现，Engine 负责状态机编排（create → approve/reject/cancel/retry），各业务 Handler 实现具体的校验与副作用逻辑。

当前注册了 4 个审批类型：

| type | Handler | 所属 domain | 作用 |
|------|---------|------------|------|
| `key` | `KeyApprovalHandler` | `domain/keys` | 申请创建平台 Key + 个人额度 |
| `member_budget` | `MemberBudgetApprovalHandler` | `domain/budget` | 部门预留池拨付至个人额度 |
| `project_budget` | `ProjectBudgetApprovalHandler` | `domain/budget` | 项目 Owner 申请追加项目额度（部门预留池 → 项目 budget） |
| `project_member_budget` | `ProjectMemberBudgetApprovalHandler` | `domain/budget` | 项目成员申请子额度（项目未分配余额 → member_budget） |

---

## 架构

```
┌─────────────────────────────────────────────────────┐
│ HTTP Handler (http/handler/approval/)                        │
│   POST /   GET /   GET /:id   GET /:id/pre-check           │
│   PUT /:id/approve   PUT /:id/reject                        │
│   PUT /:id/cancel    PUT /:id/retry                         │
└───────────────────────────┬─────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────┐
│ Engine (domain/approval/engine.go)                           │
│   状态机：pending → approved / rejected / cancelled / failed │
│   事务编排：PreApprove → Tx(OnApprovedTx) → PostApprove     │
│   补偿链：PostApprove 失败 → Compensate → markFailed        │
└───────────────────────────┬─────────────────────────┘
                            │ dispatch by req.Type
         ┌──────────┬───────┼──────────┬───────────────┐
         ▼          ▼       ▼          ▼               │
   KeyApproval  MemberBudget  ProjectBudget  ProjectMemberBudget
   Handler      Handler       Handler        Handler
```

---

## 审批权限模型

| 审批类型 | 谁可以审批 | 权限/校验逻辑 |
|---------|-----------|--------------|
| `key` | 管理员 | `budget:approve` |
| `member_budget` | 管理员 | `budget:approve` |
| `project_budget` | 管理员 | `budget:approve` |
| `project_member_budget` | 项目 Owner | `caller == project.OwnerID`（不需要 `budget:approve`） |

HTTP 层通过 `authorizeResolve` 方法在 approve/reject 前做细粒度鉴权。`decorateCanResolve` 在 list 返回时逐条标记 `canResolve: boolean`。

---

## 状态流转

```
              ┌──── cancel (仅申请人) ────→ cancelled
              │
pending ──────┼──── approve ──────────────→ approved
              │                               │
              │                    PostApprove 失败
              │                               ▼
              │                            failed ──── retry ──→ approved
              │
              └──── reject ───────────────→ rejected
```

---

## Handler 接口

每个审批类型实现 `approval.Handler` 接口（8 个方法）：

```go
type Handler interface {
    Type() types.ApprovalType
    Validate(ctx, input CreateInput) error
    PreApprove(ctx, req ApprovalRequest) error
    OnApprovedTx(ctx, req ApprovalRequest, tx Store) (ApproveResult, error)
    PostApprove(ctx, req ApprovalRequest, result ApproveResult) error
    Compensate(ctx, req ApprovalRequest, result ApproveResult) error
    OnRejected(ctx, req ApprovalRequest, tx Store) error
    PreCheck(ctx, req ApprovalRequest) (json.RawMessage, error)
}
```

### 方法职责

| 方法 | 调用时机 | 事务 | 职责 |
|------|---------|------|------|
| `Validate` | 创建申请时 | 无 | 校验 metadata 合法性（字段、权限、模型白名单等） |
| `PreApprove` | 审批通过前 | 无 | 快速失败检查（如预留池余额）。无锁，可能有 stale read |
| `OnApprovedTx` | 审批通过 | **事务内** | 核心业务副作用：扣余额、创建 Key。**必须加 `AcquireBudgetLock`** |
| `PostApprove` | 事务提交后 | 无 | 外部 IO（同步 NewAPI Token 等）。失败触发 Compensate |
| `Compensate` | PostApprove 失败 / Retry 前 | 无 | 幂等回滚 OnApprovedTx 的数据。`result=nil` 时从 DB 推断 |
| `OnRejected` | 拒绝时 | 事务内 | 拒绝的副作用（当前所有 Handler 均为 no-op） |
| `PreCheck` | 前端审批前预检 | 无 | 返回 JSON 供前端展示条件（如余额是否充足） |

---

## 各 Handler 业务流程

### `key` — 申请创建 Key

- Validate: metadata 合法（模型在白名单内等）
- OnApprovedTx: 创建 Key + 增加 personal budget
- PostApprove: 同步 NewAPI Token

### `member_budget` — 个人额度追加

- Validate: amount > 0, reason 非空
- PreApprove: 部门预留池余额 ≥ amount
- OnApprovedTx: 部门预留池 -= amount, personal_budget += amount
- PostApprove: 入队 rebalance

### `project_budget` — 项目额度追加

- Validate: applicant == project.OwnerID, amount > 0
- PreApprove: 部门预留池余额 ≥ amount
- OnApprovedTx: 部门预留池 -= amount, project.Budget += amount

### `project_member_budget` — 项目成员子额度

- Validate: applicant ∈ project.MemberIDs, amount > 0
- PreApprove: 项目未分配余额 ≥ amount
- OnApprovedTx: project.MemberBudgets[applicant] += amount

---

## 并发安全

所有涉及余额操作的 `OnApprovedTx` 必须：

1. 在方法开头调用 `tx.Budget().AcquireBudgetLock(ctx)` — 公司级 advisory lock
2. 获取锁后再读取 members/budget 数据 — 确保无 stale read
3. 执行校验（reservedPool 是否足够等）
4. 写入变更

`PreApprove` 作为无锁快速失败，防止明显不满足条件的请求进入事务。

---

## 数据库

```sql
CREATE TABLE approval_requests (
    id              UUID PRIMARY KEY,
    company_id      UUID NOT NULL REFERENCES companies(id),
    type            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    applicant_id    UUID NOT NULL,
    applicant_name  TEXT NOT NULL,
    scope_id        UUID NOT NULL,
    metadata        JSONB NOT NULL DEFAULT '{}',
    approver_id     UUID,
    approver_name   TEXT,
    reject_reason   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ,
    CONSTRAINT valid_approval_status CHECK (status IN ('pending','approved','rejected','cancelled','failed'))
);
```

- `scope_id`：通用作用域标识。对 `key` / `member_budget` 类型存 `department_id`；对 `project_budget` / `project_member_budget` 类型存 `project_id`
- `metadata` 字段存储各类型的业务参数，由各 Handler 自行序列化和解析
- 部门名、项目名等展示字段下沉到 `metadata` 内
- 支持 `scopeIds` 过滤参数（Owner 查看自己项目下的待审批）

租户隔离：所有查询带 `company_id` 条件（通过 `store.CompanyID(ctx)` 从请求上下文获取）。

---

## 前端

### 目录结构

```
features/approval/
├── index.ts                              -- barrel export
├── hooks/
│   ├── use-approval-page.ts             -- 审批列表页 hook
│   └── use-approval-pending-count-query.ts -- sidebar 待审批计数
├── components/
│   └── approval-page-shell.tsx          -- 页面壳
└── lib/
    ├── query-keys.ts                    -- TanStack Query keys
    └── types.ts                         -- ApprovalTab 等类型

features/workflow/
├── definitions/approval.ts              -- 注册 approval-submit + approval-review
└── workflows/
    ├── approval-submit.tsx              -- 发起申请面板
    └── approval-review.tsx              -- 审批处理面板（含 pre-check）
```

### 交互流程

1. 用户点击「发起申请」→ 打开 `approval-submit` workflow 面板 → 填写类型/理由/额度/模型 → POST `/approvals`
2. 审批人在审批列表看到 pending 记录 → 点击打开 `approval-review` workflow 面板
3. 面板调用 `GET /approvals/:id/pre-check` 展示余额是否充足
4. 审批人点击「通过」→ PUT `/approvals/:id/approve`；或「拒绝」→ 弹出 `reject-reason` 面板 → PUT `/approvals/:id/reject`

### 项目审批入口

- **项目 Owner** 在 `ProjectDetail` 页看到「申请追加额度」按钮 → 打开 `approval-submit`（defaultType = `project_budget`）
- **项目成员** 在 `ProjectMembersSection` 表格自己行看到「申请额度」按钮 → 打开 `approval-submit`（defaultType = `project_member_budget`）
- 两种项目类型仅在 payload 含 `projectId` 时才在类型选择器中显示

### 权限

| 操作 | 所需权限 |
|------|---------|
| 提交申请 / 撤回 | `self:approval` |
| 查看列表 | `self:approval` 或 `budget:approve` |
| 通过 / 拒绝 / 重试（key / member_budget / project_budget） | `budget:approve` |
| 通过 / 拒绝（project_member_budget） | 项目 Owner（后端校验 `caller == project.OwnerID`） |

---

## 如何添加新的审批类型

### 后端

1. 在 `domain/types/approval.go` 定义新的 `ApprovalType` 常量和 `Metadata` struct
2. 实现 `approval.Handler` 接口（放在业务所属 domain 包内）
3. 在 `app/compose_domain_wire.go` 的 `wireApprovalEngine` 中注册 Handler
4. 如有余额相关操作，`OnApprovedTx` 中加 `AcquireBudgetLock`
5. `PreCheck` 返回前端需要的预检数据
6. 如需自定义审批权限（非 `budget:approve`），在 `http/handler/approval/handler.go` 的 `authorizeResolve` 和 `decorateCanResolve` 中加分支

### 前端

1. 在 `api/types/approval.ts` 的 `ApprovalType` union 中添加新值
2. 在 `features/workflow/payloads/keys.ts` 扩展 `approval-submit` payload（如需额外参数）
3. 在 `approval-submit.tsx` 的类型选择器中添加选项，条件渲染对应表单字段
4. 在 `approval-review.tsx` 中添加对应 metadata 的展示逻辑
5. 如需独立的 workflow 面板，在 `workflows/` 下新建组件并注册到 `definitions/approval.ts`

---

## 注意事项

- `OnApprovedTx` 内只做纯 DB 操作。外部 IO（HTTP 调用第三方）放 `PostApprove`
- `Compensate` 必须幂等 — 可能被调用多次（PostApprove 失败 + Retry 前）
- `PreApprove` 不要加锁 — 它是事务外快速失败，接受 stale read
- Retry 只允许 `status=failed` 的记录。流程：`Compensate(nil)` → 重走 `PreApprove` → `OnApprovedTx` → `PostApprove`
- `metadata` schema 变更要考虑已有 pending 记录的兼容性（建议 optional 字段 + 默认值）
- 前端通过 `useApprovalPendingCountQuery` 轮询 pending 计数（30s 间隔），展示在 sidebar badge
