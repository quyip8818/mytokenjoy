# 权限管理

> **Source of Truth**：`packages/contracts/permission/manifest.json`  
> **读者**：后端 / 前端开发

---

## 1. 架构总览

```
请求 → Cookie/Bearer JWT → RequireSession（parse + PDP 求值）→ RequireAnyPermission → Handler
```

- Token **仅含 identity**（`sub`=memberID, `company_id`, `sid`, `exp`），不含 permissions/roles
- 授权在服务端 PDP **每请求求值**
- Capability 以 `packages/contracts/permission/manifest.json` 为唯一真相，代码生成前后端常量
- 命名规范：`{domain}:{admin|manage|read}`，层级继承自动展开
- 中间件 / 前端路由均为 **OR 语义**（持有任一即可访问）

---

## 2. 权限清单

### 企业权限（19）

| Domain | Key | 说明 |
|--------|-----|------|
| org | `org:admin` | 组织架构、数据源、角色管理 |
| org | `org:manage` | 成员管理 |
| org | `org:read` | 组织查看 |
| budget | `budget:admin` | 超限策略、预警规则 |
| budget | `budget:manage` | 预算分配、项目管理 |
| budget | `budget:approve` | 预算审批（独立职责） |
| budget | `budget:read` | 预算查看 |
| model | `model:manage` | 模型 CRUD + 路由配置 |
| model | `model:read` | 模型查看 |
| keys | `keys:admin` | 平台 Key 签发/管理 |
| keys | `keys:manage` | 供应商 Key 管理 |
| keys | `keys:read` | Key 查看 |
| billing | `billing:manage` | 充值操作 |
| billing | `billing:read` | 钱包/账单查看 |
| dashboard | `dashboard:read` | 成本看板 + 用量分析 |
| audit | `audit:read` | 审计日志查看 |
| self | `self:keys` | 我的 Key |
| self | `self:approval` | 我的审批 |
| api | `api:call` | API 调用能力 |

### 平台权限（3）

| Key | 说明 |
|-----|------|
| `platform:admin` | 平台超级管理（仅 DirectPermissions 授予） |
| `platform:manage` | 平台日常管理（企业/模型/定价/汇率 CRUD） |
| `platform:read` | 平台只读 |

---

## 3. 层级继承

```
org:admin      → org:manage → org:read
budget:admin   → budget:manage → budget:read
model:manage   → model:read
keys:admin     → keys:manage → keys:read
billing:manage → billing:read
platform:admin → platform:manage → platform:read
```

不参与继承（独立授予）：`budget:approve`、`dashboard:read`、`audit:read`、`self:*`、`api:call`

后端 `permission.ExpandHierarchy()` 在 `ResolveMemberPermissions()` 末尾做 fixpoint 展开。  
前端 `expandHierarchy()` 在 `usePermissions` 中 defense-in-depth 二次展开。

---

## 4. 预设角色

| 角色 | 权限（展开前） |
|------|---------------|
| 超级管理员 | `*`（所有企业权限） |
| 组织管理员 | org:admin, budget:admin, budget:approve, model:manage, keys:admin, billing:manage, dashboard:read, audit:read, self:keys, self:approval |
| 普通成员 | self:keys, self:approval |
| 只读审计员 | org:read, budget:read, keys:read, model:read, billing:read, dashboard:read, audit:read, self:approval |
| API 调用者 | api:call |
| 平台管理员 | platform:manage, self:keys |
| 平台只读 | platform:read, self:keys |

- `*` 展开为全部 19 个企业权限，不含 `platform:*`
- 预设角色 ID 固定（`00000000-0000-0000-0000-00000000000x`），定义在 `domain/grants/roles.go`

---

## 5. 认证（AuthN）

### JWT Session

| Claim | 说明 |
|-------|------|
| `sub` | `members.id` |
| `company_id` | 租户 ID |
| `user_id` | 关联 users 表 |
| `sid` | 会话 ID |
| `iat`/`exp` | 签发/过期 |

签名 HS256，密钥 `SESSION_SECRET`，TTL `SESSION_TTL_SEC`（默认 900s）。  
Cookie：`tokenjoy_session_member`（HttpOnly, SameSite=Lax, Secure）。  
Bearer：`Authorization: Bearer <jwt>`，同一 parse 逻辑。

**禁止**在 token 内放 permissions/roles。

---

## 6. 授权（AuthZ）

### 组件

