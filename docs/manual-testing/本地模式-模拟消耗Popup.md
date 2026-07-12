# 本地模式 — 模拟消耗 Popup（设计 & 测试说明）

> **读者**：产品、QA、研发 — 需要在本地开发环境 fake 一笔 token 消耗，并在 UI 上验证 Ingest 计费链路。  
> **性质**：**设计说明 + 实现清单 + 验收步骤**（当前仓库尚未实现；本文档供后续开发使用）。  
> **技术选型**：**方案 A** — logs 写入 + ingest 队列入队 + Worker 入账 + 投影（见 §4.4 与生产的等价边界）。  
> **相关**：[Backend-Ingest架构.md](../Backend-Ingest架构.md) · [Backend-计费模式.md](../Backend-计费模式.md)

---

## 1. 要解决什么问题

本地联调用量计费时，常见痛点：

| 现状 | 问题 |
| --- | --- |
| Demo 种子（`BOOTSTRAP_MODE=demo`） | 只能看初始数据，**不能增量**产生新消耗 |
| SQL + Webhook 注入 | 步骤多（psql、查 `newapi_key_id`、curl），且无法方便地换模型 / token 数 |
| Gateway 真实调用 | 依赖 NewAPI 全栈，慢且可能调真模型 |

**目标体验**：Header 点「模拟消耗」→ Popup 选模型、填 input/output tokens → 提交后进入 **queued** 状态 → 轮询至 **ingested** → 在 `/audit/calls`、`/wallet`、`/budget` 验收（整体约 **5–15 秒**）。

---

## 2. 功能范围

### 2.1 做什么

- 本地 Dev 工具栏 **「模拟消耗」** → **Popup（Dialog）**
- Popup 配置：**Platform Key**、**模型（modelId）**、**Input / Output tokens**
- 提交后走方案 A 核心入账路径：
  ```text
  INSERT logs.newapi.logs → ingest_jobs 入队 → IngestWorker → usage_ledger → 投影 job
  ```
- API 返回 **queued**（非「已成功计费」）；UI 轮询状态后再 Toast / 刷新 / 跳转验收

### 2.2 不做什么

- **不在 staging / production 暴露**（404）
- **不替代** `pnpm verify:integration`（后者覆盖真实 webhook HTTP 通路）
- **不模拟 Gateway 预检**（403 挡单需另测 Gateway）
- **不要求真实 LLM 调用**

---

## 3. 显示条件（双重门禁）

| 层级 | 条件 | 说明 |
| --- | --- | --- |
| **前端** | `import.meta.env.DEV === true` | Vite dev；生产 build 不含 Popup |
| **后端** | `DEPLOY_ENV=local` | staging/production 不注册路由 |

建议：`GET /api/dev/capabilities` 且 `simulateConsume && ingestEnabled` 时才渲染入口。

```json
{
  "deployEnv": "local",
  "simulateConsume": true,
  "ingestEnabled": true
}
```

---

## 4. 技术方案：方案 A

### 4.1 链路

```text
Popup 提交
    │
    ▼
POST /api/dev/simulate-consume          （session 鉴权，local gate）
    │
    ▼
① modelId → catalog 解析 model.Type → 写入 logs.model_name
   （计费只信 catalog，不接受前端传 modelName）
    │
    ▼
② 校验 Platform Key：active + mapping.synced + newapi_key_id 非空
    │
    ▼
③ 按 input/output tokens + 模型单价计算 point → 反推 quota
    │
    ▼
④ INSERT logs.newapi.logs（type=2, token_id, model_name, quota, prompt/completion tokens）
    │
    ▼
⑤ ingestQueue.Enqueue(log_id, "webhook")    // 与 Worker 消费路径一致
    │
    ▼
⑥ 202 { logId, status: "queued", pollUrl, idempotencyKey, estimatedPoint, ... }
    │
    ▼ （异步，约 5–10s）
IngestWorker → IngestByLogID → usage_ledger
    → EnqueueAfterIngest → budget_projection / dashboard_project / wallet_sync
    │
    ▼
GET /api/dev/simulate-consume/{logId}  →  status: ingested | failed
```

**语义约定**：`queued` = 假 log 已写入且 job 已入队；**不等于**钱包已扣、ledger 已写。Worker 失败时 poll 返回 `failed` + `lastError`。

### 4.2 与生产的等价边界

| 覆盖 | 不覆盖 |
| --- | --- |
| 共享 logs 库 consume 行格式 | `POST /api/internal/webhooks/newapi-log` HTTP handler |
| `ingest_jobs` 入队 + Worker `IngestByLogID` | Webhook secret 鉴权、`RecordNotifySuccess` metrics |
| `usage_ledger` 幂等键 `newapi:{log_id}` | NewAPI 真实结算 / Gateway 预检 |
| 投影 job（budget / dashboard / wallet_sync） | Reconcile 补洞逻辑（可另测） |

