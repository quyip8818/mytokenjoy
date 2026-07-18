# 认证与授权系统

本文档描述系统当前的认证架构，供 AI 和开发者参考。

---

## 1. 三层身份

| 身份 | 认证方式 | 载体 | 用途 |
|------|----------|------|------|
| **Member Session** | 邮箱+密码 → JWT access token + refresh token | Cookie | Web 管理面板所有操作 |
| **Platform Admin** | 同上（super company 下的 member） | 同上 | SaaS 平台管理（公司创建/充值等） |
| **Platform Key** | API 密钥 `sk-xxx` | `Authorization: Bearer sk-xxx` | AI 网关 API 调用 |

---

## 2. Member Session（Access + Refresh Token）

### 2.1 Token 对

| Token | 载体 | 有效期 | 存储 |
|-------|------|--------|------|
| Access Token | Cookie `tokenjoy_session_member`，Path `/`，SameSite=Lax | 15 min (`SESSION_TTL_SEC`) | 无状态 JWT |
| Refresh Token | Cookie `tokenjoy_refresh`，Path `/api/auth/refresh`，SameSite=Strict | 7 天 (`REFRESH_TOKEN_TTL_SEC`) | DB `sessions` 表（存 SHA-256 hash） |

### 2.2 JWT Claims

```json
{ "sub": "<memberID>", "company_id": "<companyID>", "user_id": "<userID>", "sid": "<sessionID>", "exp": ..., "iat": ... }
```

### 2.3 Refresh Token 格式

```
<sessionID>.<randomHex(32)>
```

不是 JWT。服务端通过 `sessions.token_hash = SHA256(raw)` 验证。生命周期内不变不 rotate。

### 2.4 数据库

```sql
CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,         -- sessionID（= JWT 中的 sid）
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    member_id   UUID NOT NULL,
    company_id  UUID NOT NULL,
    token_hash  TEXT NOT NULL,            -- SHA-256(refresh_token)
    user_agent  TEXT NOT NULL DEFAULT '',
    ip          TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ              -- 非空 = 已吊销
);
```

### 2.5 关键流程

**登录**（Login / AcceptInvite / Register）：
1. 验证凭证 → 拿到 member
2. `sessiontoken.NewSessionID()` 生成 sid
3. `sessiontoken.IssueAccessToken(secret, ttl, companyID, memberID, userID, sid)` 签发 JWT
4. 生成 refresh token `sid + "." + RandomHex(32)`
5. `sessions` 表写入 `(sid, SHA256(refreshRaw), userID, memberID, companyID, expiresAt)`
6. 设置两个 cookie

**Refresh**（`POST /auth/refresh`）：
1. 从 cookie 取 refresh token
2. 解析出 sid
3. `sessions` 表查 `WHERE id = sid AND revoked_at IS NULL AND expires_at > NOW()`
4. 比对 `SHA256(raw) == token_hash`
5. 签发新 access token（refresh token 不变）
6. **无 DB 写操作**，天然幂等

**登出**（`POST /auth/logout`）：
1. 从 JWT 取 sid → `UPDATE sessions SET revoked_at = NOW() WHERE id = sid`
2. 清除两个 cookie

**前端 401 处理**：
1. 任意请求返回 401 → 调 `POST /api/auth/refresh`
2. refresh 成功 → 重试原请求
3. refresh 失败 → emit `'unauthorized'` → 跳转 login 页
4. 并发多个 401 共享同一个 refresh Promise（singleton pattern）

### 2.6 RequireSession Middleware

`RequireSession` 是 JWT 验签 + 权限加载的两阶段 middleware：

1. **JWT 验签**（纯无状态）：验证签名 + 过期时间 → 提取 `companyID`、`memberID`
2. **AuthzSvc 权限加载**：调用 `AuthzSvc.GetSessionContext(companyID, memberID)` 获取 member 状态、角色权限、billing 信息

**不查 `sessions` 表**——revoke session 后 access token 在 TTL（15min）内仍有效。但 member 被 deactivate 后立即拒绝（通过 AuthzSvc 的 member status 检查）。

#### AuthzSvc 缓存架构（性能关键路径）

