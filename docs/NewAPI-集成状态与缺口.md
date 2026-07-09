# NewAPI 集成缺口

> **最后对齐**：2026-07-08  
> **定位**：仅记录 **尚未解决** 的问题。可勾选 backlog 见 [plan.md](./plan.md) §1。

---

## P0 — 生产功能或数据一致性

| 问题                                  | 位置                                                | 后果                                                                                                                         |
| ------------------------------------- | --------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Relay 关闭时 lifecycle **静默 no-op** | `domain/relay/lifecycle_ops.go`（`Enabled()` 早退） | Toggle / Revoke / Update、Worker outbox 可能 **DB 与 NewAPI 不一致**；创建/审批已有 `requireRelay()` → 503，其余写路径无护栏 |
| `noopWalletService` `AvailableQuota` 恒 0 | `domain/company/wallet.go`                          | Relay 关闭时 Gateway 预检 / `wallet_sync` 失效；`GetWallet` 读 Postgres lot，不受 noop 影响 |

---

## P1 — 误配、SaaS、可观测

| 问题                                     | 位置                                           | 后果                                                              |
| ---------------------------------------- | ---------------------------------------------- | ----------------------------------------------------------------- |
| `NEW_API_PUBLIC_URL` 未使用              | `config.go`                                    | 配置冗余；控制台/文档无法与对外 Relay URL 对齐                    |
| `RELAY_GATEWAY_ENABLED` 无启动级组合校验 | `config.go` / `router.go`                      | 只开 Gateway、不开 `NEW_API_ENABLED` → 路由不挂载，进程仍正常启动 |
| `wireGatewayService` 失败静默            | `app/registry.go`                              | 装配失败时 `RelayGateway == nil`，仅 Router 打 error 日志         |
| Rebalance / Overrun Relay 关闭时空转     | `domain/budget/overrun.go`、`lifecycle_ops.go` | 队列可入队，Worker 侧 `Enabled()` 早退 `return nil`               |
| `NOTIFY_WEBHOOK_URL` 失败对调用方静默    | `infra/notification/service.go`                | HTTP 失败写 log + warn，但 `Send()` 仍 `return nil`               |
| `processOrgSync` 固定 `DefaultCompanyID` | `infra/worker/org_sync_processor.go`           | SaaS 多企业下定时 org 同步只覆盖默认企业                          |
| `host.docker.internal` 跨平台            | `apps/newapi/docker-compose.yml`               | Linux 非 Docker Desktop 本地 webhook 常连不上 Backend             |
| `gate-verify` 不测 Backend Gateway       | `apps/newapi/scripts/gate-verify.sh`           | 脚本通过 ≠ `/v1/*` Gateway 可用                                   |

---

## P2–P3 — 清理与文档

| 问题                                  | 位置                               | 建议                                                             |
| ------------------------------------- | ---------------------------------- | ---------------------------------------------------------------- |
| `NEW_API_PUBLIC_URL` 落地或删除       | —                                  | 与 P1 重复项，二选一做完即可                                     |
| `ingest_notify_total` 幂等重复也 +1   | `infra/ingestmetrics/collector.go` | 仅首次 ledger 插入计数，或改名并文档化                           |
| `OutboxKindRebuildAbilities` 等死常量 | `store/relay.go`                   | 清理或接 Worker（`RebuildAbilities` 现仅在 models 同步路径调用） |
| `Backend-架构.md` 等子文档            | `docs/`                            | 与实现扫尾对齐                                                   |

---

## 建议修复顺序

```
P0
├── Relay 关闭时：所有 mutating 管理面操作报错或禁止（不仅 create/approval）
└── noop 钱包：Gateway 启用时 fast-fail，或区分 Relay 未配置与余额为零

P1
├── config：Gateway 组合校验 + wireGatewayService 失败即启动失败
├── SaaS：org_sync 按企业迭代（若产品需要）
└── gate-verify 增加 Backend /v1 Gateway 步骤

P2
├── NEW_API_PUBLIC_URL 落地或删除
├── Rebalance/Overrun、通知 webhook 失败可观测性
└── Linux 本地 webhook URL 文档化或 compose 可配置

P3
├── ingest metrics 语义
└── 死常量 / 子文档清理
```

---

## 联调时仍须注意的坑

- `host.docker.internal`：Linux 需改 `MANAGEMENT_WEBHOOK_URL` 或加 `extra_hosts`
- `pnpm gate:verify` 不覆盖 Backend Gateway，Gateway 须单独验
- 只开 `RELAY_GATEWAY_ENABLED` 不会生效，须同时 `NEW_API_ENABLED=true`
