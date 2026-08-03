# 权限模型 v2

## 概述

权限系统从 v1（26 碎片化 key + 散落 implies 逻辑）精简为 v2（19 企业 + 3 平台），统一使用 `admin` / `manage` / `read` 三层命名，引入声明式 `hierarchy` 自动展开机制。

核心设计：
1. 单点真相：`packages/contracts/permission/manifest.json` → 代码生成前后端常量
2. 层级继承：`admin → manage → read` 声明在 manifest，后端 `ExpandHierarchy()` + 前端 `expandHierarchy()` 双重展开
3. 平台隔离：`ScopePermissions` 确保非平台企业永远看不到 `platform:*` 权限
4. OR 语义：中间件 / 前端路由均为"持有任一即可访问"

## 权限清单

### 企业权限（19）

| Domain | Key | 说明 | Group |
|--------|-----|------|-------|
| org | `org:admin` | 组织架构、数据源、角色管理 | 组织 |
| org | `org:manage` | 成员管理 | 组织 |
| org | `org:read` | 组织查看 | 组织 |
| budget | `budget:admin` | 超限策略、预警规则 | 预算 |
| budget | `budget:manage` | 预算分配、项目管理 | 预算 |
| budget | `budget:approve` | 预算审批（独立职责） | 预算 |
| budget | `budget:read` | 预算查看 | 预算 |
| model | `model:manage` | 模型 CRUD + 路由配置 | 模型 |
| model | `model:read` | 模型查看 | 模型 |
| keys | `keys:admin` | 平台 Key 签发/管理 | 凭证 |
| keys | `keys:manage` | 供应商 Key 管理 | 凭证 |
| keys | `keys:read` | Key 查看 | 凭证 |
| billing | `billing:manage` | 充值操作 | 财务 |
| billing | `billing:read` | 钱包/账单查看 | 财务 |
| dashboard | `dashboard:read` | 成本看板 + 用量分析 | 看板 |
| audit | `audit:read` | 审计日志查看 | 审计 |
| self | `self:keys` | 我的 Key | 成员 |
| self | `self:approval` | 我的审批 | 成员 |
| api | `api:call` | API 调用能力 | API |

### 平台权限（3）

| Key | 说明 |
|-----|------|
| `platform:admin` | 平台超级管理（仅 DirectPermissions 授予） |
| `platform:manage` | 平台日常管理（企业/模型/定价/汇率 CRUD） |
| `platform:read` | 平台只读 |

## 层级继承

```
org:admin      → org:manage → org:read
budget:admin   → budget:manage → budget:read
model:manage   → model:read
keys:admin     → keys:manage → keys:read
billing:manage → billing:read
platform:admin → platform:manage → platform:read
```

不参与继承（独立授予）：`budget:approve`、`dashboard:read`、`audit:read`、`self:*`、`api:call`

展开时机：`ResolveMemberPermissions()` 末尾调用 `permission.ExpandHierarchy()`，fixpoint 迭代直到无新增。

## 预设角色

| 角色 | 权限（展开前） |
|------|---------------|
| 超级管理员 | `*`（所有企业权限） |
| 组织管理员 | org:admin, budget:admin, budget:approve, model:manage, keys:admin, billing:manage, dashboard:read, audit:read, self:keys, self:approval |
| 普通成员 | self:keys, self:approval |
| 只读审计员 | org:read, budget:read, keys:read, model:read, billing:read, dashboard:read, audit:read, self:approval |
| API 调用者 | api:call |
| 平台管理员 | platform:manage, self:keys |
| 平台只读 | platform:read, self:keys |

说明：
- `*` 展开为全部 19 个企业权限，不含 `platform:*`
- 预设角色 ID 固定（`00000000-0000-0000-0000-00000000000x`），全局唯一

## 后端 API 权限映射

### /api/org

| 操作 | 权限 |
|------|------|
| GET 全部读接口 | `org:read` |
| 数据源/字段映射/同步配置 写入 | `org:admin` |
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
| GET dashboard（我的用量）| `self:keys` |
| profile/password/phone/email/sessions | session only（无额外权限） |

### /api/notifications

| 操作 | 权限 |
|------|------|
| 用户自有通知 CRUD | session only |
| admin 日志/统计/测试发送 | `audit:read` |

### /api/platform（SaaS only）

| 操作 | 权限 |
|------|------|
| 全部（企业/模型/定价/汇率 CRUD）| `RequirePlatformAdmin`：SaaS 模式 + 超级公司 + `platform:manage` |

### /api/dev（仅 DEPLOY_ENV=local）

| 操作 | 权限 |
|------|------|
| GET /readiness | 无认证（健康检查性质）|
| GET /platform-keys/{id}/bearer | `keys:admin` |

