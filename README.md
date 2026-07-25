# TokenJoy Monorepo

两个独立产品，三个部署形态，一个代码仓库。

| 产品 | 部署形态 | 用户 | 核心职责 |
|------|---------|------|---------|
| **TokenJoy** | Local（私有化）/ SaaS（公有云） | 客户管理员 | Key 管理、预算、组织、审计、Gateway |
| **SMS** | 中心部署 | TJ 运营团队 | 供应商管理、模型目录、合同、定价发布 |

---

## 目录结构

```
mytokenjoy/
├── apps/                    ← 客户侧产品（TokenJoy Local + SaaS）
│   ├── frontend/            ← React SPA（@tokenjoy/frontend）
│   ├── backend/             ← Go 后端（github.com/tokenjoy/backend）
│   ├── newapi/              ← NewAPI Docker 构建 + 脚本
│   ├── dev-mock-llm/        ← 本地模拟 LLM 上游
│   └── docs/                ← TokenJoy 产品文档
│
├── sms/                     ← 内部运营产品
│   ├── frontend/            ← React SPA（@sms/frontend）
│   ├── backend/             ← Go 后端（module: sms/backend）
│   ├── newapi/              ← NewAPI 配置
│   └── docs/                ← SMS 产品文档
│
├── packages/
│   └── contracts/           ← 跨产品共享契约（permission codegen 等）
│
├── scripts/                 ← 开发脚本
│   ├── dev.sh               ← 主调度器（start/reset/test 路由）
│   ├── dev-sms.sh           ← SMS start/reset
│   ├── lib/common.sh        ← 共享路径定义
│   ├── lib/db-reset.sh      ← DB reset 共享函数
│   └── postgres-init/       ← Docker 容器首次 init 脚本
│
├── docker-compose.yml       ← 统一开发基础设施
├── pnpm-workspace.yaml      ← workspace 配置
├── package.json             ← 根命令入口
└── CLAUDE.md                ← AI 编码助手规则
```

---

## 开发环境

### 端口

| 服务 | apps (TokenJoy) | sms | 说明 |
|------|----------------|-----|------|
| Postgres | 5510 | 5510 | 共用一个容器，不同 database |
| Redis | 6310 | 6310 | 共用一个容器，不同 db number |
| Backend | 8010 | 8020 | |
| Frontend | 5173 | 5174 | |
| NewAPI | 3010 | 3020 | |

### 数据库隔离

一个 Postgres 容器，两个用户，六个 database：

| Database | Owner | 产品 |
|----------|-------|------|
| tokenjoy | tokenjoy | apps 主库 |
| newapi | tokenjoy | apps NewAPI |
| logs | tokenjoy | apps 日志 |
| sms | sms | sms 主库 |
| sms_newapi | sms | sms NewAPI |
| sms_logs | sms | sms 日志 |

Redis: apps NewAPI=db0, sms NewAPI=db1, apps backend=db2, sms backend=db3

### 命令

```bash
# 基础设施
pnpm infra               # 启动 docker compose（postgres + redis + 两个 NewAPI）
pnpm infra:down          # 停止

# Apps (TokenJoy)
pnpm start               # 启动 apps（backend + frontend + mock）
pnpm reset               # 重置 apps 库（默认 local 模式）
pnpm reset saas          # 重置 apps 库（SaaS 多租户模式）

# SMS
pnpm start sms           # 启动 sms（backend + frontend）
pnpm reset sms           # 重置 sms 库

# 全部
pnpm start all           # 并行启动 apps + sms
pnpm reset all           # 重置全部

# 测试
pnpm test                # apps 全量测试
pnpm test:sms            # sms 全量测试
pnpm lint                # apps lint
pnpm lint:sms            # sms lint

# 构建
pnpm build               # apps frontend
pnpm build:sms           # sms frontend
```

---

## 核心规则

1. **两个产品，两个边界** — `apps/` 是客户的，`sms/` 是内部的
2. **Local ↔ SaaS** — 同一前后端，运行时 `SUPPORT_SAAS` flag 分流
3. **跨产品通信仅 HTTP API** — 不共享 DB、不共享 domain 代码
4. **共享层放 packages/** — 跨产品共享的契约、类型放这里
5. **制品交付** — 客户只拿 `apps/` 的 image，`sms/` 永不出门
6. **独立演进** — sms 不阻塞客户侧发版，反之亦然
7. **基础设施共用** — 开发环境一份 docker-compose，端口段隔离，database 名隔离

### 跨产品约束

- apps/ 和 sms/ 之间禁止 Go import
- apps/ 和 sms/ 之间禁止 TypeScript import
- 共享类型/契约只能放 packages/contracts/
- 跨产品通信只能通过 HTTP API

### 文档放置

- 根 `docs/`：仓库级/跨产品文档
- `apps/docs/`：TokenJoy 产品文档
- `sms/docs/`：SMS 产品文档

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