| 组件 | 包 | 职责 |
|------|---|------|
| PIP | `store/org` | `GetMemberAuthz(companyID, memberID)` → member + roles |
| PDP | `identity/authz` | `ResolveMemberPermissions` → `[]capability` + `readOnly` |
| PEP-Session | `middleware/session` | Parse JWT → 租户校验 → `GetSessionContext` |
| PEP-Authz | `middleware/authz` | `HasAny(permissions, required...)` — OR 语义 |
| PAP | `domain/org` | 角色 CRUD，变更时 bump `authz_revision` |

### 请求流程

```go
claims := httpx.ParseMemberToken(r, issuer)
sessionCtx := authzSvc.GetSessionContext(ctx, claims.CompanyID, memberID)
// 内部: revision 缓存(5s TTL) → LRU 缓存(companyID+memberID+revision)
// cache miss: store.Org().GetMemberAuthz → ResolveMemberPermissions → ExpandHierarchy

authz.HasAny(sessionCtx.Permissions, required...)  // OR 判断
```

### 权限展开规则

1. 遍历成员所属角色
2. Preset 角色：从 manifest `presetRoles` 查表展开
3. Custom 角色：DB `p-*` ID → manifest `permissionIdMap` → capability 字符串
4. 合并 `member.DirectPermissions`（如 `platform:admin`）
5. `ExpandHierarchy()`：按 manifest `hierarchy` fixpoint 展开
6. `readOnly`：无 manifest `writeCapabilities` 中任一项时为 true

### authz_revision 缓存失效

LRU cache key = `(companyID, memberID, revision)`。  
角色 CRUD / 成员-角色绑定变更 / 成员禁用 → 事务内 bump `companies.authz_revision` → 全租户缓存 miss。

---

## 7. 平台权限隔离（Defense in Depth）

`platform:*` 权限在 local 模式下绝对不可用。三层独立防护：

| Layer | 机制 | 效果 |
|-------|------|------|
| Session | `ScopePermissions()` 在 `!SupportSaas` 时剔除 `platform:*` | 前端拿不到权限 → 菜单不显示 |
| Router | `if cfg.SupportSaas { platformhandler.Mount(...) }` | Local 模式 `/api/platform/*` 返回 404 |
| Middleware | `RequirePlatformAdmin`: `!supportSaas → 403` + `companyID != TokenJoyCompanyID → 403` + `!HasAny("platform:manage") → 403` | 即使路由被误注册仍独立阻断 |

任意单点失效，其余两层仍能独立阻断。

---

## 8. 后端 API 权限映射

### /api/org

| 操作 | 权限 |
|------|------|
| GET 全部读接口 | `org:read` |
| 数据源/字段映射/同步配置写入 | `org:admin` |
| 部门 CRUD | `org:admin` |
| 角色 CRUD + 角色成员管理 | `org:admin` |
| 成员 CRUD + 批量邀请/导入/转移 | `org:manage` |

### /api/budget

| 操作 | 权限 |
|------|------|
| GET 全部读接口 | `budget:read` |
| 部门/成员预算分配 + 项目 CRUD | `budget:manage` |
| 超限策略 + 预警规则 CRUD | `budget:admin` |

### /api/models

| 操作 | 权限 |
|------|------|
| GET 列表/路由 | `model:read` |
| 模型 CRUD + 路由配置 | `model:manage` |

### /api/keys

| 操作 | 权限 |
|------|------|
| GET provider/platform 列表 | `keys:read` |
| Provider Key CRUD | `keys:manage` |
| Platform Key CRUD + simulate-bearer | `keys:admin` |

### /api/billing

| 操作 | 权限 |
|------|------|
| GET 钱包/充值记录 | `billing:read` |
| 充值/确认支付 | `billing:manage` |

### /api/dashboard

| 操作 | 权限 |
|------|------|
| 全部（成本/用量）| `dashboard:read` |

### /api/audit

| 操作 | 权限 |
|------|------|
| GET settings/operations/calls | `audit:read` |
| PUT settings | `org:admin` |

### /api/approvals

| 操作 | 权限 |
|------|------|
| GET 列表/详情 | `self:approval` OR `budget:approve` |
| 提交/取消 | `self:approval` |
| 审批/拒绝/重试 | `budget:approve` OR `self:approval` |

### /api/me

| 操作 | 权限 |
|------|------|
| GET dashboard | `self:keys` |
| profile/password/phone/email/sessions | session only |

### /api/notifications

| 操作 | 权限 |
|------|------|
| 用户自有通知 CRUD | session only |
| admin 日志/统计/测试发送 | `audit:read` |

### /api/platform（SaaS only）

| 操作 | 权限 |
|------|------|
| 全部 | `RequirePlatformAdmin`：SaaS + 超级公司 + `platform:manage` |

---

## 9. 前端权限

### 路由守卫

