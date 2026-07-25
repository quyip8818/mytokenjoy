# 多版本单代码库架构

本文档描述 TokenJoy monorepo 的现行架构，作为日常开发和维护的指导。

---

## 产品与部署

| 产品 | 部署形态 | 用户 | 核心职责 |
|------|---------|------|---------|
| **TokenJoy** | Local（私有化） | 客户管理员 | Key 管理、预算、组织、审计、自定义模型、Channel |
| **TokenJoy** | SaaS（公有云） | SaaS 客户 | Local 功能子集，无自定义模型/Channel |
| **SMS** | 中心部署 | TJ 运营团队 | 供应商管理、模型目录、合同、采购订单、定价发布 |

TokenJoy Local 和 SaaS 共用同一套前后端代码，运行时通过 `SUPPORT_SAAS` flag 分流。

---

## 开发环境：同机并行

### 设计原则

1. **共用基础设施容器** — 一个 Postgres、一个 Redis，省资源
2. **端口段隔离** — apps 用 xx10 段，sms 用 xx20 段
3. **数据库名隔离** — 同一 Postgres 不同 database，各自独立
4. **独立 reset** — `pnpm reset` 只动 apps 的库，`pnpm reset sms` 只动 sms 的库

### 端口总表

| 服务 | apps (TokenJoy) | sms |
|------|----------------|-----|
| Postgres | 5510 (共用) | 5510 (共用) |
| Redis | 6310 (共用) | 6310 (共用) |
| Backend | 8010 | 8020 |
| Frontend (Vite) | 5173 | 5174 |
| NewAPI | 3010 | 3020 |

### 数据库隔离

一个 Postgres 容器内 6 个 database：

| Database | Owner | 产品 | 用途 |
|----------|-------|------|------|
| tokenjoy | tokenjoy | apps | apps backend 主库 |
| newapi | tokenjoy | apps | apps NewAPI 数据 |
| logs | tokenjoy | apps | apps NewAPI 日志 |
| sms | sms | sms | sms backend 主库 |
| sms_newapi | sms | sms | sms NewAPI 数据 |
| sms_logs | sms | sms | sms NewAPI 日志 |

### Redis 隔离

| 产品 | Redis DB | 用途 |
|------|----------|------|
| apps NewAPI | db 0 | session/cache |
| sms NewAPI | db 1 | session/cache |
| apps backend | db 2 | budget cache |
| sms backend | db 3 | 预留 |

---

## docker-compose.yml（根目录）

统一编排：postgres、redis、newapi-apps、newapi-sms 四个 service。`pnpm infra` 一键启动。

---

## 开发命令

所有命令从仓库根目录执行：

```bash
pnpm start               # apps（backend + frontend + mock）
pnpm start sms           # sms（backend + frontend）
pnpm start all           # 并行
pnpm reset [local|saas]  # 重置 apps 库
pnpm reset sms           # 重置 sms 库
pnpm reset all           # 全部
pnpm infra               # 启动基础设施
pnpm test                # apps 测试
pnpm test:sms            # sms 测试
pnpm lint                # apps lint
pnpm lint:sms            # sms lint
```

### Reset 机制

- 使用 SQL `DROP/CREATE DATABASE`，不再用 `docker compose down -v`
- 每个 reset 只动自己的 database，不影响对方
- init 脚本（`scripts/postgres-init/01-create-all-dbs.sh`）仅在 volume 首次创建时执行
- 共享逻辑在 `scripts/lib/db-reset.sh`（`reset_apps_databases` / `reset_sms_databases`）

---

## 核心规则

1. **两个产品，两个边界** — `apps/` 是客户的，`sms/` 是内部的
2. **跨产品通信仅 HTTP API** — 不共享 DB、不共享 domain 实现代码
3. **共享层放 packages/** — 跨产品共享的契约、类型放这里
4. **制品交付** — 客户只拿 `apps/` 的 image，`sms/` 永不出门
5. **独立演进** — sms 不阻塞客户侧发版，反之亦然

### 跨产品约束

- apps/ 和 sms/ 之间禁止 Go import
- apps/ 和 sms/ 之间禁止 TypeScript import
- 共享类型/契约只能放 packages/contracts/
- 跨产品通信只能通过 HTTP API

---

## 文档管理

| 位置 | 内容 |
|------|------|
| `docs/` | 仓库级/跨产品文档（本文档、整体 Roadmap） |
| `apps/docs/` | TokenJoy 产品文档（架构、PRD、ADR、plan 等） |
| `sms/docs/` | SMS 产品文档 |

禁止在 apps/frontend/、apps/backend/、sms/frontend/、sms/backend/ 下新建 .md（README.md 除外）。

---

## 构建产物

| 产物 | 来源 | 交付 |
|------|------|------|
| tokenjoy/frontend | apps/frontend/ | 客户 |
| tokenjoy/backend | apps/backend/ | 客户 |
| tokenjoy/newapi | apps/newapi/ | 客户 |
| sms-ui | sms/frontend/ | 仅内部 |
| sms-backend | sms/backend/ | 仅内部 |
| sms-newapi | sms/newapi/ | 仅内部 |

---

## 跨产品通信

```
Apps Backend (8010)  ──GET  /api/v1/pricing/latest──→  SMS Backend (8020)
                     ←── { version, model_ratio, completion_ratio } ──┘
```

接口极少，耦合极低。两边独立开发、独立部署。