## 前端路由权限

| 路由 | requiredPermissions（OR） |
|------|--------------------------|
| /dashboard/cost | `dashboard:read` |
| /dashboard/usage | `dashboard:read` |
| /keys/platform | `keys:admin`, `keys:read` |
| /approvals | `budget:approve`, `self:approval` |
| /keys/provider | `keys:manage`, `keys:read` |
| /models/list | `model:manage`, `model:read` |
| /models/routing | `model:manage` |
| /budget | `budget:read` |
| /budget/alerts | `budget:admin` |
| /billing | `billing:read` |
| /org/data-source | `org:admin` |
| /org/structure | `org:manage`, `org:read` |
| /org/roles | `org:admin`, `org:read` |
| /audit/operations | `audit:read` |
| /audit/calls | `audit:read` |
| /me/keys | （无限制）|
| /me/usage | （无限制）|
| /me/settings | （无限制）|
| /platform/models | `platform:manage` |
| /platform/companies | `platform:manage` |
| /platform/currencies | `platform:manage` |

前端同时通过 `PermissionGate` 组件对写操作按钮做细粒度控制（keys、models 已覆盖）。

## 技术实现

### Source of truth

`packages/contracts/permission/manifest.json`（version: 2）

### Code generation

- `packages/contracts/permission/generate-backend.go` → `apps/backend/internal/infra/permission/keys.go`（常量 + CompanyPermissions + AllPermissions + PermissionIDMap）
- `packages/contracts/permission/generate-frontend.ts` → `apps/frontend/src/lib/permission-keys.ts`（PERMISSION 对象 + PermissionKey 类型）

### 后端展开

`permission.ExpandHierarchy(perms []string) []string` — fixpoint 迭代，读取 manifest `hierarchy` 字段，在 `ResolveMemberPermissions` 末尾调用。

### 前端展开

`expandHierarchy(perms)` in `apps/frontend/src/lib/permissions.ts` — 同样的 fixpoint 逻辑，defense-in-depth（后端已展开，前端二次确认）。

### ScopePermissions

`authz.ScopePermissions(perms, companyType, supportSaas)` — 返回 session 前过滤 `platform:*`。仅 SaaS 模式下的 `platform_admin` 类型企业保留平台权限。

### ReadOnly 判定

`IsReadOnlySession(permissions)` — 遍历 `writeCapabilities`，若无任何写权限则标记 `readOnly=true`。前端据此全局 disable 写操作。

### 中间件栈

```
RequireSession → [CompanyResolve → authz.GetSessionContext] → RequireAnyPermission(perms...)
```

平台 API 额外：`RequirePlatformAdmin(tokenJoyCompanyID, supportSaas)`

### 缓存

- authz revision per company（5s TTL）避免每请求查 DB
- session context LRU cache（keyed by companyID+memberID+revision）

## 涉及文件

### 权限定义层
- `packages/contracts/permission/manifest.json` — 唯一 source of truth
- `apps/backend/internal/infra/permission/manifest.json` — go:embed 副本
- `apps/backend/internal/infra/permission/keys.go` — 生成的常量
- `apps/backend/internal/infra/permission/manifest.go` — ExpandHierarchy, PresetRoleCapabilities, WriteCapabilities
- `apps/backend/internal/infra/permission/grants.go` — NormalizeGrantIDs, RoleGrantIDs
- `apps/backend/internal/domain/grants/roles.go` — 预设角色名 + 固定 UUID
- `apps/frontend/src/lib/permission-keys.ts` — 生成的前端常量
- `apps/frontend/src/lib/permissions.ts` — expandHierarchy, hasPermission, isReadOnlySession

### 权限执行层
- `apps/backend/internal/identity/authz/resolve.go` — ResolveMemberPermissions + ExpandHierarchy
- `apps/backend/internal/identity/authz/service.go` — GetSessionContext + ScopePermissions + 缓存
- `apps/backend/internal/http/middleware/authz.go` — RequireAnyPermission
- `apps/backend/internal/http/middleware/require_platform.go` — RequirePlatformAdmin
- `apps/backend/internal/http/middleware/session.go` — RequireSession
- `apps/backend/internal/http/middleware/routes.go` — ReadRoutes / SessionRoutes helpers

### 权限消费层
- `apps/backend/internal/http/handler/{org,budget,models,keys,billing,dashboard,audit,approval,me,notification}/handler.go`
- `apps/frontend/src/config/routes.ts` — 前端路由权限守卫
- `apps/frontend/src/features/session/use-permissions.ts` — usePermissions hook
- `apps/frontend/src/features/session/components/permission-gate.tsx` — PermissionGate 组件
