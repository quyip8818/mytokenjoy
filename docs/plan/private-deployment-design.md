# 私有化部署设计方案

## 架构总览

```
┌──────────────────────────────────────────────┐
│              云端（你管理）                     │
│  ┌────────┐   ┌──────────────────────────┐   │
│  │ NewAPI  │◄──│ Model Config Service     │   │
│  │(源数据) │   │ (Go+React+PostgreSQL)    │   │
│  └────┬───┘   └────────────┬─────────────┘   │
│       │                    │                  │
└───────┼────────────────────┼──────────────────┘
        │                    │ GET /api/v1/catalog (每10min pull)
        │                    ▼
┌───────┼──────────────────────────────────────────────┐
│       │        客户私有化环境（单租户）                  │
│       │  ┌────────────────┐    ┌───────────────────┐ │
│       │  │ Tokenjoy       │    │ NewAPI（本地,隐藏）  │ │
│       │  │ ├─ Gateway ────┼───►│ ├─ platform channel│─┼─► 云端 NewAPI → Provider
│       │  │ ├─ sync worker │    │ └─ custom channel ─│─┼─► Provider 直连
│       │  │ └─ ingest      │◄───│   (webhook)        │ │
│       │  └────────────────┘    └───────────────────┘ │
└───────┼──────────────────────────────────────────────┘
```

## 核心决策摘要

| 决策项 | 结论 |
|--------|------|
| 部署拓扑 | 集中托管：Model Config Service + 云端 NewAPI 在你的云端 |
| 客户侧 LLM 路由 | 本地 NewAPI 统一入口，channel 区分平台/自有 |
| 模型判定粒度 | Per-model，models 表 `source` 字段（platform/custom） |
| 计费规则 | 平台模型：扣预算 + 扣钱包；自有模型：只扣预算 |
| 价格存储 | 两边都写（models 表 + 本地 NewAPI ModelRatio） |
| 同步方式 | 客户侧 pull，10 分钟间隔 |
| 用量记录 | 本地 NewAPI webhook，不依赖远端日志 |
| Gateway | 单一 NEW_API_BASE_URL，precheck 按 source 跳过钱包检查 |
| Ingest | 查 models 表 source，custom 跳过 ConsumeLotsLocked |
| 下架策略 | 软删除（enabled=false）+ 现有 org_node fallback 兜底 |
| 认证 | 客户专属 key，一个 key 用于 LLM 代理 + catalog pull |
| Model Config Service | 独立服务，Go + React + PostgreSQL |

---

## Phase 1: Model Config Service

### 1.1 产品定位

Model Config Service 是模型目录的**唯一管理入口**，对 NewAPI 做一层抽象封装：
- 从 NewAPI 导入原始模型数据
- 人工编辑（选择性发布、价格加成、元信息覆写、分组排序）
- 统一定价后通过 API 下发给所有私有化客户

### 1.2 模块划分

```
model-config-service/
├── cmd/server/              — 入口
├── internal/
│   ├── config/              — 环境变量配置
│   ├── domain/
│   │   ├── catalog/         — 模型目录管理（核心）
│   │   ├── pricing/         — 定价策略
│   │   ├── customer/        — 客户管理 + 认证
│   │   └── sync/            — NewAPI 数据同步
│   ├── http/handler/        — HTTP handlers
│   ├── store/postgres/      — 数据持久化
│   └── integration/newapi/  — NewAPI HTTP 客户端（可复用 tokenjoy 的）
├── web/                     — React 前端
└── docker-compose.yml
```

### 1.3 数据模型设计

#### models（模型目录）

| 字段 | 类型 | 说明 |
|------|------|------|
| model_id | UUID | 主键 |
| type | TEXT | 模型标识，如 `gpt-4o`（唯一） |
| name | TEXT | 显示名称（可覆写） |
| provider | TEXT | openai / anthropic / deepseek / qwen |
| input_price | NUMERIC | 输入价格（元/1M tokens） |
| output_price | NUMERIC | 输出价格（元/1M tokens） |
| max_context | INT | 最大上下文窗口 |
| max_tokens | INT | 最大输出 tokens |
| capabilities | TEXT[] | vision, function_calling, reasoning 等 |
| group_name | TEXT | 分组标签（推荐/经济型/高性能） |
| sort_order | INT | 展示排序 |
| description | TEXT | 模型描述 |
| status | TEXT | active / deprecated / removed |
| source_ratio | NUMERIC | NewAPI 原始 model_ratio（成本参考） |
| source_completion_ratio | NUMERIC | NewAPI 原始 completion_ratio |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

#### customers（客户）

| 字段 | 类型 | 说明 |
|------|------|------|
| customer_id | UUID | 主键 |
| name | TEXT | 客户名称 |
| api_key | TEXT | 客户专属 key（用于 catalog pull + LLM 代理认证） |
| status | TEXT | active / suspended |
| last_pull_at | TIMESTAMPTZ | 上次拉取 catalog 的时间 |
| created_at | TIMESTAMPTZ | |