```
请求 → JWT 验签(0 IO) → GetSessionContext
                            ├─ revisionCache.get(companyID)       [5s TTL 内存缓存]
                            │    └─ miss → SELECT authz_revision FROM companies (1 DB)
                            ├─ LRUCache.Get(companyID, memberID, revision)  [进程内 LRU, 容量 4096]
                            │    └─ miss → GetMemberAuthz (1 DB, JOIN roles)
                            └─ ResolveCompanyChargeRate            [⚠️ 当前未缓存]
                                 ├─ Company.GetByID (1 DB)
                                 └─ Billing.GetCurrency (1 DB)
```

| 场景 | DB 查询数 | 说明 |
|------|----------|------|
| 完全命中 | 2 | billing 无缓存（Company + Currency） |
| revision 命中 + authz miss | 3 | 新 member 或 revision 变更 |
| 全 miss | 4 | 冷启动、新进程首次请求 |

### 2.7 Authz Revision 机制

权限变更（角色调整、成员增删）会 bump `companies.authz_revision`。

- **后端**：response header `X-Authz-Revision` 携带当前 revision
- **前端**：每次 API 响应检查 `X-Authz-Revision`，若 revision > 本地缓存 → 触发 session refetch
- **跨 tab**：通过 `BroadcastChannel('tokenjoy-authz')` 同步

效果：权限变更后前端 **秒级感知**（下一个 API 响应即触发刷新），无需等 access token 过期。

---

## 3. Platform Admin

Platform admin **不是独立身份系统**，而是 `TokenJoyCompanyID`（super company `00000000-0000-7000-8000-000000000001`）下的 member，拥有 `platform:manage` 权限。

- 登录：`POST /api/platform/auth/login` → `AuthenticateMember(TokenJoyCompanyID, email, password)` → `issueTokenPair`
- 路由保护：`RequireSession` + `RequirePlatformAdmin`（检查 `companyID == TokenJoyCompanyID` + `platform:manage`）
- 使用与租户 member 完全相同的 cookie、refresh 机制

Bootstrap 配置：
- `PLATFORM_BOOTSTRAP_EMAIL` — 首个 platform admin 邮箱
- `PLATFORM_BOOTSTRAP_PASSWORD` — 首个 platform admin 密码

---

## 4. Platform Key（API 网关认证）

Platform Key 是分配给租户 member/project 的 API 密钥，用于 AI 网关调用，与 Web session 完全无关。

### 4.1 认证流程

```
客户端 → Authorization: Bearer sk-xxx → 网关
```

1. `parseBearerSecret(header)` 提取 `sk-xxx`
2. `store.HashPlatformKey(secret)` → SHA-256 hash
3. Per-key 限流检查（Redis token bucket，fail-open）
4. `GatewayPrecheck`：用 key_hash 查 `platform_keys` 表，验证 key 状态、过期时间、预算、模型白名单
5. 通过 → 反向代理转发到 NewAPI

### 4.2 数据模型

```sql
platform_keys (
    id, company_id, name, key_prefix, key_hash,
    scope ('member'|'project'|'project_member'),
    member_id, project_id, status, budget, expires_at, ...
)
```

- `key_hash` 是密钥的不可逆 hash，用于查找
- `key_prefix`（如 `sk-abc1...`）供 UI 展示
- 完整密钥仅在创建时返回一次，服务端不存储明文

### 4.3 与 Session 的关系

无关。Platform Key 不经过 `RequireSession` middleware，不涉及 JWT、cookie、refresh token。网关是独立的 HTTP handler，直接处理 Bearer token。

---

