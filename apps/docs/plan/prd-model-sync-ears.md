# TokenJoy × NewAPI × SMS 模型管理集成 PRD

> 基于 EARS (Easy Approach to Requirements Specification) 方法论编写

## 1. 概述

### 1.1 目标

建立从供应商管理（SMS）→ LLM 网关（NewAPI）→ 客户平台（TokenJoy）的完整模型管理链路，实现模型配置的单一数据源管理和自动化同步。

### 1.2 系统边界

| 系统 | 角色 | 职责 |
|------|------|------|
| SMS | 配置源 | 供应商管理、渠道配置、模型定价的单一数据源 |
| NewAPI | 网关层 | LLM 请求路由、Token 管理、消耗计量 |
| TokenJoy | 消费层 | 面向客户的 API 管控、预算管理、用量展示 |

### 1.3 用户角色

| 角色 | 系统 | 职责 |
|------|------|------|
| 供应商管理员 | SMS | 配置渠道、模型、定价 |
| 平台管理员 | TokenJoy | 查看模型列表、管理白名单、分配预算 |
| API 调用者 | TokenJoy | 使用 API Key 调用 LLM 模型 |

---

## 2. 需求规格（EARS 格式）

### 2.1 通用需求（Ubiquitous）

**REQ-U01** 系统应使用 HTTPS 进行所有跨公网的服务间通信。

**REQ-U02** 系统应使用 OAuth2 Client Credentials 作为跨系统认证机制。

**REQ-U03** 系统应使用「元/百万 token」作为统一的价格单位。

**REQ-U04** 系统应使用以下定价转换公式：
```
modelRatio      = inputPrice / 2
completionRatio = outputPrice / inputPrice
inputPrice      = modelRatio × 2
outputPrice     = modelRatio × completionRatio × 2
```

**REQ-U05** 模型价格同步应采用 read-modify-write merge 语义——只覆盖 SMS 管理的模型 key，不删除其他系统手动配置的模型。

---

### 2.2 事件驱动需求（Event-Driven）

**REQ-E01** 当 TokenJoy 后台定时任务触发时（每 10 分钟），系统应调用 SMS 同步 API 拉取最新模型目录。

**REQ-E02** 当同步拉取到新的渠道数据时，系统应将渠道配置写入 TokenJoy-NewAPI（创建或更新）。

**REQ-E03** 当同步拉取到新的模型定价时，系统应将 ModelRatio 和 CompletionRatio 写入 TokenJoy-NewAPI 系统选项。

**REQ-E04** 当同步拉取到新的模型元数据时，系统应将模型信息（名称、提供商、类型、价格）upsert 到 TokenJoy 的 models 表。

**REQ-E05** 当渠道写入 TokenJoy-NewAPI 完成后，系统应调用 RebuildAbilities 刷新模型路由表。

**REQ-E06** 当 NewAPI 记录一条消耗日志时，系统应通过 Webhook POST 到 TokenJoy 后端触发 ingest pipeline。

**REQ-E07** 当 TokenJoy 调用 NewAPI 返回 401 时，系统应自动从 NewAPI 数据库重新读取 admin token 并重试请求。

**REQ-E08** 当 SMS 删除一个已同步的模型时，下次同步应将该模型在 TokenJoy 中标记为 deprecated（不主动删除）。

---

### 2.3 条件需求（If-Then / State-Driven）

**REQ-S01** 如果 TokenJoy models 表中已存在 `source="manual"` 的同名模型，同步不应覆盖该模型（SMS 同步的 source 标记为 `"sms"`）。

**REQ-S02** 如果 SMS 同步 API 不可达，系统应记录错误日志并在下次轮询时重试，不应影响 TokenJoy 现有功能。

**REQ-S03** 如果 OAuth2 token 获取失败（SMS 不可达或 secret 错误），系统应记录错误并跳过本轮同步。

**REQ-S04** 如果 `NEW_API_ENABLED=false`，系统不应执行任何 NewAPI 相关操作（渠道同步、定价写入、webhook 接收均跳过）。

**REQ-S05** 如果 `SMS_API_BASE_URL` 未配置，系统不应启动 SMS 同步 worker。

**REQ-S06** 如果渠道的 API Key 为空，同步应跳过该渠道并记录警告。

---

### 2.4 不需要的行为（Unwanted / Shall-Not）

**REQ-N01** 系统不应在公网传输中暴露供应商 API Key 的明文（应通过 HTTPS 加密传输）。

**REQ-N02** 同步过程不应删除 TokenJoy-NewAPI 中非 SMS 管理的渠道。

**REQ-N03** 同步过程不应删除 TokenJoy-NewAPI 定价选项中非 SMS 管理的模型 key。

**REQ-N04** 同步失败不应导致 TokenJoy 服务降级或不可用。

**REQ-N05** 同步 worker 不应在单次轮询未完成时启动下一次轮询（防止并发冲突）。

