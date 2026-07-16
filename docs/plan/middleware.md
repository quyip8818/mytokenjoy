# Middleware 整体设计方案

> 创建日期：2026-07-16  
> 状态：Draft

---

## 1. 现状审计

### 当前 Middleware 链

```
RealIP → RequestID → Recover → CORS
  ├─ /v1/*    → GatewayService (bearer auth + precheck + reverse proxy)
  ├─ /healthz → 无保护
  └─ /api/*   → CompanyResolve → AuthzRevision → CompanyGate → 子路由
```

### 已有 ✓

| 能力 | 备注 |
|------|------|
| RealIP / RequestID / Recover / CORS | 完备 |
| 多租户解析 (CompanyResolve) | JWT → companyID → context |
| 认证 (RequireSession / PlatformAuth) | 按路由组挂载 |
| 鉴权 (RequireAnyPermission) | RBAC |
| 租户挂起 (CompanyGate) | read-only 强制 |
| Body size limit | DecodeJSON 1MB / Gateway 4MB |
| Graceful shutdown | 10s timeout |

### 缺失项

| 缺失 | 风险 | 说明 |
|------|------|------|
| **Rate Limiting** | 高 | 登录暴力破解 + /v1 刷量 + 租户超量 |
| **Access Log** | 高 | 无结构化请求日志 |
| **Security Headers** | 中 | 无 HSTS / nosniff / frame-deny |
| **Request Timeout (/api)** | 中 | 无 per-request context deadline |
| **WriteTimeout / IdleTimeout** | 中 | Server 只有 ReadHeaderTimeout=5s |
| **Gateway Transport 调优** | 中 | 默认连接池，无超时，性能受限 |
| **RequestID → Logger** | 低 | 生成了但没注入 slog |
| **CORS Max-Age** | 低 | 缺 preflight 缓存 |

---

## 2. /v1 Gateway 性能分析

### 当前请求生命周期

```
请求到达
  → parseBearerSecret (内存)
  → SHA256 hash key (CPU, ~1μs)
  → MaxBytesReader + ReadAll body (内存)
  → json.Unmarshal 取 model 字段 (CPU)
  → precheck.Run:
      ├─ LoadPrecheckContext: PG 查询 (CTE: platform_keys + companies + model_allowlist)
      └─ budgetRemainCheck: Redis GET (可选，可能 Noop)
  → Evaluate (纯内存逻辑)
  → ReverseProxy → upstream (可能 streaming 数分钟)
```

### 改进方案

#### A. Transport 调优

```go
proxy.Transport = &http.Transport{
    DisableCompression:    true,
    MaxIdleConns:          200,
    MaxIdleConnsPerHost:   100,
    IdleConnTimeout:       90 * time.Second,
    TLSHandshakeTimeout:  5 * time.Second,
    ResponseHeaderTimeout: 120 * time.Second, // LLM 首 token 可能较慢
    ForceAttemptHTTP2:     true,
}
proxy.FlushInterval = -1  // 支持 SSE streaming 实时 flush
```

理由：
- `MaxIdleConnsPerHost=100` — 连接复用，避免每次 TCP+TLS 握手
- `ResponseHeaderTimeout=120s` — 防 upstream 永久挂起，但不截断正常 LLM 响应
- `ForceAttemptHTTP2` — 多路复用减少连接数
- `FlushInterval=-1` — streaming 响应实时推送到客户端

#### B. Precheck PG 查询

当前每次请求查 PG (3 表 CTE)。经评估：
- **不做 Redis 全量缓存** — 字段含 key_status/company_status/wallet_remain，一致性代价太高
- 现有 `CombinedKeyCache` (budget Redis) 已经独立覆盖了高频变化字段
- PG 查询本身是索引查找（key_hash），延迟应在 <5ms
- **等 QPS 真到瓶颈时再加缓存**，到时按稳定字段 vs 易变字段拆分

---

## 3. 限流方案

### 3 层限流（共用基建）

| 层级 | Key | 算法 | 默认阈值 | 位置 | Fail 策略 |
|------|-----|------|----------|------|-----------|
| **/v1 Per-Key** | `rl:v1:{keyHash}` | Token Bucket (Redis Lua) | 30 req/s burst 60 | GatewayService 内部，parseBearerSecret 之后、precheck 之前 | fail-open |
| **/api Per-Tenant** | `rl:api:{companyID}` | Token Bucket (Redis Lua) | 100 req/s burst 200 | /api group, CompanyResolve 之后 | fail-open |
| **Login Per-IP** | `rl:login:{ip}` | Sliding Window (Redis) | 5 req/min | 登录路由独立挂载 | **fail-closed (本地内存 fallback)** |

