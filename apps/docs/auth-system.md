# 认证与授权系统

本文档描述系统当前的认证架构，供 AI 和开发者参考。

---

## 1. 三层身份

| 身份               | 认证方式                                     | 载体                           | 用途                             |
| ------------------ | -------------------------------------------- | ------------------------------ | -------------------------------- |
| **Member Session** | 邮箱+密码 → JWT access token + refresh token | Cookie                         | Web 管理面板所有操作             |
| **Platform Admin** | 同上（super company 下的 member）            | 同上                           | SaaS 平台管理（公司创建/充值等） |
| **Platform Key**   | API 密钥 `sk-xxx`                            | `Authorization: Bearer sk-xxx` | AI 网关 API 调用                 |

---

## 2. Member Session（Access + Refresh Token）

### 2.1 Token 对

| Token         | 载体                                                                 | 有效期                         | 存储                                |
| ------------- | -------------------------------------------------------------------- | ------------------------------ | ----------------------------------- |
| Access Token  | Cookie `tokenjoy_session_member`，Path `/`，SameSite=Lax             | 15 min (`SESSION_TTL_SEC`)     | 无状态 JWT                          |
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

| 场景                       | DB 查询数 | 说明                                 |
| -------------------------- | --------- | ------------------------------------ |
| 完全命中                   | 2         | billing 无缓存（Company + Currency） |
| revision 命中 + authz miss | 3         | 新 member 或 revision 变更           |
| 全 miss                    | 4         | 冷启动、新进程首次请求               |

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

| 变量                          | 默认值                                 | 说明                                       |
| ----------------------------- | -------------------------------------- | ------------------------------------------ |
| `SESSION_SECRET`              | —（必填）                              | JWT 签名密钥（HS256）                      |
| `SESSION_TTL_SEC`             | `900`                                  | Access token 有效期（秒）                  |
| `REFRESH_TOKEN_TTL_SEC`       | `604800`                               | Refresh token 有效期 + cookie MaxAge（秒） |
| `AUTHZ_CACHE_SIZE`            | `4096`                                 | AuthzSvc LRU 缓存条目数（进程内）          |
| `TOKENJOY_COMPANY_ID`         | `00000000-0000-7000-8000-000000000001` | Super company（platform admin 归属）       |
| `PLATFORM_BOOTSTRAP_EMAIL`    | —                                      | 首个 platform admin 邮箱                   |
| `PLATFORM_BOOTSTRAP_PASSWORD` | —                                      | 首个 platform admin 密码                   |
| `SECURE_COOKIE`               | `false`                                | 生产环境设为 `true`（HTTPS only）          |

---

## 6. 代码位置

| 功能                                 | 路径                                                          |
| ------------------------------------ | ------------------------------------------------------------- |
| JWT 签发/解析                        | `internal/identity/sessiontoken/issuer.go`                    |
| Cookie 操作                          | `internal/identity/httpx/token.go`                            |
| Token Pair 签发                      | `internal/identity/httpx/issue.go`                            |
| Session 存储                         | `internal/store/session_repo.go` + `postgres/session_repo.go` |
| Auth handler（Login/Logout/Refresh） | `internal/http/handler/auth/`                                 |
| RequireSession middleware            | `internal/http/middleware/session.go`                         |
| RequirePlatformAdmin middleware      | `internal/http/middleware/require_platform.go`                |
| AuthzSvc（权限缓存）                 | `internal/identity/authz/service.go`                          |
| AuthzSvc LRU 缓存                    | `internal/identity/authz/cache.go`                            |
| Authz Revision header                | `internal/http/middleware/authz_revision.go`                  |
| Gateway（Platform Key 认证）         | `internal/domain/gateway/gateway_service.go`                  |
| Platform Key hash                    | `internal/store/platform_key_mapping_repo.go`                 |

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

---

## 8. 成员状态与生命周期

### 8.1 状态

| 状态 | 值 | 说明 | 可登录 |
|------|------|------|--------|
| 待激活 | `pending` | 已邀请，未完成注册 | 否 |
| 已启用 | `active` | 正常使用 | 是 |
| 已禁用 | `disabled` | 管理员禁用或移除 | 否 |