---

### 2.5 可选/复杂需求（Optional）

**REQ-O01** [Phase 2] 系统应支持 SMS 前端展示每个模型的"已同步到 TokenJoy"状态。

**REQ-O02** [Phase 2] 系统应记录同步日志（成功/失败/新增/更新/跳过的条目数）。

**REQ-O03** [Phase 2] 平台管理员应能在 TokenJoy 前端区分 SMS 同步的模型和手动创建的模型。

---

## 3. 数据模型

### 3.1 SMS 同步 API 响应（GET /api/sync/catalog）

```json
{
  "channels": [
    {
      "name": "deepseek-official",
      "type": 43,
      "baseUrl": "https://api.deepseek.com",
      "key": "sk-xxx",
      "models": ["deepseek-chat", "deepseek-coder"],
      "group": "default",
      "priority": 0,
      "settings": {}
    }
  ],
  "models": [
    {
      "modelId": "deepseek-chat",
      "displayName": "DeepSeek Chat V3",
      "provider": "deepseek",
      "callType": "chat",
      "inputPrice": 1.0,
      "outputPrice": 2.0
    }
  ],
  "syncedAt": "2026-07-19T10:00:00Z"
}
```

### 3.2 OAuth2 Token 接口

**请求：** `POST /api/oauth/token`
```json
{
  "grant_type": "client_credentials",
  "client_id": "tokenjoy-sync",
  "client_secret": "xxx"
}
```

**响应：**
```json
{
  "access_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 600,
  "scope": "sync:read"
}
```

### 3.3 TokenJoy models 表扩展

| 字段 | 类型 | 说明 |
|------|------|------|
| source | TEXT | `"sms"` / `"manual"` / `"seed"` — 区分来源 |
| sms_synced_at | TIMESTAMPTZ | 最后一次从 SMS 同步的时间 |

### 3.4 SMS oauth_clients 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| client_id | TEXT | 唯一标识（如 `tokenjoy-sync`） |
| client_secret_hash | TEXT | bcrypt hash |
| scope | TEXT | 权限范围（如 `sync:read`） |
| created_at | TIMESTAMPTZ | 创建时间 |

---

## 4. 非功能需求

### 4.1 性能

| 指标 | 要求 |
|------|------|
| 同步接口响应时间 | < 3s（含所有模型+渠道数据） |
| 单次同步总耗时 | < 30s（含 NewAPI 写入） |
| 轮询间隔 | 10 分钟（可配置） |

### 4.2 可靠性

| 场景 | 行为 |
|------|------|
| SMS 不可达 | 跳过本轮，下轮重试，不影响 TokenJoy |
| NewAPI 不可达 | 记录错误，模型元数据仍写入本地 |
| 部分渠道写入失败 | 继续处理其余渠道，记录失败项 |
| 同步中途崩溃 | 幂等设计，下次全量重跑无副作用 |

### 4.3 安全

| 维度 | 措施 |
|------|------|
| 传输 | HTTPS（生产）/ HTTP（本地开发） |
| 认证 | OAuth2 Client Credentials，token 10 分钟过期 |
| 授权 | scope 限制只读同步接口 |
| 密钥 | client_secret bcrypt 存储，不可逆 |
| 审计 | token 签发记录（client_id + IP + time） |

---

## 5. 实现路线

### Phase 1：基础同步链路

1. SMS 新增 OAuth2 token 端点 + oauth_clients 表
2. SMS 新增 `GET /api/sync/catalog` 接口
3. TokenJoy 新增 SMS HTTP client（含 OAuth2 token 管理）
4. TokenJoy 新增定时 worker（10 分钟轮询）
5. 同步写入：渠道 → NewAPI，定价 → NewAPI，元数据 → models 表

### Phase 2：可观测性增强

6. 同步日志记录（同步结果持久化）
7. SMS 前端"已同步"状态标记
8. TokenJoy 模型列表"来源"列

### Phase 3：运维工具

9. 手动全量同步按钮（SMS 前端或 TokenJoy 后台）
10. 同步健康检查 API（最后成功时间、失败次数）

---

## 6. 验收标准

| AC | 描述 |
|----|------|
| AC-01 | SMS 配置新模型后，10 分钟内出现在 TokenJoy 模型列表 |
| AC-02 | SMS 修改模型价格后，TokenJoy 模型列表价格同步更新 |
| AC-03 | SMS 新增渠道后，TokenJoy-NewAPI 出现对应渠道，模型可路由 |
| AC-04 | TokenJoy 手动创建的模型不被 SMS 同步覆盖 |
| AC-05 | SMS 不可达时，TokenJoy 正常运行，模型列表显示上次同步数据 |
| AC-06 | OAuth2 secret 吸销后，同步停止，不影响 TokenJoy 其他功能 |
| AC-07 | API 调用者能通过 TokenJoy 签发的 Key 调用 SMS 同步过来的模型 |
