# NewAPI 集成状态

TokenJoy 与 NewAPI 的集成现状与已落地能力。入账终态以共享 `logs` 库 + notify 主路径 + reconcile 兜底为准（方案 B）。**剩余缺口与联调 checklist** 见 [plan.md](./plan.md) §1。

---

## 1. 当前架构（已落地）

```
NewAPI settle
  ① COMMIT → logs.newapi.logs
  ② management.EnqueueNotify(log_id) ──POST──► Backend /api/internal/webhooks/newapi-log
                                                      │
                                                      ▼
                                            IngestByLogID → usage_ledger

Backend Worker（LOG_DATABASE_URL 非空即跑，与 NEW_API_ENABLED 无关）
  · 每 5s：ingest_failures retry
  · 每 300s + 启动：reconcile 按全局 cursor 补洞
```

| 原则    | 说明                                                      |
| ------- | --------------------------------------------------------- |
| 单入口  | notify / reconcile / retry 均走 `IngestByLogID`           |
| DB SSOT | raw 字段只从 `logs` 库读，不再 HTTP `GET /api/log/`       |
| 幂等    | `newapi:{log_id}`                                         |
| 失败    | 业务错写 `ingest_failures` + retry；不再用 webhook outbox |

**NewAPI 侧**：`apps/newapi/patches/management/notify.go` + Dockerfile 打 patch；`settle_webhook.sh` 仅本地调试，非生产路径。

**Backend 侧**：`internal_ingest_handler.go`（webhook + metrics）、`IngestWorker`、`OutcomeFor` + `FailureRecorder`。

---

## 2. 已完成（自旧审计移除）

以下项曾在 [历史] `NewAPI-集成问题与方案.md` / `Backend-集成缺口审计.md` 中列为缺口，**已实现**：

| 类别       | 项                                                                                 |
| ---------- | ---------------------------------------------------------------------------------- |
| 入账主路径 | NewAPI settle 后 `EnqueueNotify` → Backend 瘦 payload `{log_id}`                   |
| 入账兜底   | `processReconcile` + `backend.reconcile_cursors` 全局水位                          |
| 失败重试   | `ingest_failures` + `ClaimPendingFailures` lease 占位                              |
| 删除旧路径 | `compensateLogs`、`ListLogs`（Admin HTTP）、`webhook outbox`、`relay_sync_cursors` |
| Worker     | `ingestLoop` 与 `relayLoop` 拆分；入账不绑 `NEW_API_ENABLED`                       |
| 数据层     | 共享 `logs` 库 DDL 单一源；`UpsertFailure` 冲突不重置 attempts/dead                |
| 模块化     | `IngestOutcome`、`FailureRecorder`、`IngestWorker`                                 |
| 配置       | `LOG_DATABASE_URL` 非空时强制 `NEW_API_WEBHOOK_SECRET`                             |

多租户入账 cursor 问题随 **全局 `reconcile_cursors`** 与 **按 token mapping 解析企业** 一并解决，不再依赖 per-company `GetLastLogID`。

---

## 3. 剩余缺口

### 3.1 高 — 生产功能或资金风险

| #   | 问题                            | 位置                                         | 后果                                                                      |
| --- | ------------------------------- | -------------------------------------------- | ------------------------------------------------------------------------- |
| 1   | Gateway 代理剥离 `/v1`          | `domain/relay/gateway_service.go`            | 客户端 `/v1/chat/completions` 可能被转成 `/chat/completions` → NewAPI 404 |
| 2   | 充值跳过 TopUp 仍标 `topped_up` | `domain/billing/service.go` `topUpAndFinish` | `NEW_API_ENABLED=false` 时 DB 已充值、NewAPI 钱包未增加                   |
| 3   | Relay 关闭时 Key 同步静默跳过   | `domain/relay/lifecycle_ops.go`              | DB 有 Key、NewAPI 无 token，无告警                                        |

### 3.2 中 — 误配、SaaS、可观测

