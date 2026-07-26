# TokenJoy 与本地 NewAPI 集成文档

## 概述

TokenJoy 使用 [NewAPI](https://github.com/QuantumNous/new-api)（开源 OpenAI 兼容网关）作为 LLM 请求路由和 Token 管理层。NewAPI 负责：
- 路由请求到上游 LLM 提供商（OpenAI、DeepSeek 等）
- API Key 签发与配额管控
- 渠道管理（上游接入点）
- 用量/消耗日志记录
- 模型定价比率配置

TokenJoy 在 upstream commit `bde9b2f4` 基础上应用了 4 个自定义 patch。

## 架构

```
┌────────────────┐         ┌──────────────┐         ┌─────────────────┐
│   TokenJoy     │◄──────►│    NewAPI     │────────►│   上游 LLM      │
│   (port 8010)  │  Admin  │  (port 3010)  │  Proxy  │  (OpenAI/DS等)  │
│                │  API    │              │         │                 │
│   后端 Go      │         │  Go+React    │         └─────────────────┘
└───────┬────────┘         └──────┬───────┘
        │                         │
        │  Webhook (消耗日志)      │
        │◄────────────────────────│
        │                         │
        ▼                         ▼
┌────────────────────────────────────────┐
│          PostgreSQL (port 5510)         │
│  tokenjoy DB | newapi DB | logs DB     │
└────────────────────────────────────────┘
```

## 本地 Docker 部署

**docker-compose.yml** 定义的 NewAPI 服务：

| 服务 | 端口 | 数据库 | 用途 |
|------|------|--------|------|
| newapi-apps | 3010 | newapi | TokenJoy 的 LLM 网关 |
| newapi-sms | 3020 | sms_newapi | SMS 的独立网关实例 |

共享基础设施：
- PostgreSQL: 端口 5510（3 库：tokenjoy, newapi, logs）
- Redis: 端口 6310

## 自定义 Patch

| Patch | 功能 |
|-------|------|
| `0001-management-webhook` | 每次消耗日志插入后 POST webhook 到 Backend |
| `0002-admin-token-contract` | admin 创建 token API 返回完整 token 对象 |
| `0003-username-max-length` | 增加用户名最大长度 |
| `0004-sk-prefix-key-format` | 确保 Key 使用 `sk-` 前缀格式 |

## Admin Token 管理

Backend **不**在 .env 中存储 NewAPI 管理员 token。而是通过 `TokenStore` 直接从 NewAPI 的 `users` 表读取 `access_token` 列：

```
TokenStore → 直接连 newapi 数据库 → 读取 users.access_token
```

DSN 来源：`NEW_API_DATABASE_URL` 环境变量，或从 `DATABASE_URL` 推导（替换数据库名为 `newapi`）。

`SelfHealingPort` 包装层自动检测 401 响应并重新读取 token，无需重启。

## Backend 调用的 NewAPI 接口

### Token 生命周期
| 方法 | 路径 | 用途 |
|------|------|------|
| POST | `/api/token/` | 创建 API token |
| PUT | `/api/token/` | 更新 token |
| GET | `/api/token/{id}` | 获取 token 详情 |
| POST | `/api/token/{id}/key` | 获取 token 明文 key |
| POST | `/api/token/{id}/regenerate` | 重新生成 key |
| DELETE | `/api/token/{id}` | 删除 token |

### 渠道管理
| 方法 | 路径 | 用途 |
|------|------|------|
| POST | `/api/channel/` | 创建渠道 |
| PUT | `/api/channel/` | 更新渠道 |
| GET | `/api/channel/{id}` | 获取渠道 |
| GET | `/api/channel/?p=N&page_size=100` | 分页列表 |
| POST | `/api/channel/fix` | 重建 abilities（模型路由） |
| GET | `/api/channel/sync` | 同步渠道 abilities |

### 定价管理
| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/api/option/` | 获取所有系统选项 |
| PUT | `/api/option/` | 更新系统选项 |
| GET | `/api/pricing` | 列出模型定价 |

### 用户管理
| 方法 | 路径 | 用途 |
|------|------|------|
| POST | `/api/user/` | 创建用户 |
| GET | `/api/user/search?keyword=X` | 按用户名搜索 |
| POST | `/api/user/manage` | 管理用户配额 |
| POST | `/api/user/login` | 登录（bootstrap 脚本用） |
| POST | `/api/setup` | 初始化 root 账户 |
| GET | `/api/user/token` | 获取 admin JWT |
| GET | `/api/status` | 健康检查 |

## adminport.Port 接口

Backend 通过 `adminport.Port` 接口抽象所有 NewAPI 操作：

```go
type Port interface {
    // Token
    CreateToken(ctx, CreateTokenInput) (TokenResult, error)
    UpdateToken(ctx, UpdateTokenInput) (TokenResult, error)
    GetToken(ctx, tokenID int64) (TokenResult, error)
    GetTokenKey(ctx, tokenID int64) (string, error)
    RegenerateToken(ctx, tokenID int64) (TokenResult, error)
    DeleteToken(ctx, tokenID int64) error

    // Channel
    UpsertChannel(ctx, UpsertChannelInput) (ChannelResult, error)
    EnsureGroup(ctx, group, displayName string) error
    RebuildAbilities(ctx) error

    // User
    CreateUser(ctx, CreateUserInput) (UserResult, error)
    ManageUser(ctx, userID int64, action string, value int64) error

    // Pricing
    ListModelPricing(ctx) ([]ModelPricing, error)
    UpdateOption(ctx, key, value string) error
    UpsertModelRatio(ctx, modelType string, inputPrice, outputPrice float64) error
}
```

## 定价模型

存储为 NewAPI 系统选项中的 JSON map：
- `ModelRatio`: `{ "model-name": ratio, ... }`
- `CompletionRatio`: `{ "model-name": ratio, ... }`

**价格转换公式：**
```
modelRatio      = inputPrice / 2
completionRatio = outputPrice / inputPrice

// 反向
inputPrice  = modelRatio * 2      (元/百万token)
outputPrice = modelRatio * completionRatio * 2
```

**写入语义：** read-modify-write merge（读取完整 JSON map → 覆盖目标 key → 写回），不删除未管理的模型。

## 本地 Bootstrap 流程

`pnpm reset` 触发 `bootstrap-local-after-reset.sh`：

1. **启动基础设施** — docker compose up postgres + redis + newapi，创建 `logs.newapi` schema
2. **确保 root 账户** — 尝试用 `root/tokenjoy123` 登录，失败则调用 `/api/setup` 创建
3. **获取 admin JWT** — `GET /api/user/token`
4. **Seed 模型定价目录** — 读取 `lib/model-catalog.json`，调用 Python helper 写入 ModelRatio/CompletionRatio
5. **配置渠道** — 确保 `platform_shared` 组存在，创建 test-model 渠道（mock）+ DeepSeek 渠道

## 本地渠道配置

| 渠道 | 类型 | 上游 | 模型 | 组 |
|------|------|------|------|-----|
| test-model | OpenAI (1) | `http://host.docker.internal:8765` | test-model | platform_shared |
| Deepseek | DeepSeek (43) | 官方 API | deepseek-v4-flash, deepseek-v4-pro | default |

## 模型定价目录（model-catalog.json）

| 模型 | ModelRatio | CompletionRatio | 价格（$/百万token） |
|------|-----------|----------------|-------------------|
| gpt-5.6-sol (+dated) | 2.5 | 6 | $5 / $30 |
| gpt-5.6-terra (+dated) | 1.25 | 6 | $2.50 / $15 |
| gpt-5.6-luna (+dated) | 0.5 | 6 | $1 / $6 |
| claude-fable-5 (+dated, +thinking) | 5.0 | 5 | $10 / $50 |
| kimi-k3 (+thinking) | 1.5 | 5 | $3 / $15 |

## Webhook：消耗日志回传

NewAPI 每次记录消耗日志后，POST 到 Backend：
```
POST http://host.docker.internal:8010/api/internal/webhooks/newapi-log
Header: X-Webhook-Secret: tokenjoy-webhook-secret
Body: { "log_id": 12345 }
```

Backend 收到后触发 ingest pipeline（用量归因、预算扣减等）。

## 环境变量

```bash
# Backend .env
NEW_API_ENABLED=true
NEW_API_BASE_URL=http://127.0.0.1:3010
NEW_API_DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5510/newapi?sslmode=disable
NEW_API_ADMIN_USER_ID=1
NEW_API_WEBHOOK_SECRET=tokenjoy-webhook-secret
NEW_API_GATEWAY_ENABLED=false

# docker-compose.yml (传给 NewAPI 容器)
MANAGEMENT_WEBHOOK_URL=http://host.docker.internal:8010/api/internal/webhooks/newapi-log
MANAGEMENT_WEBHOOK_SECRET=tokenjoy-webhook-secret
```

## 关键文件路径

| 文件 | 用途 |
|------|------|
| `docker-compose.yml` | 基础设施定义 |
| `apps/newapi/Dockerfile` | 带 patch 的 NewAPI 镜像构建 |
| `apps/newapi/scripts/bootstrap-local-after-reset.sh` | 本地 bootstrap 入口 |
| `apps/newapi/scripts/setup-dev-mock-channel.sh` | 渠道创建 |
| `apps/newapi/scripts/seed-model-catalog.sh` | 定价 seed |
| `apps/newapi/scripts/lib/model-catalog.json` | 定价数据源 |
| `apps/newapi/scripts/lib/newapi_admin.py` | Python 管理工具 |
| `apps/backend/internal/integration/newapi/` | Go HTTP 客户端 |
| `apps/backend/internal/domain/adminport/port.go` | Port 接口定义 |
| `apps/backend/internal/integration/newapi/tokenstore.go` | 直连 DB 读 token |
| `apps/backend/internal/integration/newapi/selfhealing.go` | 401 自动重试 |
