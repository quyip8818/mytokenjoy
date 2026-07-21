# ADR: NewAPI Admin Token — 直接从 NewAPI 数据库读取

## 状态

已接受

## 背景

Backend 通过 `adminport.Port` 接口与 NewAPI 通信，所有请求使用 admin token 作为 Bearer 认证。当前 token 明文存储在 `apps/backend/.env`，存在安全暴露和手动运维问题。

## 决策

**完全移除 `.env` 中的 `NEW_API_ADMIN_TOKEN`。Backend 启动时直接从 NewAPI 的 Postgres 数据库中读取 admin 用户的 `access_token`。**

## 核心洞察

Backend 和 NewAPI 共享同一个 Postgres 实例（相同用户 `tokenjoy`、相同 host），只是不同 database：
- Backend → `tokenjoy` 库
- NewAPI → `newapi` 库

Backend 已有 Postgres 连接凭据（`DATABASE_URL`）。只需连接 `newapi` 库执行一条 SELECT 即可获取永远与 NewAPI 实例一致的 token：

```sql
SELECT access_token FROM users WHERE id = $1;  -- admin_user_id，默认 1
```

## 设计

### 架构

```
Backend 启动
    │
    ▼
连接 newapi 库 → SELECT access_token FROM users WHERE id=1
    │
    ▼
用 access_token 构建 adminPort Client
    │
    ▼
运行时 401 → 重新读 DB → 更新 token → retry
```

### 为什么这个方案最优

| 对比维度 | .env 文件 | system_settings 表 | 直接读 NewAPI DB |
|---------|-----------|-------------------|------------------|
| Token 来源一致性 | 需同步 | 需同步 | **永远一致** |
| 额外凭据 | token 本身 | root 密码 | **无需**（复用 DATABASE_URL 凭据） |
| 新增 schema | 无 | system_settings 表 | **无** |
| reset 后是否自愈 | ❌ | 需 auto-mint | **✅ 天然自愈** |
| 长期运行失效 | token 不会过期 | token 不会过期 | **token 不会过期** |
| 安全性 | 文件系统明文 | 同 DB 安全级别 | **同 DB 安全级别** |

核心优势：**token 的 Source of Truth 就是 NewAPI 的 users 表，直接读取消除了所有同步问题。**

### 连接 NewAPI 数据库

从 `DATABASE_URL` 派生 `newapi` 库连接串：

```
DATABASE_URL = postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable
                                                          ^^^^^^^^
NewAPI DSN   = postgres://tokenjoy:tokenjoy@127.0.0.1:5432/newapi?sslmode=disable
                                                          ^^^^^^
```

实现：替换 URL path 中的 database name。

### 代码改动

#### 1. 新增 token 获取器

```go
// internal/integration/newapi/tokenstore.go
package newapi

// TokenStore reads the admin access_token directly from NewAPI's database.
type TokenStore struct {
    dsn         string  // newapi database DSN
    adminUserID int64
}

func (ts *TokenStore) FetchToken(ctx context.Context) (string, error) {
    conn, err := pgx.Connect(ctx, ts.dsn)
    if err != nil { return "", err }
    defer conn.Close(ctx)

    var token string
    err = conn.QueryRow(ctx,
        `SELECT access_token FROM users WHERE id = $1`,
        ts.adminUserID).Scan(&token)
    return token, err
}
```

#### 2. SelfHealingClient

```go
// internal/integration/newapi/selfhealing.go
type SelfHealingClient struct {
    *Client
    tokenStore *TokenStore
    mu         sync.Mutex
}

// 拦截 401，重新从 NewAPI DB 读取最新 token，retry
func (c *SelfHealingClient) do(ctx context.Context, method, path string, body, out any) error {
    err := c.Client.do(ctx, method, path, body, out)
    if !isUnauthorized(err) {
        return err
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    newToken, mintErr := c.tokenStore.FetchToken(ctx)
    if mintErr != nil {
        return fmt.Errorf("token refresh failed: %w (original: %w)", mintErr, err)
    }
    c.Client.adminToken = newToken
    return c.Client.do(ctx, method, path, body, out)
}
```