#### sync_logs（同步日志）

| 字段 | 类型 | 说明 |
|------|------|------|
| log_id | UUID | 主键 |
| direction | TEXT | from_newapi / to_customer |
| status | TEXT | success / failed |
| model_count | INT | 涉及模型数量 |
| error_message | TEXT | 失败原因 |
| created_at | TIMESTAMPTZ | |

### 1.4 API 设计

#### 管理端 API（React 前端调用）

```
# 模型目录
POST   /api/admin/sync              — 触发从 NewAPI 同步
GET    /api/admin/models             — 模型列表（含未发布）
PUT    /api/admin/models/:id         — 编辑模型（名称/价格/分组/排序）
PUT    /api/admin/models/:id/status  — 发布/下架/废弃
DELETE /api/admin/models/:id         — 删除（仅 removed 状态可删）

# 定价管理
GET    /api/admin/pricing            — 价格总览（成本 vs 售价）
PUT    /api/admin/pricing/batch      — 批量调价（加成比例）

# 客户管理
GET    /api/admin/customers          — 客户列表
POST   /api/admin/customers          — 创建客户 + 生成 key
PUT    /api/admin/customers/:id      — 编辑客户
DELETE /api/admin/customers/:id      — 停用客户
POST   /api/admin/customers/:id/rotate-key — 轮换 key

# 同步状态
GET    /api/admin/sync/status        — 上次同步时间 + 各客户 pull 状态
GET    /api/admin/sync/logs          — 同步日志列表
```

#### 客户端 API（客户侧 Tokenjoy 调用）

```
GET /api/v1/catalog
Authorization: Bearer <customer_api_key>
Query: ?version=<last_known_version>

Response 200:
{
  "version": "v20240724-abc123",
  "models": [
    {
      "type": "gpt-4o",
      "name": "GPT-4o",
      "provider": "openai",
      "input_price": 3.0,
      "output_price": 9.0,
      "max_context": 128000,
      "max_tokens": 16384,
      "capabilities": ["vision", "function_calling"],
      "group": "推荐",
      "sort_order": 1,
      "status": "active"
    }
  ]
}

Response 304: (version 未变化，无需更新)
```

### 1.5 前端页面设计

#### 页面 1：模型目录

- 顶部操作栏：「从 NewAPI 同步」按钮 + 上次同步时间
- 表格列：模型标识(type) | 显示名称 | Provider | 分组 | 状态 | 输入价格 | 输出价格 | 操作
- 状态筛选 Tab：全部 / 已发布(active) / 已下架(deprecated) / 未发布(removed)
- 行操作：编辑、发布/下架
- 批量操作：批量发布、批量下架
- 支持拖拽排序（sort_order）

#### 页面 2：定价管理

- 成本对照表：模型 | NewAPI 原始成本 | 当前售价 | 利润率
- 批量加成：选择模型 → 设置加成比例（如 +50%）→ 预览新价格 → 确认
- 单个调价：直接编辑 input_price / output_price

#### 页面 3：客户管理

- 客户列表：名称 | API Key（脱敏） | 状态 | 上次 Pull 时间 | 创建时间
- 创建客户：输入名称 → 生成 key → 展示（仅一次可见）
- 操作：停用/启用、轮换 Key

#### 页面 4：同步状态

- 仪表盘：NewAPI 同步状态 | 客户 Pull 健康度
- 时间线：最近同步/pull 事件日志
- 告警：超过 30 分钟未 pull 的客户高亮

### 1.6 核心业务流程

#### 从 NewAPI 同步模型

```
1. 调用 NewAPI GET /api/pricing → 获取 ModelRatio + CompletionRatio
2. 调用 NewAPI GET /api/channel → 获取可用模型列表
3. 对比本地 DB：
   - 新增模型 → INSERT (status=removed，待人工发布)
   - 已有模型 → UPDATE source_ratio/source_completion_ratio
   - NewAPI 移除的 → 标记 status=deprecated
4. 写 sync_log
```

#### 客户 Pull Catalog

```
1. 验证 Bearer key → 解析 customer_id
2. 对比请求 version vs 当前 catalog version
3. version 相同 → 304
4. version 不同 → 返回 status=active 的模型列表 + 新 version
5. 更新 customer.last_pull_at
```

---

## Phase 2: Tokenjoy 私有化改造

### 2.1 models 表 schema 变更

```sql
ALTER TABLE models ADD COLUMN source TEXT NOT NULL DEFAULT 'custom';
-- source: 'platform' | 'custom'
```

单租户下，平台模型和自有模型共用 company_id，通过 source 区分。

### 2.2 新增 Sync Worker

位置：`internal/worker/catalogsync/worker.go`