`routes.ts` 中 `requiredPermissions`（OR 语义）控制页面可见性。

| 路由 | requiredPermissions |
|------|---------------------|
| /dashboard/* | `dashboard:read` |
| /keys/platform | `keys:admin`, `keys:read` |
| /keys/provider | `keys:manage`, `keys:read` |
| /approvals | `budget:approve`, `self:approval` |
| /models/list | `model:manage`, `model:read` |
| /models/routing | `model:manage` |
| /budget | `budget:read` |
| /budget/alerts | `budget:admin` |
| /billing | `billing:read` |
| /org/data-source | `org:admin` |
| /org/structure | `org:manage`, `org:read` |
| /org/roles | `org:admin` |
| /audit/* | `audit:read` |
| /me/* | （无限制） |
| /platform/* | `platform:manage` |

### 组件级 PermissionGate

```tsx
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'

<PermissionGate permission={PERMISSION.BUDGET_MANAGE}>
  <Button>创建项目</Button>
</PermissionGate>

<PermissionGate write permission={PERMISSION.KEYS_ADMIN}>
  <Button>签发 Key</Button>
</PermissionGate>
```

### Session 刷新策略

| 触发 | 行为 |
|------|------|
| 响应头 `X-Authz-Revision` > 当前 revision | `refreshSession()` |
| `window.focus` 且距上次 > 60s | `refreshSession()` |
| `BroadcastChannel('tokenjoy-authz')` 消息 | `refreshSession()` |
| 业务 API 返回 403 | `refreshSession()` 一次 |
| PAP mutation 成功 | `broadcastAuthzChange()` |

---

## 10. 路由注册指南

```go
// 仅 Session
middleware.SessionRoutes(r, p)

// Session + 读 capability
middleware.ReadRoutes(r, p, "budget:read")

// 追加写 capability
write := middleware.ReadRoutes(r, p)
write.With(middleware.RequireAnyPermission("budget:manage")).Put("/departments/{id}", handler)
```

新增端点：查 manifest.json → 用 `ReadRoutes` / `RequireAnyPermission` → 如需新 capability 先加 manifest 再 `pnpm generate:permissions`。

---

## 11. RBAC 数据模型

```
members.roles[]  →  roles.permissions[] (TEXT[])
members.direct_permissions[] (TEXT[])
```

- Preset 角色：名称固定，展开由 manifest 决定
- Custom 角色：DB 存 `p-*` ID 引用，PDP 按 `permissionIdMap` 展开
- 角色变更 bump `companies.authz_revision`

---

## 12. manifest.json 结构

```jsonc
{
  "version": 2,
  "capabilities": [...],           // 企业权限
  "platformCapabilities": [...],   // 平台权限
  "permissionIdMap": {...},        // DB p-* → capability
  "presetRoles": {...},            // 预设角色 → capabilities
  "writeCapabilities": [...],      // 写能力集合（判断 readOnly）
  "hierarchy": {...}               // admin → manage → read 展开规则
}
```

变更流程：编辑 manifest.json → `pnpm generate:permissions` → 后端/前端使用生成的常量。

---

## 13. 源码索引

| 模块 | 路径 |
|------|------|
| 契约 | `packages/contracts/permission/manifest.json` |
| 后端生成常量 | `internal/infra/permission/keys.go` |
| 层级展开 | `internal/infra/permission/manifest.go` → `ExpandHierarchy` |
| PDP 服务 | `internal/identity/authz/service.go` |
| 权限展开 | `internal/identity/authz/resolve.go` |
| Session 中间件 | `internal/http/middleware/session.go` |
| Authz 中间件 | `internal/http/middleware/authz.go` |
| Platform 中间件 | `internal/http/middleware/require_platform.go` |
| 路由注册辅助 | `internal/http/middleware/routes.go` |
| 角色常量 | `internal/domain/grants/roles.go` |
| 前端生成常量 | `lib/permission-keys.ts` |
| 前端权限工具 | `lib/permissions.ts` |
| 前端 Session hook | `features/session/use-permissions.ts` |
| 前端 PermissionGate | `features/session/components/permission-gate.tsx` |
| 前端路由定义 | `config/routes.ts` |

---

## 14. 配置

| 变量 | 必填 | 说明 |
|------|------|------|
| `SESSION_SECRET` | ✅ | JWT 签名密钥 |
| `SESSION_TTL_SEC` | — | Access token TTL，默认 900 |
| `REFRESH_TOKEN_TTL_SEC` | — | Refresh token TTL |
| `AUTHZ_CACHE_SIZE` | — | LRU 大小，默认 4096 |
| `SUPPORT_SAAS` | — | 启用 SaaS 模式（平台面） |