### 8.2 转换规则

```
pending ──（注册激活）──▶ active ◀──▶ disabled
   │                        │
   ▼                        ▼
 硬删                    disabled（软删）
```

| 从 | 到 | 方式 |
|----|------|------|
| pending | active | 用户完成注册（AcceptInvite） |
| pending | 硬删 | 管理员删除，物理移除 member + invite 记录 |
| active | disabled | 管理员禁用/移除 |
| disabled | active | 管理员启用 |

**禁止**：管理员手动将 pending 改为 active。

### 8.3 删除行为

| 状态 | 行为 |
|------|------|
| pending | 硬删，物理移除 member + invite 记录 |
| active / disabled | 软删，状态改为 `disabled`，API Key 失效，记录保留 |

---

## 9. 邀请与注册流程

### 9.1 核心概念

| 概念 | 说明 |
|------|------|
| **创建来源 (source)** | member 进入系统方式：`manual` / `csv` / `feishu` / `dingtalk` / `wecom` / `invited` |
| **邀请渠道 (inviteChannel)** | 链接分发渠道：`sms` / `email` / `admin_link` |
| **注册渠道 (registrationChannel)** | 用户最终完成注册的渠道（嵌入加密 token） |

### 9.2 邀请流程

一个 member 只有一条 invite 记录（`company_invites`）。不同渠道分发的是同一条 invite 的不同加密 token——token 内嵌 `ch` 字段区分渠道。

```
CreateMember(input):
  1. 创建 member 记录（status = pending）
  2. 创建 invite 记录（生成 invite_code）
  3. if input.phone: encrypt({code, ch=sms}) → 发送短信
  4. if input.email: encrypt({code, ch=email}) → 发送邮件
  5. if 都没有: 等待管理员手动获取链接
```

BatchImport 对每行执行同样逻辑。

### 9.3 加密 Token

XOR 加密（HMAC-SHA256 派生密钥流）+ 4 字节截断 MAC，密钥 `INVITE_SECRET`（`internal/pkg/invitetoken`）。为兼容短信长度限制（阿里云变量 ≤35 字符）刻意做成 18 字符紧凑 token；预留升级路径为 AES-GCM（若字符预算允许 50+ 字符）。

**明文**：9 字节 `[8字节 code][1字节 channel]`（`channel` 为单字符：`s`=sms、`e`=email、`a`=admin_link），**不含** exp——过期只在 DB 侧 `expires_at` 校验。

- Token 不可伪造（HMAC MAC 校验）
- 过期校验依赖 DB `expires_at`
- 一次性使用（accepted_at 标记后不可重复）

### 9.4 AcceptInvite

```
POST /auth/accept-invite { inviteCode, password, name? }
  → 解密 token → code + channel
  → 查 invite 验证未过期未使用
  → 创建/关联 user，设置密码
  → member.status → active，registrationChannel → channel
  → 签发 JWT Session
```

### 9.5 相关 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/org/members` | 创建成员（status=pending + invite + 发送通知） |
| POST | `/org/members/{id}/invite-link` | 管理员获取邀请链接（ch=admin_link） |
| POST | `/org/members/batch-invite` | 批量重发邀请 |
| POST | `/auth/accept-invite` | 注册激活 |

### 9.6 配置

| 环境变量 | 说明 |
|----------|------|
| `INVITE_SECRET` | AES-256 密钥（32 bytes hex），支持多密钥轮转 |
| `INVITE_EXPIRE_HOURS` | 有效时间，默认 168（7天） |

---

## 10. Member/User 数据边界

### 10.1 数据模型

| 表 | 字段 | 说明 |
|----|------|------|
| `users` | name, phone, email, password_hash | 全局用户身份（跨公司） |
| `members` | alias, department_id, employee_id, job_title, hire_date, status, source, registration_channel | 公司内成员身份 |

### 10.2 API 边界

