# NewAPI 集成现状与优化点

> **定位**：对照 `apps/newapi` 与 `apps/backend`，描述 **当前接通能力** 与 **可优化点**。  
> 上线前可勾选 backlog 见 [plan.md](./plan.md) §1。  
> 命名与包边界（Gateway / NewAPISync / PlatformKey）见 [Backend-架构.md](./Backend-架构.md) §0。

---

## 1. 现状摘要

| 维度 | 状态 |
| --- | --- |
| PlatformKey / NewAPIKey 生命周期（Create / Update / Toggle / Revoke / Rotate / Delete） | 已接通 |
| Gateway Precheck + `/v1` 反代 | 已接通 |
| 用量回写（webhook + reconcile） | 已接通 |
| 额度同步（wallet_sync / TopUp / GetUserQuota） | 已接通 |
| Provider → Channel Upsert | 已接通 |
| 从 NewAPI 拉模型目录 / GetGroups | 未做（TokenJoy 自管目录；推 `model_limits`） |
| 管理端 API 封装 | NewAPIKey + User/Quota + Channel；无 GetAllModels / GetGroups / GetLogByKey |

---

## 2. 写路径与 Gateway 约定

### 2.1 Platform Key

| 操作 | 模式 |
| --- | --- |
| Create / Approve→Create / Toggle / Revoke / Rotate / Delete | 同步 **Remote-first**（先 NewAPI，再写 Postgres）；NewAPI 关 → `503`，DB 不变 |
| Update 配额 / 白名单 | `requireNewAPI` + **先写 DB 再 `SyncUpdatePlatformKey`，失败回滚**（非 async outbox） |
| Rebalance / ModelLimits / Provider Channel | 仍可走 newapi_sync outbox → Worker |

Rotate 使用 NewAPI `POST /api/token/{id}/regenerate`（patch），保持 `newapi_key_id` 不变以利 ingest。

### 2.2 Gateway

- 单例 `ReverseProxy`（`NewGatewayService` 创建一次）
- 精确路径白名单：`/v1/chat/completions`、`/v1/completions`、`/v1/embeddings`、`/v1/models`
- Body 上限：`MaxBytesReader` 4MB
- Precheck：key / 公司 / 部门账期与预算 / 配额快照 / NewAPI remain / wallet cap / wallet_sync 漂移 / 模型白名单

### 2.3 Worker outbox

`IsPermanentNewAPISyncOutboxError` 将 `503`（含 `newapi not enabled`）及不可恢复错误标为永久 `failed`，不再无限重试。

### 2.4 入账（方案 B）

NewAPI notify → `POST /api/internal/webhooks/newapi-log` → **入队** `ingest_jobs`（pending）并立即 `200 accepted`；Worker `ProcessPending` 异步 `IngestByLogID`；另有 `reconcile_cursors` 水位补洞（直读 `LOG_DATABASE_URL` → `newapi.logs`）。

### 2.5 设计约束（明确不做）

| 项 | 原因 |
| --- | --- |
| delete+create 式 Rotate | 破坏 ingest `newapi_key_id` 连续性 |
| Toggle 改回 async outbox | 用户操作应同步可见 |
| 「无 NewAPI 的 Platform Key」 | 统一要求 NewAPI；关则 `503` |
| Gateway 用 `HasPrefix` 放行 | 精确匹配为安全目标 |

---

## 3. 可优化点

| 问题 | 位置 | 说明 |
| --- | --- | --- |
| Update 非严格 Remote-first | `platform_key_update.go` | 先写 DB 再 sync；失败可回滚，崩溃窗口仍可能短暂不一致 |
| `noopWalletService` `AvailableQuota` 恒 0 | `domain/company/wallet.go` | NewAPI 关闭时 Gateway 预检 / `wallet_sync` 失效 |
| demo 下 Gateway 组合无校验 | `config.go` | 只开 Gateway、不开 `NEW_API_ENABLED` → 路由不挂载，进程仍启动 |
| Rebalance / Overrun NewAPI 关闭时空转 | `overrun.go`、lifecycle enqueue | Worker 侧可能 `return nil`，掩盖误配 |
| `NOTIFY_WEBHOOK_URL` 失败静默 | `infra/notification/service.go` | HTTP 失败写 log，`Send()` 仍 `return nil` |
| `processOrgSync` 固定 `DefaultCompanyID` | `org_sync_processor.go` | SaaS 多企业 org 同步范围受限 |
| `NEW_API_PUBLIC_URL` 未使用 | Backend `config.go` | 配置冗余（仅 newapi 侧示例） |
| `host.docker.internal` 跨平台 | `apps/newapi/docker-compose.yml` | Linux 非 Docker Desktop webhook 常连不上 Backend |
| `gate-verify` 不测 Backend Gateway | `gate-verify.sh` | 脚本通过 ≠ `/v1/*` Gateway 可用 |
| `ingest_notify_total` 幂等重复也 +1 | `ingestmetrics/collector.go` | 宜仅首次 ledger 插入计数 |
| `GET /internal/metrics/ingest` 无鉴权 | ingest metrics | 生产可加 secret 或仅 bind localhost |

上线前修复项与联调签字见 [plan.md](./plan.md) §1。

---

## 4. 联调注意坑

- `host.docker.internal`：Linux 需改 `MANAGEMENT_WEBHOOK_URL` 或加 `extra_hosts`
- `pnpm gate:verify` 不覆盖 Backend Gateway，Gateway 须单独验
- 只开 `NEW_API_GATEWAY_ENABLED` 不会生效，须同时 `NEW_API_ENABLED=true`
- `DEPLOY_ENV=production` 时 NewAPI + Gateway + 入账为 **启动硬依赖**（见 [Backend-配置架构.md](./Backend-配置架构.md)）