若要验收 **完整 webhook HTTP 通路**，仍用 `pnpm verify:integration` 或手动 curl webhook。

**实现建议**：dev handler 内部复用 `ingestQueue.Enqueue()`（与 webhook handler 入队相同），不要绕开 queue 直接调 `IngestByLogID`，以保持与生产 Worker 路径一致。

### 4.3 前置依赖

| 依赖 | 说明 |
| --- | --- |
| `LOG_DATABASE_URL` | `IngestEnabled()` 为 true |
| `NEW_API_WEBHOOK_SECRET` | 本地 ingest 配置校验 |
| Platform Key + **synced** `newapi_key_id` | 无 mapping 的 Key **不可选** |
| Backend Worker 运行中 | 消费 `ingest_jobs` |

> Demo 种子 Key `plk-1` 若未 sync，**不会**出现在 options 下拉中；默认选中 **第一个 synced active Key**，无则空态引导去 `/keys/platform` 建 Key。

### 4.4 quota 如何从 tokens 算出

Ingest 计费读 log 行的 **`quota`**（`CostFromLog`）；`prompt_tokens` / `completion_tokens` 进入 **审计展示**。

```text
point = (inputTokens / 1_000_000) × model.inputPrice
      + (outputTokens / 1_000_000) × model.outputPrice

modelPricePoint = model.inputPrice + model.outputPrice    // ingest 侧 ModelPricePoint

quota = point × QuotaPerUnit / modelPricePoint            // QuotaPerUnit = 500_000
```

入账时 `CostFromLog(quota, model.Type, ...)` 应 **round-trip 回到上述 point**（前后端共用同一公式，建议抽 shared 说明或后端为准、前端只展示）。

**展示币**：`estimatedDisplay = point / DefaultPointsPerUnit`（1000 point = ¥1）。

#### 示例 A — 小额（理解公式）

模型 `gpt-4o-mini`（inputPrice=150、outputPrice=600 point/M tokens），**input=1200，output=800**：

| 字段 | 值 |
| --- | --- |
| point | 0.18 + 0.48 = **0.66** |
| estimatedDisplay | **¥0.00066** |
| quota | 0.66 × 500000 / 750 ≈ **440** |

#### 示例 B — Popup 默认值（UI 可见变化）

同模型，**input=12_000_000，output=8_000_000**：

| 字段 | 值 |
| --- | --- |
| point | 1800 + 4800 = **6600** |
| estimatedDisplay | **¥6.60** |
| quota | 6600 × 500000 / 750 = **4_400_000** |

> Popup 默认 token 应用 **示例 B** 量级；示例 A 仅说明公式，不适合作为默认验收数据。

### 4.5 模型契约（modelId）

| 规则 | 说明 |
| --- | --- |
| 请求只收 **`modelId`** | 不接受前端传 `modelName` 作为计费依据 |
| 后端解析 | `modelId` → catalog → `model.Type` 写入 `logs.model_name` |
| 白名单 | 模型须 **enabled**，且在 Key 白名单 ∩ 部门允许模型内 |
| 展示 | 响应 `model` 字段返回解析后的 `model.Type`（如 `gpt-4o-mini`） |

---

## 5. API 契约（提案）

### 5.1 提交模拟消耗

```http
POST /api/dev/simulate-consume
Cookie: session=...
Content-Type: application/json

{
  "platformKeyId": "<synced-key-id>",
  "modelId": 2,
  "inputTokens": 12000000,
  "outputTokens": 8000000
}
```

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `platformKeyId` | 否 | 省略时取 options 中第一个 synced Key |
| `modelId` | **是** | catalog ID（2 = gpt-4o-mini） |
| `inputTokens` | **是** | 整数 ≥ 1 |
| `outputTokens` | **是** | 整数 ≥ 0 |

**权限**：`billing:read` 或 `keys:admin`。

**响应 202 Accepted**：

```json
{
  "logId": 900042,
  "platformKeyId": "plk-sync-1",
  "model": "gpt-4o-mini",
  "inputTokens": 12000000,
  "outputTokens": 8000000,
  "quota": 4400000,
  "estimatedPoint": 6600,
  "estimatedDisplay": 6.6,
  "idempotencyKey": "newapi:900042",
  "status": "queued",
  "pollUrl": "/api/dev/simulate-consume/900042"
}
```

**同步错误码**（提交阶段）：

