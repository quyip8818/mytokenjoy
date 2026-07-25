# NewAPI 集成设计方案

## 背景

- **SMS**（本项目 `sms/`）：AI 模型供应商管理系统，维护模型目录和采购价格
- **TokenJoy**（`apps/`）：AI API 网关平台，内置 NewAPI（OpenAI 兼容代理），管理用户 token、channel 和计费

目标：SMS 成为模型价格的唯一数据源（SOT），自动同步定价到 SMS 的 NewAPI 实例。

## 部署约束

- SMS 和 NewAPI 在同一 monorepo（mytokenjoy），共用统一 docker-compose
- 共用同一个 Postgres 实例（端口 5510），三个数据库：`sms`、`sms_newapi`、`sms_logs`
- 数据库隔离原则：SMS 只在获取 admin PAT 时访问 sms_newapi 数据库，其他操作走 HTTP API
- 不需要 patch NewAPI，使用原版

## 架构

```
                      同一台机器 / 同一个 Postgres 实例（端口 5510）
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│  sms backend (:8020)                    NewAPI-SMS (:3020)        │
│  ┌──────────────────┐                   ┌──────────────────┐    │
│  │                  │ ── HTTP API ──────>│  /api/option/    │    │
│  │  newapisync svc  │    (localhost)     │  (RootAuth)      │    │
│  │                  │                    └──────────────────┘    │
│  └────────┬─────────┘                             │              │
│           │                                       │              │
│           │ 仅读 PAT（401 时自动刷新）               │              │
│           ▼                                       ▼              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    PostgreSQL (:5510)                      │   │
│  │  ┌─────────┐    ┌─────────────┐    ┌──────────────────┐  │   │
│  │  │ sms DB  │    │ sms_newapi   │    │   sms_logs       │  │   │
│  │  │         │    │ users/tokens │    │ newapi schema:   │  │   │
│  │  │ models  │    │ channels     │    │   consume_log    │  │   │
│  │  │ ...     │    │ options      │    │                  │  │   │
│  │  └─────────┘    └─────────────┘    └──────────────────┘  │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

**数据库分工**：
- `sms` — SMS 后端业务数据（供应商、模型目录、合同、采购单等）
- `sms_newapi` — NewAPI 自身数据（users、tokens、channels、options）
- `sms_logs` — NewAPI 的 consume_log（schema: `newapi`），未来 SMS 可读取做用量分析

## 数据模型变更

### models 表新增 `model_id` 字段

SMS 现有 `model_name` 是显示名（如"GPT-4o"、"Claude 3.5 Sonnet"），但 NewAPI 的 ratio map 使用标准 model identifier（如 `gpt-4o`、`claude-3-5-sonnet-20241022`）。

**必须分离这两个概念**，否则同步链路无法命中：

```sql
ALTER TABLE models ADD COLUMN model_id VARCHAR(128);
-- model_id: NewAPI 使用的标准标识符，用于 ratio map 的 key
-- model_name: 保持不变，用于 UI 显示
```

| 字段 | 用途 | 示例 |
|------|------|------|
| `model_id` | NewAPI ratio key，同步时使用 | `gpt-4o-2024-11-20` |
| `model_name` | 管理员可读的显示名 | `GPT-4o` |

前端创建/编辑模型时新增 `model_id` 输入框，同步逻辑以 `model_id` 为准。`model_id` 为空的模型跳过同步。

## Token 获取策略

从 newapi 数据库的 `users` 表读取 root 用户的 Personal Access Token（PAT）：

```sql
SELECT access_token FROM users WHERE id = $1  -- admin_user_id，默认 1
```

NewAPI 的 PAT 是持久化的，不会过期（除非用户手动 regenerate）。认证时作为 `Authorization: Bearer <access_token>` 发送，NewAPI 通过 `ValidateAccessToken` 直接查 DB 校验。

**读取策略**：
1. 启动时从 newapi DB 读取一次，缓存在内存
2. HTTP client 收到 **401** 响应时，自动重新从 DB fetch 并 retry 一次（应对 regenerate 场景）
3. retry 仍然 401 → 返回错误，需人工介入

**DSN 推导**：从 SMS 的 `DATABASE_URL` 替换 dbname 为 `sms_newapi`：
```
postgres://sms:sms@localhost:5510/sms → postgres://sms:sms@localhost:5510/sms_newapi
```

## 后端设计

### 新增文件

```
internal/domain/newapisync/
  service.go        -- 同步逻辑 + pricing 换算
  ports.go          -- AdminPort interface

internal/integration/newapi/
  client.go         -- HTTP client 实现 AdminPort（含 401 自动刷新 PAT）
  tokenstore.go     -- 从 newapi DB 读取 root 用户 PAT