```
配置：
  MODEL_CONFIG_SERVICE_URL   — Model Config Service 地址
  MODEL_CONFIG_API_KEY       — 客户专属 key（复用 NEW_API_PLATFORM_TOKEN）
  CATALOG_SYNC_INTERVAL=10m  — Pull 间隔

流程（每 10 分钟）：
1. GET {MODEL_CONFIG_SERVICE_URL}/api/v1/catalog?version={last_version}
2. 304 → skip
3. 200 → 开启事务：
   a. Upsert models 表 (WHERE source='platform')
      - catalog 中存在 → INSERT ON CONFLICT UPDATE
      - catalog 中不存在但本地有 → SET enabled=false（软删除）
   b. 调用本地 NewAPI UpsertModelRatio（写入价格）
   c. 调用本地 NewAPI RebuildAbilities
   d. 更新 last_version
4. 记录同步日志
```

### 2.3 Gateway Precheck 改造

`internal/domain/gateway/evaluate.go` 变更：

```go
// 当前逻辑：
// if wallet_quota_remain < 1 → reject

// 改为：
// if model.source == "platform" && wallet_quota_remain < 1 → reject
// if model.source == "custom" → 跳过钱包检查，只检查预算
```

**PrecheckContextRow 变更：**
- JOIN models 表获取 source 字段
- 缓存 key 扩展：包含 model source 信息

### 2.4 Ingest 改造

`internal/domain/ingest/` 变更：

```go
// 当前逻辑：
// 1. 算 cost
// 2. ConsumeLotsLocked（扣钱包）
// 3. 更新 budget_consumed（扣预算）

// 改为：
// 1. 算 cost
// 2. 查 models 表 source
// 3. if source == "platform" → ConsumeLotsLocked（扣钱包）
// 4. 更新 budget_consumed（扣预算）← 无论 source 都执行
```

### 2.5 新增环境变量

```env
# 私有化模式标识
DEPLOY_MODE=private          # saas | private

# Model Config Service 连接
MODEL_CONFIG_SERVICE_URL=https://config.your-cloud.com
MODEL_CONFIG_API_KEY=sk-customer-xxx

# 本地 NewAPI platform channel（指向云端）
NEW_API_PLATFORM_URL=https://api.your-cloud.com
NEW_API_PLATFORM_TOKEN=sk-customer-xxx
```

### 2.6 前端适配

- 模型列表页：区分展示平台模型（只读，标记「平台」tag）和自有模型（可编辑）
- 平台模型不可删除/编辑价格（由 Model Config Service 控制）
- 自有模型 CRUD 不变

---

## Phase 3: 本地 NewAPI 编排 + 部署

### 3.1 Docker Compose 编排

```yaml
services:
  tokenjoy:
    image: tokenjoy/backend:latest
    ports:
      - "8080:8080"        # 对外暴露
    env_file: .env
    depends_on:
      - postgres
      - redis
      - newapi-local

  newapi-local:
    image: tokenjoy/newapi:latest
    # 不暴露端口，仅内部网络可达
    expose:
      - "3000"
    env_file: .env.newapi

  postgres:
    image: postgres:16
    volumes:
      - pg_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine

  frontend:
    image: tokenjoy/frontend:latest
    ports:
      - "3000:80"          # 对外暴露
```

关键点：`newapi-local` 不对外暴露端口（无 `ports` 映射），只在 Docker 内部网络可访问。

### 3.2 Platform Channel 自动初始化

在 Tokenjoy Backend 启动时（或首次 catalog sync 时）：

```go
// bootstrap/private.go
func EnsurePlatformChannel(cfg config.Config, port adminport.Port) error {
    // 1. 检查 platform channel 是否已存在
    // 2. 不存在则创建：
    //    - type: 1 (openai 兼容)
    //    - name: "platform_upstream"
    //    - key: cfg.NewAPIPlatformToken
    //    - base_url: cfg.NewAPIPlatformURL
    //    - group: "platform_shared"
    //    - status: 1 (enabled)
    // 3. RebuildAbilities
}
```

### 3.3 部署清单

客户拿到的交付物：
1. `docker-compose.yml` — 编排文件
2. `.env.example` — 环境变量模板（需客户填入 key）
3. 部署文档 — 一页纸：填 key → docker compose up → 访问

客户需要你提供的信息：
- `MODEL_CONFIG_API_KEY` / `NEW_API_PLATFORM_TOKEN`（同一个 key）
- `MODEL_CONFIG_SERVICE_URL`（你的云端地址）
- `NEW_API_PLATFORM_URL`（你的云端 NewAPI 地址）

### 3.4 健康检查 + 自愈

```yaml
# docker-compose.yml
newapi-local:
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:3000/api/status"]
    interval: 30s
    retries: 3

tokenjoy:
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
    interval: 30s
```

Catalog sync worker 失败处理：
- 网络不通 → 用本地缓存继续服务，下次重试
- 连续 3 次失败 → 日志告警（不影响已有模型使用）
- Key 失效（401）→ 明确错误日志，管理员需更换 key