#### /v1 限流说明

Gateway 是高成本面（每个请求都消耗 LLM 算力/费用）。没有限流 = Transport 调优给滥用方加速。

- 按 `keyHash` 限流（parseBearerSecret → SHA256 hash 即可得到，不需要查 DB）
- 预算/余额控制交给现有 precheck + budgetcheck（它们管"能不能用"）
- rate limit 管"用多快"，两者互补
- 实现为 GatewayService 内部调用（非 chi middleware），因为需要先解析 Bearer token 才能得到 key；但底层复用同一个 `ratelimit.Limiter` 基建

#### Login 限流 fail-closed 说明

暴力破解场景不能 fail-open（Redis 挂了 = 无限试密码）。策略：
1. 正常走 Redis sliding window
2. Redis 不可用时 **降级到本地内存** (sync.Map + 简单计数器)
3. 本地内存限流在多实例下不精确，但仍然有效限制单实例攻击速率

#### 覆盖的登录路由

- `/api/auth/login`
- `/api/platform/auth/login`
- `/api/auth/accept-invite`

#### 不走 Per-Tenant 限流的路由说明

以下路由 `companyID=0`，Per-Tenant 限流不生效（预期行为）：

- **登录/注册** (`/api/auth/login`, `/api/auth/accept-invite`, `/api/platform/auth/login`) — 已被 Login Per-IP 独立保护
- **Internal webhooks** (`/api/internal/*`) — CompanyResolve 对此路径 skip，靠 webhook secret 认证，来源固定（上游服务）
- **SaaS 模式下无 JWT 的请求** — 只能触达上述公开端点