## 5. 配置项

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SESSION_SECRET` | —（必填） | JWT 签名密钥（HS256） |
| `SESSION_TTL_SEC` | `900` | Access token 有效期（秒） |
| `REFRESH_TOKEN_TTL_SEC` | `604800` | Refresh token 有效期 + cookie MaxAge（秒） |
| `AUTHZ_CACHE_SIZE` | `4096` | AuthzSvc LRU 缓存条目数（进程内） |
| `TOKENJOY_COMPANY_ID` | `00000000-0000-7000-8000-000000000001` | Super company（platform admin 归属） |
| `PLATFORM_BOOTSTRAP_EMAIL` | — | 首个 platform admin 邮箱 |
| `PLATFORM_BOOTSTRAP_PASSWORD` | — | 首个 platform admin 密码 |
| `SECURE_COOKIE` | `false` | 生产环境设为 `true`（HTTPS only） |

---

## 6. 代码位置

| 功能 | 路径 |
|------|------|
| JWT 签发/解析 | `internal/identity/sessiontoken/issuer.go` |
| Cookie 操作 | `internal/identity/httpx/token.go` |
| Token Pair 签发 | `internal/identity/httpx/issue.go` |
| Session 存储 | `internal/store/session_repo.go` + `postgres/session_repo.go` |
| Auth handler（Login/Logout/Refresh） | `internal/http/handler/auth/` |
| RequireSession middleware | `internal/http/middleware/session.go` |
| RequirePlatformAdmin middleware | `internal/http/middleware/require_platform.go` |
| AuthzSvc（权限缓存） | `internal/identity/authz/service.go` |
| AuthzSvc LRU 缓存 | `internal/identity/authz/cache.go` |
| Authz Revision header | `internal/http/middleware/authz_revision.go` |
| Gateway（Platform Key 认证） | `internal/domain/gateway/gateway_service.go` |
| Platform Key hash | `internal/store/platform_key_mapping_repo.go` |

---

## 7. 安全要点

- 两个 cookie 都是 HttpOnly（XSS 无法读取）
- Refresh cookie SameSite=Strict + Path 限定（CSRF 防护）
- Access token 短命（15min），泄露窗口小
- 服务端可吊销 session（revoke → 最多 15min 后 access 过期）
- Member deactivate 即时生效（AuthzSvc 每次请求验证 member status）
- Platform Key 明文仅创建时返回，DB 存 hash
- Super company 不可被删除/suspend
- Gateway 限流 fail-open（Redis 不可用时放行，优先可用性）

---

## 8. 性能优化：当前状态与改进方向

### 8.1 当前性能瓶颈

`RequireSession` 是 **每个认证请求的热路径**。当前最大问题：

**`ResolveCompanyChargeRate` 在 LRU 缓存之外，每次请求无条件执行 2 次 DB 查询。**

即使 authz LRU 完全命中，仍有 `Company.GetByID` + `Billing.GetCurrency` 两次查询。对于高频 API（如轮询、列表页），这是不必要的重复开销。

### 8.2 推荐改进（按优先级）

#### P0：将 billing rate 纳入缓存

billing currency 和 points_per_unit 是公司级配置，变更频率极低（管理员操作）。应该和 member authz 一起缓存在 LRU 中，以 revision 为失效键。

```go
// 目标：GetSessionContext 在完全命中时 = 0 DB 查询
type cacheValue struct {
    member          types.Member
    permissions     []string
    readOnly        bool
    billingCurrency string  // ← 新增
    pointsPerUnit   int64   // ← 新增
}
```

将 `ResolveCompanyChargeRate` 调用移入 cache miss 分支。当 authz_revision 不变时，billing 信息也不变。

**效果**：热路径 DB 查询从 2 降至 0（revision 5s TTL 内）。

#### P1：revision 查询走 Redis

当前 revision 缓存是进程内 5s TTL。多实例部署时每个实例独立缓存，revision bump 后 5s 窗口内可能返回过期权限。

改为 Redis GET/SET with TTL，所有实例共享同一 revision 视图。bump revision 时 DEL key 即时失效。

**效果**：多实例一致性从 5s 降至亚秒级。单实例场景无区别。

#### P2：`CompanyType` 缓存

`companyTypeFromContext` 在 context 无值时（测试或特殊路径）会 fallback 到 `Company.GetByID`。正常请求路径通过 CompanyResolve middleware 已注入 context，此查询几乎不触发。低优先级，但如果做了 P0 可以顺带收掉。

### 8.3 改进后目标性能

| 场景 | DB 查询 | 备注 |
|------|---------|------|
| 完全命中（同一 member 连续请求） | 0 | revision 5s TTL 内 + LRU 命中 |
| revision miss（首次或 >5s） | 1 | 仅 SELECT authz_revision |
| authz miss（新 member / revision 变更） | 3 | revision + memberAuthz + billing |
| 冷启动 | 3 | 同上 |

对比当前：热路径从 **2 DB/req** 降至 **0 DB/req**，仅 revision TTL 边界有 1 次查询。
