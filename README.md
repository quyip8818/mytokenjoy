# TokenJoy Monorepo

两个独立产品，三个部署形态，一个代码仓库。

| 产品 | 部署形态 | 用户 | 核心职责 |
|------|---------|------|---------|
| **TokenJoy** | Local（私有化）/ SaaS（公有云） | 客户管理员 | Key 管理、预算、组织、审计、Gateway |
| **Web** | SaaS 官网 | 公众访客 | 产品官网、注册入口（iframe 嵌入 App 认证） |
| **SMS** | 中心部署 | TJ 运营团队 | 供应商管理、模型目录、合同、定价发布 |

TokenJoy Local 和 SaaS 共用同一套前后端代码，运行时通过 `SUPPORT_SAAS` flag 分流。

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
├── web/                     ← 产品官网（@tokenjoy/web）
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
| Web (官网) | 5175 | — | |
| NewAPI | 3010 | 3020 | |

### 数据库隔离

一个 Postgres 容器，两个用户，六个 database：

| Database | Owner | 产品 | 用途 |
|----------|-------|------|------|
| tokenjoy | tokenjoy | apps | 主库 |
| newapi | tokenjoy | apps | NewAPI 数据 |
| logs | tokenjoy | apps | NewAPI 日志 |
| sms | sms | sms | 主库 |
| sms_newapi | sms | sms | NewAPI 数据 |
| sms_logs | sms | sms | NewAPI 日志 |

Redis db number: apps NewAPI=0, sms NewAPI=1, apps backend=2, sms backend=3

### 命令

```bash
# 基础设施
pnpm infra               # docker compose up（postgres + redis + 两个 NewAPI）
pnpm infra:down          # 停止

# Apps (TokenJoy)
pnpm start               # backend + frontend + mock
pnpm reset               # 重置 apps 库（默认 local 模式）
pnpm reset saas          # 重置 apps 库（SaaS 多租户模式）
pnpm reset [local|saas] [--empty|--minimal|--full]

# SMS
pnpm start sms           # backend + frontend
pnpm reset sms           # 重置 sms 库

# 全部
pnpm start all           # 并行启动 apps + sms
pnpm reset all           # 重置全部

# 测试
pnpm test                # apps 全量测试
pnpm test:sms            # sms 全量测试
pnpm test:web            # web 官网测试
pnpm test:integration    # apps 后端集成测试
pnpm test:sms:integration
pnpm test:e2e            # apps 前端 E2E
pnpm test:sms:e2e

# 质量
pnpm lint                # apps lint
pnpm lint:sms            # sms lint
pnpm lint:web            # web 官网 lint

# 构建
pnpm build               # apps frontend
pnpm build:sms           # sms frontend
pnpm build:web           # web 官网

# 官网开发
pnpm start:web           # 启动官网 dev server (port 5175)
```

### Web 平台 (apps/web/)

官网是纯展示型轻量 SPA（Tailwind v3 + React 19），不含路由库、状态管理、认证逻辑。

- **认证集成**：登录/注册通过 iframe 嵌入 App 的 `/embed.html` 独立入口（Vite 多入口，不加载主 SPA 的 router/providers）
- **postMessage 协议**：iframe 认证成功后发 `{ type: 'auth:success' }` 通知官网跳转
- **部署域名**：`www.tokenjoy.com`（App 在 `app.tokenjoy.com`，Cookie 共享 `.tokenjoy.com`）
- **联调**：需同时运行 `pnpm start`（App）+ `pnpm start:web`（官网）

### Reset 机制

- SQL `DROP/CREATE DATABASE`，不用 `docker compose down -v`
- 每个 reset 只动自己的 database，不影响对方
- 共享逻辑：`scripts/lib/db-reset.sh`（`reset_apps_databases` / `reset_sms_databases`）

---

## 核心规则

1. **两个产品，两个边界** — `apps/` 是客户的，`sms/` 是内部的
2. **Local ↔ SaaS** — 同一前后端，运行时 `SUPPORT_SAAS` flag 分流
3. **跨产品通信仅 HTTP API** — 不共享 DB、不共享 domain 代码
4. **共享层放 packages/** — 跨产品共享的契约、类型放这里
5. **制品交付** — 客户只拿 `apps/` 的 image，`sms/` 永不出门
6. **独立演进** — sms 不阻塞客户侧发版，反之亦然

### 跨产品约束

- apps/ 和 sms/ 之间禁止 Go import
- apps/ 和 sms/ 之间禁止 TypeScript import
- 共享类型/契约只能放 packages/contracts/
- 跨产品通信只能通过 HTTP API

### 文档放置

| 位置 | 内容 |
|------|------|
| `apps/docs/` | TokenJoy 产品文档（架构、PRD、ADR、plan） |
| `sms/docs/` | SMS 产品文档 |

禁止在 apps/frontend/、apps/backend/、sms/frontend/、sms/backend/ 下新建 .md（README.md 除外）。

---

## 构建产物

| 产物 | 来源 | 交付 |
|------|------|------|
| tokenjoy/frontend | apps/frontend/ | 客户 |
| tokenjoy/backend | apps/backend/ | 客户 |
| tokenjoy/newapi | apps/newapi/ | 客户 |
| tokenjoy/web | web/ | 公开（www.tokenjoy.com） |
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
