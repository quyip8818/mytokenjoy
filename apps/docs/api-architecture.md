# API 架构指南

## 总览

后端采用 **Go + chi router**，按领域划分 handler，所有业务路由挂载在 `/api` 前缀下。前端通过 `apps/frontend/src/api/` 下对应的 TypeScript client 消费。

```
请求流:
Browser → /api/* → 全局中间件链 → 域路由组 → Handler → Domain Service → Store
```

## 请求处理管道

### 全局中间件 (router.go)

```
RealIP → RequestID → LoggerContext → AccessLog → Recover → SecurityHeaders → CORS
  → [/v1/* Gateway 代理 (如启用)]
  → [/api/* 子路由]:
      RequestTimeout → CompanyResolve → RateLimitTenant → RateLimitLoginPaths
      → AuthzRevisionHeader → CompanyReadOnly
```

### 路由级中间件 (handler 内部)

| 工具函数 | 作用 | 用法 |
|---------|------|------|
| `SessionRoutes(r, p)` | 要求有效 session (已登录) | 用户个人操作 |
| `ReadRoutes(r, p, perms...)` | session + 至少拥有任一 permission | 管理页面读操作 |
| `.With(RequireAnyPermission(...))` | 追加写权限守卫 | 变更操作 |

## 域路由清单

| 域 | 路由前缀 | Handler 包 | 前端 Client | 权限域 |
|---|---|---|---|---|
| Auth | `/auth/*` | `handler/auth` | `api/auth.ts` | 无 (公开) |
| Register | `/auth/register/*` | `handler/register` | 同 auth.ts | 无 (SaaS only) |
| Session | `/session` | `handler/session` | `api/session.ts` | 无 (session only) |
| Me | `/me/*` | `handler/me` | `api/me.ts` | `self:keys` |
| Org | `/org/*` | `handler/org` | `api/org.ts` | `org:*` |
| Budget | `/budget/*` | `handler/budget` | `api/budget.ts` | `budget:*` |
| Keys | `/keys/*` | `handler/keys` | `api/keys.ts` | `keys:*` |
| Models | `/models/*` | `handler/models` | `api/models.ts` | `model:*` |
| Dashboard | `/dashboard/*` | `handler/dashboard` | `api/dashboard.ts` | `dashboard:*` |
| Audit | `/audit/*` | `handler/audit` | `api/audit.ts` | `audit:read` |
| Approval | `/approvals/*` | `handler/approval` | `api/approval.ts` | `self:approval`, `budget:approve` |
| Notification | `/notifications/*` | `handler/notification` | `api/notification.ts` | session / `audit:read` |
| Billing | `/billing/*` | `handler/billing` | `api/billing.ts` | `billing:*` |
| Platform | `/platform/*` | `handler/platform` | (超管专用) | `platform:manage` |
| Ingest | `/internal/*` | `handler/ingest` | - | Webhook Secret |
| Dev | `/dev/*` | `handler/dev` | `api/dev.ts` | 本地开发 only |

## Handler 代码结构约定

每个 handler 包遵循统一模式:

```go
package budget

type Handler struct {
    shared.ProtectedHandlerBase          // 内嵌 Protected 依赖
    service domainbudget.Service          // 注入领域服务接口
}

func NewHandler(p httpdeps.Protected, service domainbudget.Service) *Handler { ... }

func (h *Handler) RegisterRoutes(r chi.Router) {
    // 1. read 组: 需要 session + 读权限
    read := httpmiddleware.ReadRoutes(r, h.Protected, permission.BudgetRead)
    read.Get("/tree", h.Tree)

    // 2. write 组: session + 具体写权限
    write := httpmiddleware.ReadRoutes(r, h.Protected)
    allocate := write.With(httpmiddleware.RequireAnyPermission(permission.BudgetAllocate))
    allocate.Put("/departments/{departmentId}", h.UpdateNode)
}

// Handler 方法: 解析输入 → 调用 service → 统一响应
func (h *Handler) Tree(w http.ResponseWriter, r *http.Request) {
    tree, err := h.service.GetTree(r.Context())
    httputil.WriteJSON(w, http.StatusOK, tree, err)
}
```

### 响应工具函数 (httputil)

| 函数 | 场景 |
|------|------|
| `WriteJSON(w, status, data, err)` | 有返回值的操作 (err 非 nil 时自动走错误路径) |
| `WriteOK(w, data)` | 确定成功、无 err 分支 |
| `WriteVoid(w, err)` | 无返回值的变更操作 (成功返回 204) |
| `WriteError(w, err)` | 手动写错误 (自动识别 DomainError) |
| `WriteStatus(w, code, msg)` | 固定状态码 + 消息 |
| `DecodeJSON(r, &body)` | 请求体解码 (限 1MB、返回 DomainError) |