```

### AdminPort 接口

```go
package newapisync

type AdminPort interface {
    // UpsertModelRatio 写入单个模型的定价到 NewAPI
    UpsertModelRatio(ctx context.Context, modelID string, inputPrice, outputPrice float64) error
    // SyncAllPricing 全量同步所有模型定价
    SyncAllPricing(ctx context.Context, entries []PricingEntry) error
    // ListCurrentRatios 读取 NewAPI 当前的 ModelRatio + CompletionRatio
    ListCurrentRatios(ctx context.Context) ([]ModelPricing, error)
}

type PricingEntry struct {
    ModelID     string  // NewAPI 的 model identifier
    InputPrice  float64 // 元/1M tokens
    OutputPrice float64
}

type ModelPricing struct {
    ModelID         string
    ModelRatio      float64
    CompletionRatio float64
    InputPrice      float64 // 换算后展示价
    OutputPrice     float64
}
```

### 定价换算

与 mytokenjoy 保持一致：

```go
// 元/1M tokens → NewAPI ratio
modelRatio      = inputPrice / 2
completionRatio = outputPrice / inputPrice   // 0 if inputPrice == 0

// NewAPI ratio → 元/1M tokens
inputPrice  = modelRatio * 2
outputPrice = modelRatio * completionRatio * 2
```

### HTTP Client 实现要点

**只使用 `/api/option/` 端点**（RootAuth），不依赖 `/api/pricing`：
- `GET /api/option/` — 读取所有 options，提取 `ModelRatio` 和 `CompletionRatio` 两个 JSON map
- `PUT /api/option/` — 写入更新后的 map

不用 `/api/pricing` 的原因：该端点受 `HeaderNavModuleAuth("pricing")` 控制，取决于后台是否开启 pricing 模块。`/api/option/` 是 root 级端点，无此限制，且数据更完整。

**认证**：`Authorization: Bearer <PAT>`，PAT 来自 TokenStore（401 时自动刷新）。

### SMS 后端新增 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/newapi/status` | 对比本地 vs NewAPI 价格，返回差异列表 |
| POST | `/api/newapi/sync` | 全量同步：SMS 所有有 model_id 的模型价格 → NewAPI |

### model service 集成

模型创建/更新价格时触发异步同步：

```go
func (s *Service) Create(ctx context.Context, input CreateInput) (*types.AiModel, error) {
    // ... 现有逻辑 ...
    // 有 model_id 且有价格时才同步
    if s.syncer != nil && input.ModelID != "" && (input.InputPrice != nil || input.OutputPrice != nil) {
        go s.syncer.UpsertModelRatio(context.Background(), input.ModelID, ...)
    }
    return m, nil
}
```

## 前端设计

在 Models 页面增加 Tab：

```
features/models/
  models-page.tsx                 -- 加 Tab 切换
  components/
    newapi-sync-panel.tsx         -- 同步面板
    pricing-diff-table.tsx        -- 差异对比表格
  hooks/
    use-newapi-sync.ts            -- 同步相关 hooks
```

UI 草图：

```
┌───────────────────────────────────────────────────────────────────┐
│  [模型列表]  [NewAPI 同步]                                         │
├───────────────────────────────────────────────────────────────────┤
│  NewAPI 定价同步                               [全量同步] 按钮      │
├───────────────────────────────────────────────────────────────────┤
│  模型ID          显示名    本地(入/出)   NewAPI(入/出)   状态       │
│  gpt-4o         GPT-4o   ¥60/¥120    ¥60/¥120      ✓ 一致      │
│  claude-3-5-... Claude   ¥45/¥90     ¥45/¥75       ⚠ 不一致     │
│  deepseek-chat  DeepSeek ¥2/¥4       —             ❌ 未同步     │
│  (无 model_id)  Qwen     ¥10/¥20     —             ⏭ 跳过       │
└───────────────────────────────────────────────────────────────────┘
```

模型编辑对话框新增 `model_id` 字段（标注"NewAPI 模型标识符，留空则不同步"）。

## 同步策略

| 场景 | 行为 |
|------|------|
| 创建/更新模型（有 model_id + 有价格） | 异步推送单条（fire & forget） |
| 手动"全量同步" | 读取当前 ratio map → merge SMS 模型 → 写回 |
| 查看同步状态 | 读 `/api/option/` 拿 ratio map，与本地对比 |
| NewAPI 不可达 | UI 提示失败，不阻塞本地操作 |
| admin token 过期（401） | 自动从 DB 刷新 PAT 并 retry |

### 写入语义：Merge，不是覆盖

NewAPI 的 `ModelRatio` / `CompletionRatio` 是全局 JSON map，包含 NewAPI 自身内置的默认 ratio 以及其他来源设置的值。SMS 的同步采用 **read-modify-write merge**：