| #   | 问题                                     | 后果                                           |
| --- | ---------------------------------------- | ---------------------------------------------- |
| 4   | `NEW_API_PUBLIC_URL` 未使用              | 配置冗余，对外 Relay URL 无法与此对齐          |
| 5   | `RELAY_GATEWAY_ENABLED` 无组合校验       | 只开 Gateway 不开 NewAPI → 路由不挂载，仅 log  |
| 6   | `wireGatewayService` 失败静默            | `registry.go` 吞错，`relayGateway == nil`      |
| 7   | Rebalance / Overrun 在 Relay 关闭时空转  | ingest 仍入队，Worker 调用时 `return nil`      |
| 8   | `noopWalletService` 余额恒 0             | Gateway 预检 403，`GetWallet` 不区分「未配置」 |
| 9   | 通知 `NOTIFY_WEBHOOK_URL` 失败静默       | HTTP 失败仍 `return nil`，调用方无感知         |
| 10  | `processOrgSync` 固定 `DefaultCompanyID` | SaaS 多企业 org 同步范围受限                   |
| 11  | `host.docker.internal` 跨平台            | Linux 非 Docker Desktop 时常不可用             |
| 12  | `gate-verify` 不测 Backend Gateway       | 验证通过 ≠ Gateway 可用                        |

### 3.3 低 — 清理与文档

| #   | 问题                                                                       | 建议                                        |
| --- | -------------------------------------------------------------------------- | ------------------------------------------- |
| 13  | `ingest_notify_total` 幂等重复也 +1                                        | 仅首次 ledger 插入时计数，或改名并文档化    |
| 14  | `GET /internal/metrics/ingest` 无鉴权                                      | 生产可加 webhook secret 或仅 bind localhost |
| 15  | `Backend-架构.md` / `Backend-预算.md` 仍可能描述旧 webhook/compensate 路径 | 与本文及实现代码对齐                        |
| 16  | `OutboxKindRebuildAbilities` 等死常量                                      | 清理或接 Worker 分支                        |

---

## 4. 修复优先级（建议）

```
P0
├── Gateway 保留完整 /v1 path（或按 base URL 条件剥离）
├── 充值：NewAPI 未调用 TopUp 时禁止 topped_up
└── Relay 关闭时管理面明确报错或 demo 标记（Key 创建）

P1
├── config：Gateway 隐含 NEW_API_ENABLED；wireGatewayService 失败 fast-fail
└── SaaS：org_sync 按企业迭代（若产品需要）

P2
├── NEW_API_PUBLIC_URL 落地或删除
├── Rebalance/Overrun、钱包 noop、通知 webhook 失败可观测性
└── gate-verify 增加 Backend /v1 Gateway 步骤

P3
├── ingest metrics 语义与鉴权
└── 架构子文档扫尾
```

---

## 5. 联调检查清单

**入账（方案 B）**

- [ ] `LOG_DATABASE_URL` 指向 `logs` 库；init 已建 `newapi` / `backend` schema
- [ ] `NEW_API_WEBHOOK_SECRET` 与 NewAPI `MANAGEMENT_WEBHOOK_SECRET` 一致
- [ ] NewAPI `LOG_SQL_DSN` → `logs`；patch 镜像已 build（非纯上游镜像）
- [ ] Backend Worker 已启动（reconcile / failure retry 依赖 Worker）
- [ ] `GET /api/internal/metrics/ingest` 可查看 `ingest_reconcile_gaps`、`ingest_failures_pending`

**Relay / 管理面（与入账独立）**

- [ ] `NEW_API_ENABLED=true` + `NEW_API_BASE_URL` + `NEW_API_ADMIN_TOKEN`（Key 同步、充值 TopUp）
- [ ] 若开 Gateway：`RELAY_GATEWAY_ENABLED=true` 且 **P0 path 问题已修**
- [ ] 不以 `settle_webhook.sh` 或 compose 里仅配置 URL 作为「notify 已接通」依据 — 以真实 POST `{log_id}` 与 ledger 为准

**本地**

- [ ] `pnpm start`：默认无 NewAPI，入账靠测试 mock / memory LogStore
- [ ] `pnpm start:relay`：完整栈；Backend 需配置 `LOG_DATABASE_URL` 与 webhook secret

---

## 6. 一句话总结

**入账主路径已切到方案 B：NewAPI notify + 共享 logs 库 + Backend 单入口 Ingest；旧 compensateLogs / HTTP 拉 log 已删除。** 剩余风险主要在 **Gateway 路径、充值状态、Relay 关闭时的静默成功**，以及 **配置校验与 SaaS 边界**，与入账链路本身已解耦。