#### 响应头

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1720000000
Retry-After: 2  (仅 429 时)
```

#### 控制

```bash
RATE_LIMIT_ENABLED=true    # 总开关
RATE_LIMIT_DRY_RUN=false   # 观察模式：只记录不拦截
```

---

## 4. Timeout 策略（/api 与 /v1 分离）

| 路径 | Context Timeout | Server WriteTimeout | 理由 |
|------|----------------|--------------------|----|
| `/api/*` | 30s | — | 管理面操作应该快速完成 |
| `/v1/*` | **不设** | — | LLM streaming 可能数分钟，靠 Transport.ResponseHeaderTimeout 控制挂死 |
| Server | — | **0 (不设)** | 因为 /v1 streaming 需要长连接，Server 级别 WriteTimeout 会误杀 |

**防止 /v1 请求挂死的手段：**
- `Transport.ResponseHeaderTimeout=120s` — upstream 120s 内没返回首字节 → 超时
- upstream 返回后 streaming 持续中途断开 → 由客户端断连或 upstream EOF 自然结束
- ReadHeaderTimeout=5s (已有) — 防慢 header 攻击

**http.Server 最终配置：**
```go
server := &http.Server{
    Addr:              ":" + cfg.Port,
    Handler:           application.Router,
    ReadHeaderTimeout: 5 * time.Second,
    IdleTimeout:       120 * time.Second,
    // WriteTimeout 不设 — /v1 streaming 需要
}
```

---

## 5. 其余 Middleware 设计

### 5.1 Access Log

```go
// middleware/access_log.go
// WrapResponseWriter 捕获 status + bytes written
// 响应后记录: method, path, status, latency_ms, request_id, company_id, ip
```

- 慢请求（>5s）标记 `slow=true`
- `/healthz` 路径跳过
- 可选关闭：`ACCESS_LOG_ENABLED=true`
- company_id 从 context 取（CompanyResolve 之后才有）；/v1 路由无 company_id 可记，记 keyHash 前 8 位作为标识
- /v1 streaming 请求 latency 可能数分钟，这是预期行为

### 5.2 Security Headers

```go
// middleware/security_headers.go
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
// HSTS 仅 SECURE_COOKIE=true (即生产 HTTPS) 时:
w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
```

### 5.3 Request Timeout (仅 /api)

```go
// middleware/timeout.go — 仅挂在 /api group
// context.WithTimeout(cfg.RequestTimeoutSec)
// 不用 http.TimeoutHandler（会 buffer response）
```

### 5.4 Logger Context

```go
// middleware/logger_context.go — 把 request_id 注入 slog context
```

### 5.5 现有改进

| 项目 | 改动 |
|------|------|
| Recover | + stack trace + request_id |
| CORS | + `Access-Control-Max-Age: 86400` + Expose `X-RateLimit-*`, `X-Authz-Revision` |

---

## 6. 目标 Middleware 链

```
RealIP
  → RequestID
  → LoggerContext
  → AccessLog
  → Recover
  → SecurityHeaders
  → CORS
  ├─ /v1/* → RateLimit/V1(per-keyHash) → GatewayService (Transport 调优)
  ├─ /healthz → pass
  └─ /api/*
       ├─ RequestTimeout(30s)
       ├─ CompanyResolve
       ├─ RateLimit/Tenant(per-companyID)
       ├─ AuthzRevision
       ├─ CompanyGate
       └─ 子路由 (.With RequireSession / RequireAnyPermission)

  登录路由 (公开，不在 RequireSession 下):
    /api/auth/login            → RateLimit/Login(per-IP, fail-closed)
    /api/platform/auth/login   → RateLimit/Login(per-IP, fail-closed)
    /api/auth/accept-invite    → RateLimit/Login(per-IP, fail-closed)
```

---

## 7. 涉及文件

| 文件 | 操作 |
|------|------|
| `internal/http/middleware/access_log.go` | 新建 |
| `internal/http/middleware/security_headers.go` | 新建 |
| `internal/http/middleware/timeout.go` | 新建 — 仅 /api |
| `internal/http/middleware/rate_limit.go` | 新建 — 3 层共用基建 |
| `internal/http/middleware/logger_context.go` | 新建 |
| `internal/infra/ratelimit/limiter.go` | 新建 — Redis Lua + 本地内存 fallback |
| `internal/domain/gateway/gateway_service.go` | 修改 — Transport 调优 + FlushInterval + rate limit 集成 |
| `internal/http/middleware/recover.go` | 修改 — stack trace + request_id |
| `internal/http/middleware/cors.go` | 修改 — Max-Age + Expose-Headers |
| `internal/http/middleware/requestid.go` | 修改 — 导出 helper |
| `internal/http/router.go` | 修改 — 新链顺序 + /api timeout |
| `internal/config/config.go` | 修改 — 新配置项 |
| `cmd/server/main.go` | 修改 — IdleTimeout + 共享 Redis |

---

## 8. 实施排期

| Phase | 内容 | 预估 | 备注 |
|-------|------|------|------|
| **1** | Security Headers + CORS + Server IdleTimeout | 0.5d | 零风险纯加法 |
| **2** | Access Log + Logger Context + Recover 改进 | 1d | 可观测性基础 |
| **3** | Gateway Transport 调优 + FlushInterval | 0.5d | **v1 性能核心** |
| **4** | Rate Limiting (3 层: v1 per-key + tenant + login) | 2d | 安全性 |
| **5** | Request Timeout (仅 /api, 30s) | 0.5d | |
| **总计** | | **4.5d** | |

---

## 9. 配置汇总

```bash
# Server
REQUEST_TIMEOUT_SEC=30              # 仅 /api

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_DRY_RUN=false
RATE_LIMIT_V1_RATE=30               # /v1 per-key tokens/sec
RATE_LIMIT_V1_BURST=60
RATE_LIMIT_TENANT_RATE=100          # /api per-tenant tokens/sec
RATE_LIMIT_TENANT_BURST=200
RATE_LIMIT_LOGIN_MAX=5              # login per-IP 次/分钟
RATE_LIMIT_LOGIN_WINDOW_SEC=60

# Access Log
ACCESS_LOG_ENABLED=true
ACCESS_LOG_SLOW_THRESHOLD_MS=5000
```

---

## 10. 不做的事

| 项目 | 原因 |
|------|------|
| Precheck context Redis 全量缓存 | 一致性代价太高（key状态/wallet），等 QPS 到瓶颈再做 |
| Global per-IP 限流 | CDN 层职责 |
| 全局 body size middleware | 已有 per-handler 限制 |
| WAF / IP 黑名单 | Cloudflare 处理 |
| Response compression | CDN 处理 |
| Concurrent request limiter | QPS 不高 |
| Server WriteTimeout | /v1 streaming 冲突，不设 |