1. `GET /api/option/` → 读取当前完整 map
2. 遍历 SMS 有 `model_id` 的模型，用 SMS 价格**覆盖对应 key**
3. 不删除 map 中 SMS 未管理的 key（保留 NewAPI 其他模型的设置）
4. `PUT /api/option/` → 写回合并后的 map

这样 SMS 只管它认识的模型，不影响 NewAPI 的其他定价配置。

### 生效延迟

NewAPI 有 `SYNC_FREQUENCY`（默认 60s）从 DB 重新加载 options 到内存。通过 `PUT /api/option/` 写入后，NewAPI 会**立即更新内存**（`updateOptionMap` 是同步调用的），无需等待周期刷新。

## 安全性

| 风险 | 缓解 |
|------|------|
| admin PAT 权限过高（root 级） | 不暴露给前端；同一台机器无网络暴露 |
| read-modify-write 竞争 | SMS 是 SOT，覆盖即期望行为 |
| NewAPI CVE（SQL LIKE 注入等） | SMS 传入的 model_id 来自管理员创建，无外部注入路径 |
| DB 跨库访问 | 仅 `SELECT access_token`（PAT）一条查询，只读 |
| token 失效 | 401 自动刷新 PAT，无需人工重启 |

## 基础设施编排

SMS 的基础设施由 monorepo 根目录的统一 `docker-compose.yml` 管理。包含一个共享 Postgres（端口 5510）、Redis（端口 6310），以及两个 NewAPI 实例（apps:3010, sms:3020）。

**启动命令**：
```bash
pnpm infra            # 启动全部基础设施
pnpm start sms        # 启动 sms backend + frontend
pnpm reset sms        # 重置 sms 数据库
```

SMS 相关数据库（sms、sms_newapi、sms_logs）由 `scripts/postgres-init/01-create-all-dbs.sh` 在容器首次启动时自动创建，owner 为 `sms` 用户。

## 配置

```env
# sms/backend/.env.development
DATABASE_URL=postgres://sms:sms@localhost:5510/sms?sslmode=disable
NEWAPI_BASE_URL=http://localhost:3020
NEWAPI_ADMIN_USER_ID=1
# NEWAPI_DATABASE_URL 可选，默认从 DATABASE_URL 推导（替换 dbname 为 sms_newapi）
```

当 `NEWAPI_BASE_URL` 为空时，同步功能整体禁用，model service 正常工作不受影响。

## 实施步骤

1. 根目录 `docker-compose.yml`（已就绪）— 基础设施
2. `models` 表加 `model_id` 字段 + 前端编辑表单适配
3. 后端 `internal/integration/newapi/tokenstore.go` — 从 newapi DB 读 PAT
4. 后端 `internal/integration/newapi/client.go` — HTTP client（含 401 retry）
5. 后端 `internal/domain/newapisync/` — 同步 service
6. 后端 handler + 路由注册
7. 后端 model service 钩子（创建/更新时触发同步）
8. 前端 `api/newapi.ts` + models 页面 Tab + 同步面板

## 最终目录结构（新增/改动）

```
mytokenjoy/sms/
├── backend/
│   ├── schema.sql                          (改) models 表加 model_id
│   └── internal/
│       ├── config/
│       │   └── config.go                   (改) 加 NEWAPI_* 配置
│       ├── app/
│       │   └── app.go                      (改) 组装 newapisync
│       ├── domain/
│       │   ├── model/
│       │   │   └── service.go              (改) 价格变更时触发同步
│       │   ├── newapisync/
│       │   │   ├── ports.go
│       │   │   └── service.go
│       │   └── types/
│       │       └── models.go               (改) AiModel 加 ModelID 字段
│       ├── integration/
│       │   └── newapi/
│       │       ├── client.go               HTTP client（含 401 refresh PAT）
│       │       └── tokenstore.go           从 sms_newapi DB 读 root 用户 PAT
│       └── http/
│           └── handler/
│               ├── newapisync/
│               │   └── handler.go
│               └── register.go             (改)
│
└── frontend/src/
    ├── api/
    │   └── newapi.ts
    └── features/
        └── models/
            ├── index.ts                    (改)
            ├── models-page.tsx             (改) Tab + model_id 字段
            ├── components/
            │   ├── newapi-sync-panel.tsx
            │   └── pricing-diff-table.tsx
            └── hooks/
                └── use-newapi-sync.ts
```

## 不做的事

- 不管 token/channel/key 生命周期
- 不做 webhook 接收
- 不做定时自动同步（手动 + 创建时自动够用）
- 不直接写 newapi 数据库（走 HTTP API）
- 不用 `/api/pricing` 端点（依赖模块开关）
- 不打 patch