#### 3. compose_infra.go 改动

```go
func buildAdminPort(ctx context.Context, cfg config.Config) (adminport.Port, error) {
    if !cfg.NewAPIEnabled || strings.TrimSpace(cfg.NewAPIBaseURL) == "" {
        return nil, nil
    }
    newAPIDSN := deriveNewAPIDSN(cfg.DatabaseURL) // 替换 dbname 为 newapi
    tokenStore := &newapi.TokenStore{DSN: newAPIDSN, AdminUserID: cfg.NewAPIAdminUserID}

    token, err := tokenStore.FetchToken(ctx)
    if err != nil {
        return nil, fmt.Errorf("read NewAPI admin token from DB: %w", err)
    }
    client := newapi.NewClient(cfg.NewAPIBaseURL, token, cfg.NewAPIAdminUserID)
    return newapi.NewSelfHealingClient(client, tokenStore), nil
}
```

#### 4. Config 清理

- 移除 `NewAPIAdminToken` 字段
- 移除 `validateNewAPI()` 中对 `NEW_API_ADMIN_TOKEN` 的校验
- `config.Load()` 中移除对该字段的 TrimSpace

#### 5. Bootstrap 脚本清理

`bootstrap-local-after-reset.sh` 中移除：
- `verify_bootstrap_newapi_admin_token` 调用
- `verify_write_env_var ... NEW_API_ADMIN_TOKEN` 步骤

保留 `verify_newapi_ensure_root`（确保 NewAPI root 账号存在）。

### 前置条件

- NewAPI 必须已启动且数据库已初始化（`ensure-infra.sh` 保证这一点）
- Backend 的 Postgres 用户需有连接 `newapi` 库的权限（当前已满足：同一用户 `tokenjoy` 是 superuser/owner）

### 边界情况

| 场景 | 行为 |
|------|------|
| NewAPI 未启动（DB 不存在） | `FetchToken` 失败，Backend 启动报错（同当前行为） |
| `pnpm reset` 后 | NewAPI 重新 setup → users 表有新 token → Backend 重启时自动读到新值 |
| 运行时 NewAPI token 被 regenerate | 下一次请求 401 → SelfHealingClient 重读 DB → 自动恢复 |
| 生产部署（分离 DB） | 需配置 `NEW_API_DATABASE_URL` env 指向 NewAPI DB |

### 生产环境适配

本方案假设 Backend 和 NewAPI 共享 Postgres 实例。如果生产环境 NewAPI 是独立部署：

```go
type NewAPIConfig struct {
    // ...existing...
    NewAPIDatabaseURL string `env:"NEW_API_DATABASE_URL"` // 可选，生产环境指定
}
```

为空时从 `DATABASE_URL` 派生（本地开发）；有值时直接使用。

## 移除项

- `apps/backend/.env` 中的 `NEW_API_ADMIN_TOKEN`
- `apps/backend/.env.development` 中的相关注释
- `config.Config.NewAPIAdminToken` 字段
- `config.validate` 中的 token 校验
- `bootstrap-local-after-reset.sh` 中写 token 到 .env 的逻辑
- `config.Load()` 中的 `cfg.NewAPIAdminToken = strings.TrimSpace(...)` 

## 实施顺序

1. `integration/newapi/tokenstore.go` — TokenStore 实现
2. `integration/newapi/selfhealing.go` — SelfHealingClient
3. `app/compose_infra.go` — 用 `buildAdminPort` 替代直接 `NewPort`
4. `config/config.go` + `config/validate.go` — 移除 `NewAPIAdminToken`
5. `.env` / `.env.development` — 移除相关行
6. `bootstrap-local-after-reset.sh` — 移除写 token 步骤
7. 验证：`pnpm reset && pnpm start` → 创建公司成功
