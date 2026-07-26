# SMS → TokenJoy 模型/渠道/价格同步方案

## 概述

SMS 是模型配置的单一数据源（SOT），TokenJoy 通过定时轮询 SMS 的 API 获取最新的渠道、模型价格和元数据，同步到自己的 NewAPI 实例和本地 models 表，最终在前端模型列表中展示。

## 架构

```
┌─────────────┐         HTTP API         ┌──────────────┐
│     SMS     │ ◄─────── 定时轮询 ────────│   TokenJoy   │
│  (配置源)   │   GET /api/sync/catalog   │   (消费方)   │
└──────┬──────┘                           └──────┬───────┘
       │                                         │
       │ 管理员配置                                │ 同步写入
       ▼                                         ▼
┌─────────────┐                           ┌──────────────┐
│ SMS-NewAPI  │                           │TokenJoy-NewAPI│
│  (port 3020)│                           │  (port 3000) │
└─────────────┘                           └──────────────┘
```

## 数据流

```
1. 管理员在 SMS 配置渠道/模型/价格
2. TokenJoy 后台 worker 定时（如每 5 分钟）调用 SMS API
3. 获取完整模型目录（渠道 + 模型 + 价格）
4. 对比本地数据，diff 后执行：
   a. 渠道 → 同步到 TokenJoy-NewAPI（创建/更新 channel）
   b. 价格 → 同步到 TokenJoy-NewAPI（upsert ModelRatio/CompletionRatio）
   c. 元数据 → 同步到 TokenJoy models 表（名称、提供商、类型、价格）
5. 前端模型列表自动展示最新数据
```

## 详细设计

### 1. SMS 侧：新增同步 API

**接口：** `GET /api/sync/catalog`

**认证：** OAuth2 Bearer Token（TokenJoy 先通过 client_credentials 换取短命 token）

**响应体：**
```json
{
  "channels": [
    {
      "name": "deepseek-official",
      "type": 1,
      "baseUrl": "https://api.deepseek.com",
      "key": "sk-xxx",
      "models": ["deepseek-chat", "deepseek-coder"],
      "priority": 0
    }
  ],
  "models": [
    {
      "modelId": "deepseek-chat",
      "displayName": "DeepSeek Chat",
      "provider": "deepseek",
      "callType": "chat",
      "inputPrice": 1.0,
      "outputPrice": 2.0
    }
  ]
}
```

**字段说明：**
- `channels[].key` — 供应商 API Key（SMS 管理，TokenJoy 直接使用）
- `models[].inputPrice / outputPrice` — 元/百万 token
- `models[].callType` — chat / embedding / image / audio

**文件位置：** `sms/backend/internal/http/handler/sync/handler.go`

---

### 2. TokenJoy 侧：定时轮询 Worker

**触发机制：** River 定时任务（每 5 分钟）或 cron

**实现步骤：**

```go
// 1. 调用 SMS API 获取目录
catalog := smsClient.FetchCatalog(ctx)

// 2. 同步渠道到 TokenJoy-NewAPI
for _, ch := range catalog.Channels {
    newAPIClient.UpsertChannel(ctx, ch)
}
newAPIClient.RebuildAbilities(ctx)

// 3. 同步价格到 TokenJoy-NewAPI
for _, m := range catalog.Models {
    newAPIClient.UpsertModelRatio(ctx, m.ModelId, m.InputPrice, m.OutputPrice)
}

// 4. 同步模型元数据到本地 models 表
for _, m := range catalog.Models {
    modelsStore.Upsert(ctx, Model{
        Name: m.ModelId, DisplayName: m.DisplayName,
        Provider: m.Provider, CallType: m.CallType,
        InputPrice: m.InputPrice, OutputPrice: m.OutputPrice,
        Source: "sms",  // 标记来源
    })
}
```

**文件位置：**
- Worker: `apps/backend/internal/worker/sms_sync_worker.go`
- SMS Client: `apps/backend/internal/integration/sms/client.go`
- 配置: `apps/backend/internal/config/sms.go`

**环境变量：**
```
SMS_API_BASE_URL=http://127.0.0.1:8020
SMS_SYNC_API_KEY=sms-sync-key-xxxxx
SMS_SYNC_INTERVAL_SEC=600
```

---

### 2.5 授信方案

**方式：OAuth2 Client Credentials（公网安全）**

两个系统通过公网通信，采用标准 OAuth2 client_credentials 流程：

```
TokenJoy                                    SMS
   │                                          │
   │ POST /api/oauth/token                    │
   │ { client_id, client_secret,              │
   │   grant_type: "client_credentials" }     │
   │ ────────────────────────────────────────► │
   │                                          │
   │ { access_token, expires_in: 600 }        │
   │ ◄──────────────────────────────────────── │
   │                                          │
   │ GET /api/sync/catalog                    │
   │ Authorization: Bearer <access_token>     │
   │ ────────────────────────────────────────► │
   │                                          │
   │ { channels, models }                     │
   │ ◄──────────────────────────────────────── │
```