| API | 用途 | 操作目标 |
|-----|------|---------|
| `POST /org/members` | 创建成员 | body `{user:{name,phone,email}, member:{alias,departmentId,...}}` → resolveOrCreateUser + 创建 member |
| `PUT /org/members/:id` | 更新 member 字段 | alias, departmentId, employeeId, jobTitle, hireDate |
| `PUT /org/members/:id/user` | 管理员改 user 字段 | name, phone, email（唯一性冲突报错） |
| `PUT /me/profile` | 用户改自己 | name, avatar, alias |
| `GET /org/members` | 列表 | JOIN users 返回 name/phone/email（只读展示） |

### 10.3 代码位置

| 功能 | 路径 |
|------|------|
| invite token 加解密 | `internal/pkg/invitetoken/` |
| 创建成员 + 邀请 | `internal/domain/org/structure/member_mutate.go` |
| 批量导入 + 邀请 | `internal/domain/org/structure/member_batch.go` |
| AcceptInvite | `internal/domain/company/service_invite.go` |
| 前端激活页 | `apps/frontend/src/routes/auth/invite-accept.tsx` |


---

## 11. SaaS 认证流程

### 11.1 认证策略

SaaS 模式（`SUPPORT_SAAS=true`）下手机号与邮箱两种验证码登录方式**同时并存**于同一登录页，无单一"主认证路径"开关（`AUTH_PRIMARY` 不存在）。用户按拥有的联系方式任选一种验证码登录/注册。

### 11.2 登录分流

`POST /auth/verify-code/verify` 验证码通过后按 membership 数量路由：

| Member 数 | 有邀请？ | action | 行为 |
|-----------|---------|--------|------|
| 1 | — | `enter` | 直接 issueTokenPair |
| ≥2 | — | `select_company` | Register Session → 企业选择页 |
| 0 | 是 | `choose` | Register Session → 邀请选择页 |
| 0 | 否 | `not_found` | 引导去注册 |
| User不存在 | — | `not_found` | 引导去注册 |

### 11.3 注册流程

1. 验证手机/邮箱 → `POST /auth/register/init`（创建/校验 User + Register Session；若命中待处理邀请返回 `choose`）
2. 有邀请 → `POST /auth/register/accept`（接受邀请加入已有企业）；无邀请 → `POST /auth/register/company`（创建 Trial 企业 + issueTokenPair）

创建的是真实企业（`type=trial`），数据永久保留。

### 11.4 多企业选择（登录时）

`POST /auth/select-company { companyId }` 依赖 Register Session（登录/验证码验证后签发的临时 cookie），校验 user 有目标 company 的 active member → issueTokenPair。**不是**运行时切换已登录企业的功能——登录后要换企业需重新登录。

### 11.5 SaaS API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/verify-code/send` | 发送验证码 |
| POST | `/auth/verify-code/verify` | 验证 → 分流（enter/select_company/choose/create_company/not_found） |
| POST | `/auth/select-company` | 登录时多企业选择（依赖 Register Session） |
| POST | `/auth/register/init` | 注册初始化 |
| POST | `/auth/register/accept` | 接受邀请注册 |
| POST | `/auth/register/company` | 创建公司（Trial） |
| GET  | `/api/setup/status` | 私有化 setup 进度检查（独立临时 setup server，非 `/auth` 下） |
| POST | `/api/setup/init` | 私有化初始化（独立临时 setup server） |

### 11.6 OTP 服务

手机验证码和邮箱验证码共用 `identity/verifycode.Service`：

| 配置 | 值 |
|------|------|
| 验证码长度 | 6 位数字 |
| 有效期 | 5 分钟（Redis TTL） |
| 发送间隔 | 60 秒 |
| 每日上限 | 10 次/target |
| 验证尝试 | 最多 5 次，超过锁定 15 分钟 |

Redis 存储（带 channel 维度，`channel` 为 `sms` / `email`）：

```
vc:code:{channel}:{address}      → code, TTL 5min
vc:lock:{channel}:{address}      → "1", TTL 60s
vc:daily:{channel}:{address}     → counter, TTL 到当日 24:00
vc:attempts:{channel}:{address}  → counter, TTL 15min
```