### 错误模型

所有业务错误通过 `domain.DomainError` 传递:

```go
domain.BadRequest("message")    // 400
domain.Forbidden("message")     // 403
domain.NotFound("message")      // 404
domain.NewDomainError(429, "rate limited") // 自定义状态码
```

Handler 不直接写 500; 非 DomainError 的 err 统一映射为 500 Internal Server Error。

## 前端 Client 约定

```
apps/frontend/src/api/
├── client.ts          // 统一 fetch 封装 (含 401 自动刷新、事件发射)
├── {domain}.ts        // 域 API 函数 (1:1 对应后端域)
├── types/             // 共享类型定义
├── api-events.ts      // unauthorized/forbidden/authzRevision 事件
├── context.tsx        // React Context 注入
└── use-apis.ts        // useApis() hook (features 层通过此获取 API)
```

消费规则:
- features 层通过 `useApis()` / `useInjectedApis()` 获取 API 对象，**禁止直接 import API 函数**
- 每个 `api/{domain}.ts` export 一个对象 (如 `budgetApi`)，聚合该域所有端点
- 请求函数签名: `(params) => request<ResponseType>(path, options)`

## 权限体系

权限定义在 `packages/contracts/permission/` (单一来源)，通过代码生成同步到:
- 后端: `internal/infra/permission/keys.go` (Go 常量)
- 前端: 对应 TypeScript 常量

权限粒度为 `{domain}:{action}`，如 `budget:allocate`、`org:members`。

---

# 修改 / 添加 API 规则

## 添加端点

1. **确认域归属** — 新端点属于哪个已有域？如果跨域，优先放在"操作发起方"所在域。
2. **后端实现**:
   - 在对应 `handler/{domain}/handler.go` 的 `RegisterRoutes` 中注册路由
   - 权限组划分: 读操作放 `read` 组，写操作加具体 permission 中间件
   - Handler 方法只做: 解析输入 → 调 service → 写响应，不放业务逻辑
3. **前端实现**:
   - 在对应 `api/{domain}.ts` 的导出对象中添加函数
   - 类型定义放 `api/types/`
4. **注册 (仅新域)** — 如果是全新域:
   - 创建 `handler/{domain}/` 包
   - 在 `handler/register.go` 的 Registry struct 添加字段 + NewRegistry 中初始化
   - 在 `RegisterAPIRoutes` 添加 `r.Route("/{domain}", reg.xxx.RegisterRoutes)`
   - 前端创建 `api/{domain}.ts`，在 `api/app-apis.ts` 中注册

## 修改端点

- **不需要向后兼容** (项目未上线)，直接改
- 改路由路径时同步改前端 client 中对应的 path 字符串
- 改请求/响应结构时同步改前端 `api/types/` 中的类型

## 命名约定

| 操作 | HTTP 方法 | 路径模式 | 示例 |
|------|-----------|----------|------|
| 列表 | GET | `/` 或 `/资源复数` | `GET /budget/projects` |
| 详情 | GET | `/{id}` | `GET /approvals/{id}` |
| 创建 | POST | `/` 或 `/资源复数` | `POST /budget/projects` |
| 全量更新 | PUT | `/{id}` | `PUT /models/{id}` |
| 部分更新 | PATCH | `/{id}/{field}` | `PATCH /notifications/{id}/read` |
| 删除 | DELETE | `/{id}` | `DELETE /budget/projects/{id}` |
| 状态切换 | PUT | `/{id}/toggle` | `PUT /keys/provider/{id}/toggle` |
| 动作 | POST | `/{id}/{verb}` 或 `/{verb}` | `POST /approvals/{id}/approve` |

## 权限规则

- 新增权限需先在 `packages/contracts/permission/` 添加定义，再重新生成后端/前端常量
- 读操作最多需要一个 read permission
- 写操作必须有明确的 write permission，不允许"仅 session"就能做变更操作 (个人操作如改密码除外)
- 同一 handler 内可以有多个权限组 (参考 budget handler 的 allocate/policy 分组)

## 禁止事项

- 禁止在 handler 中写业务逻辑 (判断、计算、多步编排) — 放 domain service
- 禁止跨域 handler 互调 — 通过 domain service 层的接口协作
- 禁止返回裸 `error` 给前端 — 必须通过 `domain.DomainError` 明确状态码和消息
- 禁止前端直接 import `api/*.ts` 的函数 — 通过 `useApis()` 注入
- 禁止新建与已有域重叠的 handler — 合并进已有域
- 禁止在 handler 中硬编码权限字符串 — 使用 `permission` 包常量