**设计要点：**

| 维度 | 决策 |
|------|------|
| Token 有效期 | 10 分钟（每次轮询前换取新 token） |
| Client 管理 | SMS 维护 oauth_clients 表（client_id + hashed_secret + scope） |
| Scope | `sync:read`（只允许读同步接口） |
| Secret 吸销 | SMS 单端操作即可，TokenJoy 下次换 token 失败，不影响已有服务 |
| 传输安全 | 生产环境 HTTPS 强制 |
| Token 缓存 | TokenJoy 本地缓存 token，过期前 30s 主动刷新 |
| 审计 | SMS 记录 token 签发日志（client_id + IP + timestamp） |

**SMS 侧环境变量：**
```
OAUTH_CLIENT_ID=tokenjoy-sync
OAUTH_CLIENT_SECRET=xxx-generated-secret
```

**TokenJoy 侧环境变量：**
```
SMS_API_BASE_URL=https://sms.example.com
SMS_CLIENT_ID=tokenjoy-sync
SMS_CLIENT_SECRET=xxx-generated-secret
SMS_SYNC_INTERVAL_SEC=600
```

**SMS 侧需要新增：**
- `POST /api/oauth/token` — 签发短命 JWT（验证 client_id + secret，返回 access_token）
- `oauth_clients` 表 — 存储注册的客户端（id, hashed_secret, scope, created_at）
- 中间件：`/api/sync/*` 路由验证 Bearer token（JWT 签名校验 + scope 检查）

---

### 3. 同步策略

| 维度 | 策略 |
|------|------|
| 渠道 | 按 name 匹配，存在则更新（key/models/baseUrl），不存在则创建 |
| 价格 | read-modify-write merge，只覆盖 SMS 管理的模型，不删除其他 |
| 模型元数据 | 按 modelId upsert 到 models 表，标记 source="sms" |
| 删除 | SMS 删除模型后，下次同步时 TokenJoy 不主动删除，标记 deprecated |
| 冲突 | SMS 为准，本地手动修改会被下次同步覆盖（source="sms" 的模型） |

---

### 4. 前端展示

TokenJoy 模型列表已有完整的展示能力：
- 模型名称 + 类型
- 提供商 badge
- 输入/输出价格（元/百万 token）
- 来源标记（可选：显示"SMS 同步"标签）

**无需额外前端改动**，同步后数据自动出现在模型列表。

---

## 任务拆分

### Phase 1：SMS 提供同步 API

| 任务 | 文件 | 说明 |
|------|------|------|
| 1.1 | `sms/backend/internal/http/handler/sync/handler.go` | 新增 GET /api/sync/catalog |
| 1.2 | `sms/backend/internal/domain/model/service.go` | 新增 ListCatalog 方法（聚合 channel + model + price） |
| 1.3 | `sms/backend/internal/http/middleware/apikey.go` | API Key 认证中间件 |

### Phase 2：TokenJoy 定时拉取

| 任务 | 文件 | 说明 |
|------|------|------|
| 2.1 | `apps/backend/internal/integration/sms/client.go` | SMS HTTP 客户端 |
| 2.2 | `apps/backend/internal/integration/sms/types.go` | 同步数据结构 |
| 2.3 | `apps/backend/internal/worker/sms_sync_worker.go` | 定时同步 worker |
| 2.4 | `apps/backend/internal/config/sms.go` | SMS 配置（URL + Key + 间隔） |

### Phase 3：写入 TokenJoy

| 任务 | 文件 | 说明 |
|------|------|------|
| 3.1 | 复用 `internal/integration/newapi/channel.go` | 渠道同步到 NewAPI |
| 3.2 | 复用 `internal/integration/newapi/option.go` | 价格同步到 NewAPI |
| 3.3 | `internal/domain/models/service.go` | UpsertFromSMS 方法写入 models 表 |

### Phase 4：可选增强

| 任务 | 说明 |
|------|------|
| 4.1 | 前端模型列表增加"来源"列（SMS 同步 / 手动创建） |
| 4.2 | 同步日志记录（成功/失败/diff 数量） |
| 4.3 | SMS 前端增加"已同步到 TokenJoy"状态标记 |

---

## 注意事项

1. **SMS 渠道的 API Key 安全**：通过 HTTPS + API Key 认证传输，TokenJoy 不存储 Key 明文（直接传给 NewAPI）
2. **幂等性**：每次同步是全量对比，重复执行不产生副作用
3. **故障隔离**：同步失败不影响 TokenJoy 现有功能，只是数据不更新
4. **来源标记**：models 表的 `source` 字段区分 SMS 同步 vs 手动创建，避免同步覆盖手动配置
