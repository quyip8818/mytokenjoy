# SMS → TokenJoy 模型/渠道/价格同步

> 状态：待实现
> SMS 是模型配置的单一数据源（SOT），TokenJoy 定时轮询获取渠道、模型价格和元数据。

---

## 1. 架构

```
┌─────────────┐    OAuth2 + HTTP API    ┌──────────────┐
│     SMS     │ ◄────── 定时轮询 ───────│   TokenJoy   │
│  (配置源)   │  GET /api/sync/catalog  │   (消费方)   │
└──────┬──────┘                         └──────┬───────┘
       │                                       │
       ▼                                       ▼
┌─────────────┐                         ┌──────────────┐
│ SMS-NewAPI  │                         │TokenJoy-NewAPI│
│  (port 3020)│                         │  (port 3010) │
└─────────────┘                         └──────────────┘
```

---

## 2. 需求规格（EARS）

### 通用

- HTTPS 跨公网通信
- OAuth2 Client Credentials 跨系统认证
- 「元/百万 token」统一价格单位
- 定价公式：`modelRatio = inputPrice / 2`，`completionRatio = outputPrice / inputPrice`
- 价格同步 read-modify-write merge（只覆盖 SMS 管理的 key）

### 事件驱动

- 定时任务触发（每 10 分钟）→ 调 SMS API 拉取目录
- 新渠道 → 写入 TokenJoy-NewAPI channel
- 新定价 → 写入 ModelRatio/CompletionRatio
- 新元数据 → upsert TokenJoy models 表（source=`sms`）
- 渠道写入后 → RebuildAbilities 刷新路由
- SMS 删除模型 → TokenJoy 标记 deprecated（不主动删）

### 条件

- `source="manual"` 的同名模型不被覆盖
- SMS 不可达 → 记录错误、下次重试、不影响现有功能
- `NEW_API_ENABLED=false` → 跳过 NewAPI 操作
- `SMS_API_BASE_URL` 未配置 → 不启动同步 worker

### 不应

- 不暴露供应商 API Key 明文（HTTPS 传输）
- 不删除非 SMS 管理的渠道/定价 key
- 同步失败不导致服务降级
- 不并发轮询

---

## 3. 数据模型

### SMS 同步 API 响应（GET /api/sync/catalog）

```json
{
  "channels": [{
    "name": "deepseek-official",
    "type": 43,
    "baseUrl": "https://api.deepseek.com",
    "key": "sk-xxx",
    "models": ["deepseek-chat", "deepseek-coder"],
    "group": "default",
    "priority": 0
  }],
  "models": [{
    "modelId": "deepseek-chat",
    "displayName": "DeepSeek Chat V3",
    "provider": "deepseek",
    "callType": "chat",
    "inputPrice": 1.0,
    "outputPrice": 2.0
  }],
  "syncedAt": "2026-07-19T10:00:00Z"
}
```

### OAuth2 Token

```
POST /api/oauth/token
{ "grant_type": "client_credentials", "client_id": "tokenjoy-sync", "client_secret": "xxx" }
→ { "access_token": "...", "token_type": "Bearer", "expires_in": 600, "scope": "sync:read" }
```

### TokenJoy models 表扩展

- `source TEXT` — `sms` / `manual` / `seed`
- `sms_synced_at TIMESTAMPTZ`

### SMS oauth_clients 表

`id, client_id, client_secret_hash, scope, created_at`

---

## 4. TokenJoy 侧同步 Worker

```go
// 每 10 分钟
func (w *Worker) syncOnce(ctx) {
    token := smsClient.GetOAuth2Token(ctx)
    catalog := smsClient.FetchCatalog(ctx, token)

    for _, ch := range catalog.Channels {
        adminport.UpsertChannel(ctx, ch)
    }
    adminport.RebuildAbilities(ctx)

    for _, m := range catalog.Models {
        adminport.UpsertModelRatio(ctx, m.CallType, m.InputPrice, m.OutputPrice)
    }

    for _, m := range catalog.Models {
        modelsStore.Upsert(ctx, Model{..., Source: "sms"})
    }
}
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `SMS_API_BASE_URL` | SMS 服务地址 |
| `SMS_CLIENT_ID` | OAuth2 client_id |
| `SMS_CLIENT_SECRET` | OAuth2 client_secret |
| `SMS_SYNC_INTERVAL_SEC` | 轮询间隔（默认 600） |

---

## 5. SMS 侧新增

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/oauth/token` | 签发短命 JWT |
| GET | `/api/sync/catalog` | Bearer token 认证，返回完整目录 |

新增 `oauth_clients` 表 + 中间件（JWT 校验 + scope 检查）。

---

## 6. 实施路线

### Phase 1：基础同步链路
1. SMS OAuth2 token 端点 + oauth_clients 表
2. SMS `GET /api/sync/catalog`
3. TokenJoy SMS HTTP client（含 token 管理）
4. TokenJoy 定时 worker
5. 写入：渠道→NewAPI，定价→NewAPI，元数据→models 表

### Phase 2：可观测性
6. 同步日志持久化
7. SMS 前端"已同步"状态
8. TokenJoy 模型列表"来源"列

### Phase 3：运维
9. 手动全量同步按钮
10. 同步健康检查 API

---

## 7. E2E 测试方案

### Layer 1: API 层（16 用例）

| 类别 | 用例 |
|------|------|
| OAuth2 | 有效 credentials→200+JWT；错误 secret→401；不存在 client→401；错误 grant_type→400 |
| Catalog | 有效 token→200 含 channels+models；无 header→401；过期 token→401；数据结构校验 |
| 同步后验证 | 触发同步→models API 返回 SMS 模型；NewAPI pricing 含同步定价；SMS 不可达→现有模型不变 |

### Layer 2: 浏览器 E2E（4 用例）

- 登录→导航模型列表→页面正常
- 同步模型正确展示（名称/价格）
- 来源标记正确
- 手动创建模型不被同步覆盖

### 测试前置

```
SMS port 8020 | TokenJoy port 8010 | NewAPI port 3010 | Frontend port 5173
SMS DB 有 oauth_clients + 模型数据 | TokenJoy .env SMS_SYNC_ENABLED=true
```

### 文件

```
apps/frontend/e2e/sms-sync-api.spec.ts      # Layer 1
apps/frontend/e2e/sms-sync-models.spec.ts   # Layer 2
```

---

## 8. 验收标准

| AC | 描述 |
|----|------|
| 1 | SMS 配新模型 → 10 分钟内出现在 TokenJoy |
| 2 | SMS 改价格 → TokenJoy 同步更新 |
| 3 | SMS 新增渠道 → TokenJoy-NewAPI 出现对应 channel |
| 4 | TokenJoy 手动创建的模型不被覆盖 |
| 5 | SMS 不可达 → TokenJoy 正常运行 |
| 6 | OAuth2 secret 吊销 → 同步停止，不影响其他功能 |
| 7 | 调用者能用 TokenJoy Key 调用 SMS 同步的模型 |
