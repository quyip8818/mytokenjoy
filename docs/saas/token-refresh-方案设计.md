# Access Token + Refresh Token 方案设计

> **目标**：无感续期 + 服务端吊销能力。  
> **约束**：不引入新依赖，前端仅改 API client 层。

---

## 1. 现状与问题

| 当前实现 | 问题 |
| --- | --- |
| 单 JWT cookie，24h 有效期 | 过期后强制重登，B2B 体验差 |
| 签出仅删 cookie | 无法服务端主动吊销（改密码、踢人无法即时生效） |
| 前端 401 直接跳 login | 无续期能力 |

---

## 2. 方案概述

| Token | 载体 | 有效期 | 存储 | 用途 |
| --- | --- | --- | --- | --- |
| **Access Token** | Cookie `tokenjoy_session_member` | 15 min | 无状态 JWT | 每次 API 鉴权 |
| **Refresh Token** | Cookie `tokenjoy_refresh` | 7 天 | DB（hash） | 换发 access token |

核心原则：
- Access token 短命（15min），泄露窗口小
- Refresh token 不透明字符串，服务端可吊销
- `RequireSession` middleware **不改**
- **不做 Rotation**——refresh token 在 7 天生命周期内不变，refresh 操作是纯读（SELECT + 签发 JWT），天然幂等无竞态

---

## 3. Access Token

与现有 `sessiontoken.Claims` 完全一致，只是 TTL 从 24h 缩短到 15min：

```json
{ "sub": "<memberID>", "company_id": "<companyID>", "user_id": "<userID>", "sid": "<sessionID>", "exp": ... }
```

现有 `RequireSession` middleware 只验签名 + 过期，逻辑不变。

---

## 4. Refresh Token

```
格式：<sessionID>.<randomHex(32)>
```

不是 JWT，服务端通过 DB 查 hash 验证。生命周期内**不变不 rotate**。

---

## 5. 数据库

```sql
CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,           -- sessionID（= JWT 中的 sid）
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    member_id   UUID NOT NULL,
    company_id  UUID NOT NULL,
    token_hash  TEXT NOT NULL,              -- SHA-256(refresh_token)，写入后不变
    user_agent  TEXT DEFAULT '',
    ip          TEXT DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,       -- refresh 过期时间
    revoked_at  TIMESTAMPTZ                 -- 非空 = 已吊销
);

CREATE INDEX idx_sessions_user_active ON sessions(user_id) WHERE revoked_at IS NULL;
```

---

## 6. Cookie 配置

| Cookie | Path | HttpOnly | Secure | SameSite | Max-Age |
| --- | --- | --- | --- | --- | --- |
| `tokenjoy_session_member` | `/` | ✅ | 生产 ✅ | Lax | 不设（session cookie，由 JWT exp 控制） |
| `tokenjoy_refresh` | `/api/auth/refresh` | ✅ | 生产 ✅ | Strict | 604800 (7d) |

Access cookie 保持 session cookie（浏览器关闭即消失）。用户重新打开浏览器时，首次请求 401 → refresh → 拿到新 access token，体验连贯。

---

## 7. API

### 7.1 `POST /auth/refresh`

```go
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
    raw := resolveRefreshCookie(r)
    if raw == "" {
        writeUnauthorized(w)
        return
    }

    sid, _, ok := splitRefreshToken(raw) // "sid.random"
    if !ok {
        writeUnauthorized(w)
        return
    }

    sess, err := h.store.Sessions.GetActive(ctx, sid) // WHERE revoked_at IS NULL AND expires_at > now()
    if err != nil || sess == nil {
        writeUnauthorized(w)
        return
    }

    if sha256Hex(raw) != sess.TokenHash {
        writeUnauthorized(w)
        return
    }

    // 签发新 access token，refresh token 不变
    accessToken, _ := h.issuer.IssueWithSid(sess.CompanyID, sess.MemberID, sess.UserID, sid)
    httpx.SetSessionCookie(w, accessToken, h.secureCookie)
    w.WriteHeader(http.StatusNoContent)
}
```

**关键点**：无 UPDATE，无 rotation。SELECT 验证 → 签发 JWT。多实例并发调用完全安全。

### 7.2 登录签发双 token

```go
func issueTokenPair(w http.ResponseWriter, r *http.Request, issuer sessiontoken.Issuer, st store.Store, secureCookie bool, companyID, memberID, userID uuid.UUID) error {
    sid := newSessionID()
    refreshRaw := sid + "." + randomHex(32)

    st.Sessions.Create(ctx, Session{
        ID:        sid,
        UserID:    userID,
        MemberID:  memberID,
        CompanyID: companyID,
        TokenHash: sha256Hex(refreshRaw),
        UserAgent: r.UserAgent(),
        IP:        realIP(r),
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
    })

    accessToken, _ := issuer.IssueWithSid(companyID, memberID, userID, sid)
    httpx.SetSessionCookie(w, accessToken, secureCookie)
    setRefreshCookie(w, refreshRaw, secureCookie)
    return nil
}
```

### 7.3 登出

```go
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    if claims, ok := httpx.ResolveMemberClaims(r, h.issuer); ok {
        h.store.Sessions.Revoke(ctx, claims.Sid) // UPDATE SET revoked_at = NOW()
    }
    httpx.ClearSessionCookie(w)
    clearRefreshCookie(w)
    w.WriteHeader(http.StatusNoContent)
}
```

---

## 8. 前端改动

仅改 `apps/frontend/src/api/client.ts`：

