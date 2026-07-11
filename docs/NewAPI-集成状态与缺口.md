# NewAPI / Gateway 未完成项

> **定位**：仅记录 **仍待做** 的 NewAPI / Gateway 缺口；完成即从本文删除。  
> **Backlog 主入口**：[plan.md](./plan.md) §1。  
> **已接通架构**（生命周期、Gateway precheck、入账、Worker、Remote-first Create 等）见 [Backend-架构.md](./Backend-架构.md) · [Backend-Ingest架构.md](./Backend-Ingest架构.md) · [Backend.md](./Backend.md)。

---

## 1. 上线前 Fix

| 项 | 位置 | 说明 |
| --- | --- | --- |
| `NOTIFY_WEBHOOK_URL` 失败不可观测 | `infra/notification/service.go` | HTTP 失败仅写 log，`Send()` 仍 `return nil`；调用方无法感知、无法重试 |
| Update 非严格 Remote-first | `platform_key_update.go` | 配额 / 白名单更新为 **先写 DB 再 `SyncUpdatePlatformKey`，失败回滚**；崩溃窗口仍可能短暂不一致。若上线要求与 Create 路径铁律一致，改为先 Remote |

---

## 2. 可选 / 延后

| 项 | 位置 | 说明 |
| --- | --- | --- |
| 审批完整 outbox / `provisioning` | `keys/approval.go` | sync 失败补偿（`revertKeyApproval`）已实现；与 `OutboxKindCreateKey` 统一的完整 outbox 仍可选，见 [plan.md](./plan.md) §3 |
| `NEW_API_PUBLIC_URL` 未使用 | Backend `config.go` | 配置冗余，可删或接入文档 / 对外 URL 展示 |
| NewAPI notify 队列 drop 不可观测 | NewAPI 侧 | 内存队列有界，满则 drop；reconcile 可兜底，缺 drop 计数 / 告警 |
| 入账 enqueue→ledger 延迟 | Backend metrics | 有 `ingest_lag_seconds` / pending；无 enqueue→ledger 直方图（见 [Backend-Ingest架构.md](./Backend-Ingest架构.md) §13.2） |
| Gateway 预检 estimate | `gateway/precheck.go` | 固定最小值；未按模型单价动态估价 |
| mapping 缺失自愈 | ingest | mapping 缺失时拒绝入账；严格审计前提下是否自动重建待评估 |

---

## 3. 管理端 API 未封装（按产品需要）

TokenJoy 自管模型目录并向 NewAPI 推 `model_limits`；下列 Admin API **未** 封装，仅在需要「从 NewAPI 拉目录 / 按 Key 查 log」时再补：

| API | 用途 |
| --- | --- |
| GetAllModels | 从 NewAPI 拉模型目录（当前 TokenJoy 侧维护） |
| GetGroups | NewAPI 分组列表 |
| GetLogByKey | 按 Platform Key 查 NewAPI 消费 log |

---

## 4. PRD 相关（非 plan §1 P0）

| 项 | 说明 |
| --- | --- |
| Anthropic `/v1/messages` | Gateway 白名单为 OpenAI 风格 `/v1/*` 精确路径；PRD 原生 Anthropic 格式未作为一等契约验收 |
| `overrun_policy.blockMessage` | 配置可持久化，Gateway 403 文案未完全消费该字段 |

---

## 5. 联调签字（发布门禁）

自动化脚本已就绪；**须在真实 full-stack 环境跑通** 方可视为上线签字完成：

```bash
# 前提：Backend full-stack .env（见 apps/backend/.env.example）、Backend 已启动
pnpm verify:gate           # 通路冒烟（自建 Key + Gateway + webhook）
pnpm verify:integration    # ledger + Toggle/Rotate/Revoke + metrics（需 NEW_API_ADMIN_TOKEN、psql）
pnpm verify                # lint + test + build（CI；不含上面两条）
```

脚本：[apps/newapi/scripts/_verify-lib.sh](../apps/newapi/scripts/_verify-lib.sh) · [gate-verify.sh](../apps/newapi/scripts/gate-verify.sh) · [integration-verify.sh](../apps/newapi/scripts/integration-verify.sh)

| 环境注意 | |
| --- | --- |
| Backend 须先启动 | `verify:gate` 起 newapi compose；Gateway 依赖 Backend `:8080` |
| `NEW_API_ADMIN_TOKEN` | 仅 `verify:integration` 需要（Rotate/Revoke 的 NewAPI Admin 断言） |
| Webhook secret | Backend `NEW_API_WEBHOOK_SECRET` = NewAPI `MANAGEMENT_WEBHOOK_SECRET` |
| webhook `200 accepted` | 仅代表入队；ledger 写入看 IngestWorker / metrics |
| 生产契约 | `DEPLOY_ENV=production` 时 NewAPI + Gateway + 入账为启动硬依赖（[Backend-配置架构.md](./Backend-配置架构.md) §7） |

**仍建议人工补验：** SaaS 多企业隔离（Worker 已遍历 active companies；跨租户 Gateway / ingest 需在真实多企业栈 smoke）。