| HTTP | 场景 |
| --- | --- |
| 404 | 非 `DEPLOY_ENV=local` |
| 400 | ingest 未启用 |
| 404 | Platform Key / Model 不存在 |
| 422 | Key 无 synced mapping / 无 `newapi_key_id` |
| 422 | 模型未启用或不在 Key 白名单 |
| 401 | 未登录 |

> **钱包不足、预算 guard 等**发生在 Worker 入账阶段，**不在 POST 同步返回**；由 poll 接口 `status: failed` + `lastError` 体现。

### 5.2 轮询入账状态

```http
GET /api/dev/simulate-consume/{logId}
Cookie: session=...
```

**响应 200**：

```json
{
  "logId": 900042,
  "idempotencyKey": "newapi:900042",
  "status": "queued",
  "lastError": null,
  "ledgerAmount": null
}
```

`status` 枚举：

| 值 | 含义 |
| --- | --- |
| `queued` | job 仍在 pending 或尚未 claim |
| `ingested` | `usage_ledger` 存在 `newapi:{logId}`；可带 `ledgerAmount` |
| `failed` | job dead 或 ingest 失败；`lastError` 有文案 |

实现可查：`ingest_jobs` 状态 + `usage_ledger.idempotency_key` 是否存在。

### 5.3 初始化选项

```http
GET /api/dev/simulate-consume/options
```

**Platform Keys**（全部须满足）：

- `status = active`
- `platform_key_mappings.sync_status = synced`
- `newapi_key_id IS NOT NULL`

**Models**（按所选 Key 动态过滤）：

- `enabled = true`
- 在 Key 白名单 ∩ 部门路由允许范围内

**响应示例**：

```json
{
  "defaultPlatformKeyId": "plk-sync-1",
  "platformKeys": [
    { "id": "plk-sync-1", "name": "张三-开发调试", "newapiKeyId": 42 }
  ],
  "models": [
    { "modelId": 2, "type": "gpt-4o-mini", "name": "GPT-4o Mini", "inputPrice": 150, "outputPrice": 600 }
  ],
  "defaults": {
    "modelId": 2,
    "inputTokens": 12000000,
    "outputTokens": 8000000
  }
}
```

无 synced Key 时：`platformKeys: []`，前端显示空态 + 引导创建 Key。

### 5.4 幂等

- 每次提交 INSERT 新 `log_id` → **`newapi:{log_id}`**
- 重复点击提交 = 多笔独立消耗（符合测试预期）

---

## 6. UI 方案：Popup

### 6.1 入口

Header Dev 工具栏（`header-dev-backend-chrome.tsx`），与「Switch member」并列。

```text
┌──────────────────────────────────────────────────────────────┐
│  成本看板              [管理员] [模拟消耗] [Switch member]      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────────┐
                    │  模拟消耗                │
                    ├─────────────────────────┤
                    │ Platform Key  [▼]       │
                    │ 模型          [▼]       │
                    │ Input tokens  [12000000]│
                    │ Output tokens [8000000] │
                    │ 预估消耗：≈ ¥6.60        │
                    │  [取消]  [提交]          │
                    └─────────────────────────┘
                              │ 提交后
                              ▼
                    Toast：已入队 log #900042，正在入账…
                    （轮询 pollUrl，最多 ~15s）
                              │
              ingested ───────┴─────── failed
                 │                      │
     Toast：入账成功，约 ¥6.60          Toast：入账失败：{lastError}
     invalidate + 可选跳转 /audit/calls   Popup 保持打开
```

### 6.2 字段默认值

| 字段 | 默认值 | 校验 |
| --- | --- | --- |
| **Platform Key** | options 中第一个 synced Key | 无 Key 时空态，不可提交 |
| **模型** | options.defaults.modelId | enabled + Key 白名单 |
| **Input tokens** | `12_000_000` | 整数 ≥ 1 |
| **Output tokens** | `8_000_000` | 整数 ≥ 0 |
| **预估消耗** | 实时计算 | 公式与 §4.4 一致；**以后端 POST 响应为准** |

### 6.3 交互流程

1. 打开 Dialog → `GET options` + `GET capabilities`
2. 修改模型 / Key / tokens → 前端刷新预估
3. 提交 → `POST simulate-consume` → 收到 `status: queued`
4. **轮询** `pollUrl`（建议 1s 间隔，最多 15 次）：
   - `ingested` → 成功 Toast、关闭 Popup、`invalidate` wallet/audit/budget/dashboard、可选跳转 `/audit/calls`
   - `failed` → 错误 Toast、`lastError` 展示、Popup 不关闭
   - 超时仍 `queued` → Toast「入账超时，请到调用审计或 ingest metrics 排查」
5. **禁止**在收到 `queued` 后立即 Toast「已成功扣费」

### 6.4 组件组织

