# Local-SaaS 架构实现计划

> 基于 [apps/docs/Local-SaaS-架构.md](../../apps/docs/Local-SaaS-架构.md) 的实现指南。
> 本文档面向开发者，描述**怎么做**，架构文档描述**做什么和为什么**。
>
> **状态：Phase 1 + Phase 2 已实现。Phase 3（前端体验）待做。**

---

## 实现清单

| # | 组件 | 状态 | 文件 |
|---|------|------|------|
| 1 | register-local 扩展（创建 user + company + member + token） | ✅ | `handler/platform/register_local.go` |
| 2 | Local Setup 持久化 platformKey + walletUserId | ✅ | `app/setup_server.go` |
| 3 | 启动时自动配置 tokenjoy-upstream channel | ✅ | `app/platform_channel.go` |
| 4 | SaaS wallet_lots endpoint | ✅ | `handler/platform/catalog_lots.go` |
| 5 | versions 响应增加 walletLots | ✅ | `handler/platform/models.go` |
| 6 | 充值/Ingest 自动 bump wallet_lots_version | ✅ | `billing/lot_confirm.go` + `usage/ingest.go` |
| 7 | Catalog Sync client: FetchWalletLots | ✅ | `integration/catalogsync/client.go` |
| 8 | Catalog Sync executor: version-gated syncWalletLots | ✅ | `worker/catalogsync/execute.go` |
| 9 | UpsertLotFromSync（lot 镜像写入） | ✅ | `store/postgres/billing_repo_lots.go` |
| 10 | RawConsumeLog.ChannelID + SQL 查询 | ✅ | `store/log_repo.go` + `postgres/log_repo.go` |
| 11 | Ingest channel-based lot consumption split | ✅ | `domain/usage/ingest.go` |
| 12 | Gateway wallet skip for selfhosted | ✅ | `domain/gateway/precheck.go` + `evaluate.go` |
| 13 | usage_ledger lot_id nullable（非平台渠道无 lot） | ✅ | `schema.sql` + `ledger_repo_write.go` |
| 14 | logs schema + channel_id 列 | ✅ | `logs_schema.sql` + `logs_schema_isolated.sql` |

---

## 关键设计决策（与原计划差异）

| 原计划 | 实际实现 | 理由 |
|--------|---------|------|
| 新增 `synced` lot kind | 直接镜像 SaaS lot（保留原始 kind） | 不需要新 enum，FIFO 直接对齐 |
| Catalog API 单一 endpoint 返回模型+余额 | 独立 endpoints + version 门控 | 各通道独立演进，无变化时跳过 |
| 每次都拉 lot | version 门控（wallet_lots_version） | 减少无谓网络 IO |
| Gateway 模型级 wallet 分流 | 公司级（selfhosted 跳过） | 实现简单，SaaS Gateway 兜底 |
| register-local 只返回 syncToken | 返回 companyId + walletUserId + platformKey + syncToken | 一次调用完成全部注册 |

---

## Phase 3（待做）

| # | 任务 | 说明 |
|---|------|------|
| 1 | 钱包页面 | 展示 TokenJoy 余额 + 自管渠道消耗统计 |
| 2 | 余额告警 | Catalog Sync 检测低于阈值 → 通知 |
| 3 | 自助充值跳转 | Local UI 嵌入 SaaS 充值页面 |
