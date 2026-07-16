# Middleware 架构

> 最后更新：2026-07-16

---

## Middleware 链

```
RealIP
  → RequestID              生成/透传 X-Request-Id
  → LoggerContext          注入 per-request slog (携带 request_id)
  → AccessLog              结构化请求日志 (跳过 /healthz)
  → Recover                panic 恢复 + stack trace
  → SecurityHeaders        nosniff / DENY / HSTS
  → CORS                   白名单 origin + Max-Age + Expose-Headers
  ├─ /v1/*    → GatewayService
  │              ├─ parseBearerSecret
  │              ├─ RateLimit/V1 (per-keyHash, token bucket, fail-open)
  │              ├─ precheck (PG + budget Redis)
  │              └─ ReverseProxy (Transport 调优, FlushInterval=-1)
  ├─ /healthz → pass
  └─ /api/*
       ├─ RequestTimeout(30s)
       ├─ CompanyResolve
       ├─ RateLimit/Tenant (per-companyID, token bucket, fail-open)
       ├─ RateLimit/Login  (per-IP, sliding window, fail-closed)
       │     仅匹配: /api/auth/login, /api/auth/accept-invite, /api/platform/auth/login
       ├─ AuthzRevisionHeader
       ├─ CompanyGate (suspended → read-only)
       └─ 子路由 (.With RequireSession / RequireAnyPermission)
```

---

## 各层说明

### 全局层