| 文件 | 职责 |
| --- | --- |
| `features/dev/components/simulate-consume-dialog.tsx` | Popup + 空态 + 轮询 UI |
| `features/dev/hooks/use-simulate-consume-dialog.ts` | 表单、mutation、poll |
| `features/dev/lib/estimate-consume.ts` | 前端预估（与 §4.4 对齐） |
| `api/dev.ts` | `simulateConsume`、`pollSimulateConsume`、`getSimulateOptions` |
| `header-dev-backend-chrome.tsx` | 入口按钮 |

---

## 7. 后端实现清单

| 模块 | 改动 |
| --- | --- |
| `internal/config/deploy.go` | `IsLocalDeploy()` |
| `internal/store/log_repo.go` + `postgres/log_repo.go` | 生产 `InsertConsumeLog` |
| `internal/domain/usage/simulate.go` | 算 quota、insert log、enqueue；`PollSimulateStatus` |
| `internal/http/handler/dev/handler.go` | POST / GET options / GET poll |
| `internal/http/handler/register.go` | 仅 local 注册 |
| `tests/handler/dev/simulate_consume_test.go` | queued → worker → ingested；failed 路径 |

---

## 8. 环境准备

```bash
pnpm start:postgres

# apps/backend/.env
DEPLOY_ENV=local
BOOTSTRAP_MODE=demo
CLOCK_ANCHOR=2026-06-19
DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable
LOG_DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5432/logs?sslmode=disable
NEW_API_WEBHOOK_SECRET=tokenjoy-webhook-secret
NEW_API_ENABLED=true
NEW_API_BASE_URL=http://127.0.0.1:3000

pnpm start:newapi    # Key sync 需要
pnpm start
```

登录：`admin@example.com` / `demo1234`  
账期：**`2026-06`**

**首次使用**：在 `/keys/platform` 新建 Key，确认 `platform_key_mappings.newapi_key_id` 存在且 `sync_status=synced`。

---

## 9. 验收步骤

### 9.1 Popup 基本流程

| # | 操作 | 预期 |
| --- | --- | --- |
| 1 | 打开 Popup | 有 synced Key；无则空态引导 |
| 2 | 默认 gpt-4o-mini，12M/8M tokens | 预估 ≈ **¥6.60** |
| 3 | 提交 | Toast「已入队」；**非**「已成功扣费」 |
| 4 | 轮询至 `ingested` | 成功 Toast；`ledgerAmount ≈ 6600` point |
| 5 | `/audit/calls` | 新行 tokens = 12M/8M，model = gpt-4o-mini |
| 6 | `/wallet` | 余额降约 ¥6.6 |
| 7 | `/budget`（2026-06） | consumed 上升 |

### 9.2 公式与变量

| # | 操作 | 预期 |
| --- | --- | --- |
| M1 | 同 tokens，gpt-4o vs gpt-4o-mini | 前者 `ledgerAmount` 更大 |
| M2 | 同模型，仅增大 output | `ledgerAmount` 随 output 单价升 |
| M3 | POST 响应 `quota` / `estimatedPoint` | 与 §4.4 示例 B 一致 |

### 9.3 Ingest 链路

| # | 操作 | 预期 |
| --- | --- | --- |
| I1 | poll → `ingested` | `idempotencyKey = newapi:{logId}` |
| I2 | ledger `source` | `webhook` |
| I3 | ingest metrics | `ingest_jobs_pending` → 0 |

### 9.4 失败路径

| # | 操作 | 预期 |
| --- | --- | --- |
| F1 | 选 dept-3 下 Key + 超大 tokens（超预算/钱包） | poll → `failed`，`lastError` 可读 |
| F2 | Worker 未启动 | poll 长期 `queued` 或超时提示 |

### 9.5 投影延迟

| 页面 | 时机 |
| --- | --- |
| `/audit/calls` | poll `ingested` 后立即可见 |
| `/wallet` | 与 ingest 同事务，同上 |
| `/budget`、`/dashboard/*` | 可能再晚 5–15s（River 投影） |

---

## 10. 实现前的临时方案

手动 SQL + webhook，或 `pnpm verify:integration`（含真实 webhook HTTP）。

---

## 11. 小结

| 项 | 结论 |
| --- | --- |
| **交互** | Popup：Key + modelId + input/output tokens |
| **入账** | logs INSERT → queue → Worker → ledger |
| **API 语义** | POST → **`queued`**；poll → **`ingested` / `failed`** |
| **默认 token** | 12M / 8M（gpt-4o-mini ≈ ¥6.6） |
| **生产等价** | 覆盖 queue + Worker + ledger；**不**覆盖 webhook HTTP |
| **当前状态** | **未实现** |
