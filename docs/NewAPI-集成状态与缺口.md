# NewAPI 集成缺口

> **最后对齐**：2026-07-09  
> **定位**：仅记录 **尚未解决** 的问题。可勾选 backlog 见 [plan.md](./plan.md) §1。

---

## 已关闭（2026-07-09 P0 必须项）

| 项 | 说明 |
| --- | --- |
| Gateway `/v1` path 透传 | `gateway_service.go` Director 显式保留客户端 path |
| Keys Remote-first | Toggle / Revoke / Rotate 先调 NewAPI，成功后再写 Postgres |
| Lifecycle 同步 503 | `SyncCreate` / `TrySyncCreate` / `SyncUpdate` / `SyncRevoke` / `SyncRotate` 在 Relay 关闭时返回 503 |
| Platform Key Rotate | `regenerate` 补丁 + `SyncRotatePlatformKey` + HTTP 200 |
| prod 启动校验 | `APP_PROFILE=prod` 强制 Relay + Gateway + 入账配置；`NEW_API_BASE_URL` 禁止带 path |
| `wireGatewayService` 失败 | 装配错误时进程启动失败（`panic`） |

---

## P0 — 生产功能或数据一致性

| 问题 | 位置 | 后果 |
| --- | --- | --- |
| `UpdatePlatformKey` 仍 DB-first + async outbox | `platform_key_update.go` | Relay 开但 sync 失败时，配额/白名单可能与 NewAPI 短暂不一致 |
| `DeletePlatformKey` 仅删 Postgres | `platform_key_actions.go` | 未 revoke Remote token（产品若保留 Delete 需补 Remote-first） |
| Worker outbox Relay 关闭 | `relay_processor.go` | 配置错误时条目仍重试，未永久 failed |
| `noopWalletService` `AvailableQuota` 恒 0 | `domain/company/wallet.go` | Relay 关闭时 Gateway 预检 / `wallet_sync` 失效 |

---

## P1 — 误配、SaaS、可观测

| 问题 | 位置 | 后果 |
| --- | --- | --- |
| `NEW_API_PUBLIC_URL` 未使用 | `config.go` | 配置冗余 |
| demo 下 Gateway 组合无校验 | `config.go` | 只开 Gateway、不开 `NEW_API_ENABLED` → 路由不挂载，进程仍启动 |
| Rebalance / Overrun Relay 关闭时空转 | `overrun.go`、`lifecycle_ops.go` | Worker 侧 `Enqueue*` 仍 `return nil` |
| `NOTIFY_WEBHOOK_URL` 失败对调用方静默 | `infra/notification/service.go` | HTTP 失败写 log，但 `Send()` 仍 `return nil` |
| `processOrgSync` 固定 `DefaultCompanyID` | `infra/worker/org_sync_processor.go` | SaaS 多企业 org 同步范围受限 |
| `host.docker.internal` 跨平台 | `apps/newapi/docker-compose.yml` | Linux 非 Docker Desktop webhook 常连不上 Backend |
| `gate-verify` 不测 Backend Gateway | `apps/newapi/scripts/gate-verify.sh` | 脚本通过 ≠ `/v1/*` Gateway 可用 |

---

## P2–P3 — 清理与文档

| 问题 | 位置 | 建议 |
| --- | --- | --- |
| Gateway singleton proxy / 精确路径白名单 | `gateway_service.go` | 性能与安全加固，非阻塞 |
| Gateway body 上限 | `gateway_service.go` | `MaxBytesReader` 可配置常量 |
| `ingest_notify_total` 幂等重复也 +1 | `infra/ingestmetrics/collector.go` | 仅首次 ledger 插入计数 |
| `Backend-架构.md` 等子文档 | `docs/` | 与实现扫尾对齐 |

---

## 建议修复顺序

```
P0 剩余
├── UpdatePlatformKey Remote-first（或失败可观测）
└── Worker：Relay 关闭时 outbox 永久 failed

P1
├── gate-verify 增加 Backend /v1 Gateway 步骤
└── demo Gateway 组合校验（可选）

P2
├── Gateway 硬化（singleton、精确路径、body limit）
└── Delete 语义（必须先 revoke）
```

---

## 联调时仍须注意的坑

- `host.docker.internal`：Linux 需改 `MANAGEMENT_WEBHOOK_URL` 或加 `extra_hosts`
- `pnpm gate:verify` 不覆盖 Backend Gateway，Gateway 须单独验
- 只开 `RELAY_GATEWAY_ENABLED` 不会生效，须同时 `NEW_API_ENABLED=true`
- `APP_PROFILE=prod` 时上述三联开为 **启动硬依赖**
