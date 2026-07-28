# 私有化部署 + 定价管理

> 状态：待实现
> 私有化客户独立运行 TokenJoy + 本地 NewAPI，通过云端 Model Config Service 获取模型目录与定价。

---

## 1. 架构总览

```
┌──────────────── 云端（你管理）─────────────────┐
│  Model Config Service (Go+React+PG)            │
│  ├─ 从 NewAPI 导入原始模型                      │
│  ├─ 人工编辑（发布/定价加成/分组）              │
│  └─ GET /api/v1/catalog 下发给客户              │
│                                                 │
│  云端 NewAPI（平台网关）                         │
│  └─ 持有上游 API Key，按 token 计费             │
└────────────────────┬────────────────────────────┘
                     │ catalog pull (每 10min) + LLM 代理
┌────────────────────┼────────────────────────────┐
│        客户私有化环境（单租户）                    │
│  ┌──────────────┐    ┌───────────────────┐      │
│  │ TokenJoy     │    │ NewAPI（本地,隐藏） │      │
│  │ ├─ Gateway ──┼───►│ ├─ platform ch ───┼─► 云端 NewAPI → Provider
│  │ ├─ sync worker│   │ └─ custom ch ─────┼─► Provider 直连
│  │ └─ ingest    │◄──│   (webhook)        │      │
│  └──────────────┘    └───────────────────┘      │
└─────────────────────────────────────────────────┘
```

---

## 2. 三端定价管理

| 版本 | 界面 | 模型管理能力 |
|------|------|-------------|
| **官方管理平台** | NewAPI 原生 admin UI（零开发） | 完整定价编辑、Channel 管理、上游接入 |
| **Local** | TJ 自研 UI | 内置模型只读 + 自定义模型可编辑 + 客户自有 Channel |
| **SaaS** | TJ 自研 UI（子集） | 内置模型只读，无自定义模型/Channel |

### Local 功能矩阵

| 功能 | 内置模型 | 自定义模型 |
|------|---------|-----------|
| 查看列表+价格 | ✅（只读） | ✅ |
| 编辑价格 | ❌ | ✅ |
| 添加/删除 | ❌ | ✅ |
| endpoint/API Key | ❌ | ✅ |
| 启停 | ✅ | ✅ |
| 路由/白名单 | ✅ | ✅ |

### 计费防 hack

计费权在平台网关侧。Local 改本地 NewAPI ratio 不影响平台账单——请求经平台网关，按平台 ratio 计费。

---

## 3. Model Config Service

### 数据模型

**models 表：** model_id, type, name, provider, input_price, output_price, max_context, capabilities[], group_name, sort_order, status(active/deprecated/removed), source_ratio, source_completion_ratio

**customers 表：** customer_id, name, api_key, status, last_pull_at

**sync_logs 表：** log_id, direction, status, model_count, error_message

### 管理端 API

```
POST   /api/admin/sync              触发从 NewAPI 同步
GET    /api/admin/models             模型列表
PUT    /api/admin/models/:id         编辑
PUT    /api/admin/models/:id/status  发布/下架
GET    /api/admin/customers          客户列表
POST   /api/admin/customers          创建+生成 key
POST   /api/admin/customers/:id/rotate-key  轮换
PUT    /api/admin/pricing/batch      批量调价
```

### 客户端 API

```
GET /api/v1/catalog?version=<last>
Authorization: Bearer <customer_api_key>
→ 200 { version, models[] } 或 304（未变化）
```

### 前端页面

1. 模型目录（表格+拖拽排序+状态筛选+批量发布）
2. 定价管理（成本对照+批量加成）
3. 客户管理（列表+创建+轮换 Key）
4. 同步状态（仪表盘+日志）

---

## 4. TokenJoy 私有化改造

### models 表

`source TEXT NOT NULL DEFAULT 'custom'` — `platform` | `custom`

### Catalog Sync Worker

```
每 10 分钟：
1. GET {MODEL_CONFIG_SERVICE_URL}/api/v1/catalog?version={last}
2. 304 → skip
3. 200 → 事务内：
   a. Upsert models (WHERE source='platform')
   b. UpsertModelRatio → 本地 NewAPI
   c. RebuildAbilities
   d. 不在 catalog 中的 → enabled=false
```

### Gateway Precheck 改造

- `model.source == "platform"` → 正常钱包检查
- `model.source == "custom"` → 跳过钱包检查，只检查预算

### Ingest 改造

- `source == "platform"` → 正常 ConsumeLotsLocked（扣钱包+预算）
- `source == "custom"` → 跳过 lot 消耗，只扣预算

### 环境变量

```env
DEPLOY_MODE=private
MODEL_CONFIG_SERVICE_URL=https://config.your-cloud.com
MODEL_CONFIG_API_KEY=sk-customer-xxx
NEW_API_PLATFORM_URL=https://api.your-cloud.com
NEW_API_PLATFORM_TOKEN=sk-customer-xxx
```

---

## 5. Docker 部署

```yaml
services:
  tokenjoy:
    image: tokenjoy/backend:latest
    ports: ['8080:8080']
    depends_on: [postgres, redis, newapi-local]
  newapi-local:
    image: tokenjoy/newapi:latest
    expose: ['3000']  # 不对外暴露
  postgres:
    image: postgres:16
  redis:
    image: redis:7-alpine
  frontend:
    image: tokenjoy/frontend:latest
    ports: ['3000:80']
```

Platform Channel 启动时自动初始化（指向云端 NewAPI）。

### 客户交付物

1. `docker-compose.yml`
2. `.env.example`（需填 key）
3. 一页纸部署文档

---

## 6. 实施路线

| Phase | 内容 |
|-------|------|
| 1 | Model Config Service（模型管理+客户管理+catalog API） |
| 2 | TokenJoy catalog sync worker + models source 字段 |
| 3 | Gateway/Ingest 按 source 分流 |
| 4 | Docker 编排 + platform channel 自动初始化 |
| 5 | 前端适配（内置只读+自定义可编辑） |

---

## 7. 决策记录

| 决策 | 理由 |
|------|------|
| 平台用 NewAPI 原生 UI | 功能匹配，零开发 |
| 客户侧用 TJ 自研 UI | 权限/体验/安全需自控 |
| 计费在平台网关侧 | 架构保证，无需防 hack |
| Local 同步强制 | 保证展示一致（不影响计费） |
| 自定义模型 Local 可编辑 | 客户自己的模型，平台不管 |