| Middleware | 文件 | 说明 |
|-----------|------|------|
| RealIP | chi 内置 | 从 X-Forwarded-For / X-Real-IP 取真实客户端 IP |
| RequestID | `middleware/requestid.go` | 生成 16 字符 hex ID，透传 `X-Request-Id` header |
| LoggerContext | `middleware/logger_context.go` | 把 `request_id` 注入 slog，下游用 `LoggerFromContext(ctx)` 取 |
| AccessLog | `middleware/access_log.go` | 响应完成后记录 method/path/status/latency_ms/request_id/company_id/ip，慢请求(>5s)标 `slow=true` |
| Recover | `middleware/recover.go` | defer recover + `runtime/debug.Stack()` + request_id |
| SecurityHeaders | `middleware/security_headers.go` | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy`, HSTS(仅 SECURE_COOKIE=true) |
| CORS | `middleware/cors.go` | 白名单 origin，`Access-Control-Max-Age: 86400`，Expose `X-RateLimit-*` / `X-Authz-Revision` / `Retry-After` |

### /v1 Gateway

| 能力 | 实现位置 |
|------|----------|
| Bearer auth | `gateway/auth.go` parseBearerSecret |
| Per-key rate limit | `gateway/gateway_service.go` checkRateLimit — Token Bucket (Redis Lua), fail-open |
| Precheck | `gateway/precheck.go` — PG CTE + Redis budget check |
| Reverse proxy | `net/http/httputil.ReverseProxy` + Transport 调优 |
| Streaming | `FlushInterval = -1`, 无 WriteTimeout, 无 context timeout |

**Transport 配置：**
- `MaxIdleConnsPerHost: 100`, `ForceAttemptHTTP2: true`
- `ResponseHeaderTimeout: 120s` (防 upstream 挂死)
- `TLS 1.2+`, `DialContext` 超时 10s

### /api 层

| Middleware | 文件 | 说明 |
|-----------|------|------|
| RequestTimeout | `middleware/timeout.go` | `context.WithTimeout(30s)`, 不用 `http.TimeoutHandler` |
| CompanyResolve | `middleware/company_resolve.go` | JWT → companyID → company context；skip `/api/platform/` 和 `/api/internal/` |
| RateLimitTenant | `middleware/rate_limit.go` | Token Bucket per companyID, fail-open；companyID=0 时 skip |
| RateLimitLoginPaths | `middleware/rate_limit.go` | Sliding Window per IP, 仅 POST + 指定路径, fail-closed (本地内存 fallback) |
| AuthzRevisionHeader | `middleware/authz_revision.go` | 返回 `X-Authz-Revision` 供前端缓存失效 |
| CompanyGate | `middleware/company_gate.go` | 租户挂起时拒绝写操作 |
| RequireSession | `middleware/session.go` | JWT 校验 + 会话状态检查，按路由组 `.With()` 挂载 |
| RequireAnyPermission | `middleware/authz.go` | RBAC 权限检查 |

---

## Rate Limiting

### 基建

- **包：** `internal/infra/ratelimit/`
- **接口：** `Limiter` — `AllowTokenBucket` / `AllowSlidingWindow` / `Close`
- **实现：**
  - `RedisLimiter` — Redis Lua 脚本原子操作 (token bucket + sorted set sliding window)
  - `MemoryLimiter` — 本地内存 fixed window，仅用于 login fail-closed fallback
- **共享 helper：** `ratelimit.WriteHeaders` / `ratelimit.WriteRejection` — 写 `X-RateLimit-*` 和 `Retry-After` 响应头

### 策略

| 层级 | Key 格式 | 算法 | 默认阈值 | Fail 策略 |
|------|----------|------|----------|-----------|
| /v1 Per-Key | `rl:v1:{keyHash}` | Token Bucket | 30 req/s burst 60 | fail-open |
| /api Per-Tenant | `rl:api:{companyID}` | Token Bucket | 100 req/s burst 200 | fail-open |
| Login Per-IP | `rl:login:{ip}` | Sliding Window | 5 req/min | fail-closed (内存 fallback) |

### 控制

```bash
RATE_LIMIT_ENABLED=true    # 总开关 (关闭后所有层级跳过)
RATE_LIMIT_DRY_RUN=false   # 观察模式: 只记录不拦截
```

---

## Timeout 策略

| 路径 | 手段 | 值 |
|------|------|-----|
| `/api/*` | `context.WithTimeout` | 30s |
| `/v1/*` | `Transport.ResponseHeaderTimeout` | 120s (仅等首字节) |
| Server | `ReadHeaderTimeout` | 5s |
| Server | `IdleTimeout` | 120s |
| Server | `WriteTimeout` | 不设 (streaming 需要) |

---

## 配置项

```bash
# HTTP
PORT=8080
CORS_ORIGINS=https://app.tokenjoy.com
REQUEST_TIMEOUT_SEC=30
ACCESS_LOG_SLOW_THRESHOLD_MS=5000

# Security
SECURE_COOKIE=true                   # 启用 HSTS + secure cookie

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_DRY_RUN=false
RATE_LIMIT_V1_RATE=30
RATE_LIMIT_V1_BURST=60
RATE_LIMIT_TENANT_RATE=100
RATE_LIMIT_TENANT_BURST=200
RATE_LIMIT_LOGIN_MAX=5
RATE_LIMIT_LOGIN_WINDOW_SEC=60

# Redis (限流 + budget check 共用)
REDIS_URL=redis://localhost:6379
```

---

## 文件清单

```
internal/http/middleware/
├── access_log.go          结构化请求日志
├── authz.go               RequireAnyPermission
├── authz_revision.go      AuthzRevisionHeader
├── company_gate.go        CompanyReadOnlyMiddleware
├── company_resolve.go     CompanyResolve
├── cors.go                CORS
├── logger_context.go      LoggerContext / LoggerFromContext
├── platform_auth.go       PlatformAuth
├── rate_limit.go          RateLimitTenant / RateLimitLogin / RateLimitLoginPaths
├── recover.go             Recover
├── requestid.go           RequestID / RequestIDFromContext
├── routes.go              SessionRoutes / ReadRoutes helpers
├── security_headers.go    SecurityHeaders
├── session.go             RequireSession
├── sync_trigger.go        AllowSyncTrigger
└── timeout.go             RequestTimeout

internal/infra/ratelimit/
├── limiter.go             Limiter 接口 + RedisLimiter
├── memory.go             MemoryLimiter (本地 fallback)
├── open.go               Open() 工厂
├── response.go           WriteHeaders / WriteRejection
└── scripts.go            Lua 脚本 (token bucket + sliding window)

internal/domain/gateway/
├── gateway_service.go     Transport 调优 + per-key rate limit + reverse proxy
├── precheck.go            PrecheckService
└── ...
```