### 11.7 Token 机制

| Cookie | 有效期 | 用途 |
|--------|--------|------|
| `tokenjoy_session_member` | 15 min | Access Token JWT |
| `tokenjoy_refresh` | 7 天 | Refresh Token（DB-backed，Path=/api/auth/refresh） |
| `tokenjoy_register_session` | 10 min | 注册中间态 JWT（仅含 userID） |

---

## 12. Trial / Demo 免费试用

### 12.1 模拟资金

注册时灌入 `LotKindMock` lot（`domain/billing.SeedTrialCredit`）。Gateway 对 `test-model` 有独立准入检查，仅 `demo`/`trial`/`testing` 类型公司可调用；`trial`/`demo` 账户仍可调用真实模型（走 mock lot 消耗）。Mock lot 正常 FIFO 消费，看板可见。

### 12.2 升级（Trial/Demo → Standard）

`domain/company.Service.UpgradeToStandard`（`internal/domain/company/service.go`），事务内：

1. 锁定 company 行（`LockForUpdate`，与并发 ingest 串行）
2. `UpdateType` → `standard`
3. `ExpireMockLots`：过期该企业全部 mock lot
4. `SumActiveLotsRemaining` 重算剩余 → `SetWalletRemainQuota`

commit 后：失效 precheck 缓存（立即拒绝 test-model）+ best-effort 同步 NewAPI wallet。只允许 `trial` 或 `demo` 类型升级，`standard` 类型调用会报错。

### 12.3 功能限制

| 功能 | Trial 行为 |
|------|-----------|
| Gateway | 仅 mock 模型（allowlist） |
| 预算/Key/组织/审计/看板 | 全功能 |
| 充值 | 禁用，提示升级 |
| 成员上限 | 50 人 |

---

## 13. AuthPopup 跨域方案

### 13.1 部署拓扑

`www.tokenjoy.me`（官网）+ App 主域（`/api` 同域路径），共享 `.tokenjoy.me` Cookie Domain。官网通过 iframe 嵌入独立构建入口（`embed.html`），与 App 之间用 `postMessage` 通信，不是独立 npm 包分发。

### 13.2 组件

```tsx
<AuthPopup open defaultMode="login|register" closable onSuccess onClose />
```

- 内部状态机：login tab（phone_verify → enter/select/choose/not_found）+ register tab（phone → info → success）
- 自带 API client（纯 fetch + credentials:include），不依赖 App React Query
- App 内通过 `AuthUnauthorizedBridge` 监听 401 跳转独立 `/login` 路由（**无** `SessionGate` 组件）

### 13.3 前端文件

```
features/auth/
├── components/auth-popup.tsx       — 统一认证弹窗
├── components/auth-card.tsx        — 登录/注册表单卡片
├── components/fake-dashboard.tsx   — 登录背景装饰
└── hooks/use-verify-countdown.ts

routes/auth/login.tsx               — FakeDashboard + AuthPopup(closable=false)
routes/auth/invite-accept.tsx       — 邀请激活页（路由 /invite/accept）
embed-main.tsx / embed.html         — 官网 iframe 嵌入独立构建入口
```

---

## 14. Company 类型与部署模式

### 14.1 type 枚举

| 值 | 含义 | 部署 |
|----|------|------|
| `standard` | SaaS 正式付费 | SaaS |
| `trial` | SaaS 免费试用 | SaaS |
| `demo` | 演示账号 | SaaS |
| `selfhosted` | 私有化部署 | 非 SaaS |
| `testing` | 开发/CI | 开发环境 |

允许 `trial → standard` 与 `demo → standard` 流转（`UpgradeToStandard`）。

### 14.2 SaaS vs 私有化

| 能力 | 私有化 | SaaS |
|------|--------|------|
| 租户数 | 1 | 多 |
| Company 解析 | cfg.CompanyID 兜底 | JWT 必须携带 |
| 自助注册 | 无 | 有 |
| 供应商 Key | 企业自管 | 平台统一管理（企业只读） |
| Platform 管理后台 | 无 | 有 |