```typescript
let refreshing: Promise<boolean> | null = null

function doRefresh(): Promise<boolean> {
  if (!refreshing) {
    refreshing = fetch(`${API_BASE_PATH}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
      .then((r) => r.ok)
      .catch(() => false)
      .finally(() => { refreshing = null })
  }
  return refreshing
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_PATH}${path}`
  const init: RequestInit = {
    credentials: 'include',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json', ...options.headers },
    ...options,
  }

  let res = await fetch(url, init)
  notifyAuthzRevision(res)

  // 401 且不是 refresh 自身 → 尝试续期一次
  if (res.status === 401 && path !== '/auth/refresh') {
    const ok = await doRefresh()
    if (ok) {
      res = await fetch(url, init)
      notifyAuthzRevision(res)
    }
  }

  if (!res.ok) {
    // ... 现有错误处理不变，401 时 emit('unauthorized')
  }

  return readJsonBody<T>(res)
}
```

并发 10 个请求同时 401 → 共享同一个 `refreshing` Promise → refresh 一次 → 各自重试。

---

## 9. 后端改动清单

| 模块 | 变更 |
| --- | --- |
| `sessiontoken` 包 | 新增 `IssueAccessToken(secret, ttl, companyID, memberID, userID, sid)` 纯函数；`Issuer` interface **不改** |
| `identity/httpx/token.go` | 新增 `SetRefreshCookie` / `ClearRefreshCookie` / `ResolveRefreshCookie` |
| `store/session_repo.go` | 新增 `Create` / `GetActive` / `Revoke` / `RevokeAllByUser` |
| `http/handler/auth/refresh.go` | 新增 refresh handler |
| `http/handler/auth/handler.go` | Login / AcceptInvite 改为调用 `issueTokenPair` |
| `http/handler/auth/handler.go` | Logout 追加 revoke + clear refresh cookie |
| Config | `SESSION_TTL_SEC` 默认值 86400 → 900 |

`RequireSession` middleware **不改**。`Issuer` interface **不改**（PlatformSessionToken 不受影响）。

### 9.1 为什么不改 `Issuer` interface

现有 `IssueWithUser` 内部生成 sid 并返回 token string。改 interface 会影响 `PlatformSessionToken`（SaaS 管理后台 token，不需要 refresh）。

新增一个纯函数 `IssueAccessToken` 解决：

```go
// sessiontoken/issue.go — 新增纯函数，不在 interface 上
func IssueAccessToken(secret []byte, ttl time.Duration, companyID, memberID, userID uuid.UUID, sid string) (string, error) {
    claims := Claims{
        CompanyID: companyID,
        UserID:    userID,
        Sid:       sid,
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   memberID.String(),
            IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
            ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(ttl)),
        },
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
```

`issueTokenPair` 和 refresh handler 都调用这个函数。现有 `IssueWithUser`（内部自动生成 sid）保持原样给其他不需要 refresh 的场景使用。

---

## 10. 迁移策略

1. 后端部署：建表 + refresh 端点 + 登录签发双 token（access TTL 降为 15min）
2. 旧 24h token 仍然有效（JWT exp 未到之前 middleware 照常放行）
3. 前端部署：client.ts 加 refresh 逻辑
4. 24h 后旧 token 全部自然过期，完成切换

零停机，无需协调发版顺序。

---

## 11. 安全

| 威胁 | 缓解 |
| --- | --- |
| XSS | 两个 cookie 都是 HttpOnly |
| CSRF | Refresh cookie SameSite=Strict + Path 限定 |
| Refresh token 泄露 | HttpOnly + 仅 HTTPS + Path 仅 `/api/auth/refresh`；管理员可 revoke session |
| 中间人 | Secure flag |
| 管理员踢人 | revoke session → 最多 15min 后 access 过期，refresh 被拒 |

---

## 12. 不做什么

| 排除 | 理由 |
| --- | --- |
| Refresh Token Rotation | 增加写操作和并发复杂度，B2B 场景 HttpOnly cookie 泄露概率极低 |
| 每次请求检查 session revoke | 过度设计，15min 窗口可接受 |
| Sessions 管理 API（列表/踢端） | YAGNI，需要时再加（纯增量） |
| Redis | PG 够用 |
| Sliding window | 每请求 Set-Cookie，缓存不友好 |

---

## 13. 配置项

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `SESSION_SECRET` | — | 已有 |
| `SESSION_TTL_SEC` | `900` | Access token TTL（原 86400） |
| `REFRESH_TOKEN_TTL_SEC` | `604800` | Refresh token TTL |

---

## 14. SaaS vs Local（私有部署）：无差异

Refresh 机制对两种部署模式**一视同仁**，不需要 `if cfg.SupportSaas` 分叉。原因：

| 关注点 | SaaS | Local | 为何无影响 |
| --- | --- | --- | --- |
| 登录时 companyID 来源 | 前端 body 传入 | `CompanyResolve` middleware 用 `LocalCompanyID` | `issueTokenPair` 在 companyID 确定之后调用 |
| refresh 时 company 解析 | JWT 过期 → `CompanyResolve` 拿不到 | 兜底到 `LocalCompanyID` | refresh handler 不依赖 `companyCtx`，从 sessions 表读 |
| Platform token（SaaS 管理后台） | 独立 issuer + cookie | 不存在 | 不涉及 refresh，`Issuer` interface 不改 |
| 多租户隔离 | sessions 表有 `company_id` | 单租户 | 同结构，无特殊处理 |

---

## 15. 工作量

| 内容 | 估时 |
| --- | --- |
| sessions 表 + store 层 | 0.5d |
| refresh handler + issueTokenPair + logout 改造 | 1d |
| `IssueAccessToken` 纯函数 + cookie helpers | 0.5d |
| 前端 client.ts 改造 | 0.5d |
| 过期 session 清理（River periodic job） | 0.5d |

**总计约 3 个工作日**。
