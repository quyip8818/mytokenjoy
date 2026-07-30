# Platform 权限隔离设计

## 纵深防御架构（Defense in Depth）

`platform:manage` 权限在 local 模式下绝对不可用。三层独立防护确保即使其中一层被误改，系统仍然安全。

### Layer 1: Session — 不返回权限

`identity/authz/service.go` 的 `scopePermissions()` 在 `!SupportSaas` 时从 session response 中
剔除 `platform:manage`。前端永远拿不到该权限 → 菜单不显示。

### Layer 2: Router — 不注册路由

`http/router.go` 中 `if d.Config.SupportSaas { platformhandler.Mount(...) }` 决定 platform API
路由是否挂载。Local 模式下 `/api/platform/*` 返回 404。

### Layer 3: Middleware — 强制检查 SaaS 模式

`RequirePlatformAdmin(tokenJoyCompanyID, supportSaas)` 第一个检查就是 `if !supportSaas → 403`。
即使 Layer 2 的路由注册被误改（比如有人去掉了 if 条件），middleware 仍然独立阻断。

### 安全保证

| 攻击面 | Layer 1 | Layer 2 | Layer 3 |
|--------|---------|---------|---------|
| 前端看到菜单 | ✅ 阻断 | - | - |
| 前端发 API 请求 | - | ✅ 404 | ✅ 403 |
| 绕过 session 直接构造请求 | - | ✅ 404 | ✅ 403 |
| 某层代码被误改 | 其余两层仍有效 | 其余两层仍有效 | 其余两层仍有效 |

任意单点失效，其余两层仍能独立阻断。
