# ADR: 多版本单代码库策略

## 决策

TokenJoy 包含两个独立产品，共三个部署形态，放在同一 monorepo：

| 产品 | 部署形态 | 用户 | 核心职责 |
|------|---------|------|---------|
| **TokenJoy** | Local（私有化） | 客户管理员 | Key 管理、预算、组织、审计、自定义模型、Channel |
| **TokenJoy** | SaaS（公有云） | SaaS 客户 | Local 功能子集，无自定义模型/Channel |
| **SMS** | 中心部署 | TJ 运营团队 | 供应商管理、模型目录、合同、采购订单、定价发布、实例管理、计费 |

---

## 仓库结构

```
mytokenjoy/
├── apps/                    ← 客户侧产品（Local + SaaS）
│   ├── frontend/            ← React SPA（Local + SaaS 共用）
│   ├── backend/             ← Go 后端（Local + SaaS，单 binary）
│   ├── newapi/              ← NewAPI 构建（local variant）
│   ├── dev-mock-llm/
│   └── docs/               ← TokenJoy 产品文档
│
├── sms/                     ← 内部运营产品（原 tokenjoy-sms 仓库整体迁入）
│   ├── frontend/            ← 运营管理 UI（独立 Vite app，@sms/frontend）
│   ├── backend/             ← Go 后端（独立 binary，module: sms/backend）
│   ├── newapi/              ← NewAPI 构建（sms variant）
│   └── docs/               ← SMS 产品文档
│
├── packages/                ← 跨产品共享
│   └── contracts/           ← 共享契约（permission、notification 等）+ codegen
│
├── docker-compose.yml       ← 统一基础设施（单个 Postgres + Redis）
├── docs/                    ← 仓库级/跨产品文档
└── scripts/
```

---

## 开发环境：同机并行方案

### 设计原则

1. **共用基础设施容器** — 一个 Postgres 实例、一个 Redis 实例，省资源
2. **端口段隔离** — apps 用 xx10 段，sms 用 xx20 段，不占默认端口，不互相打架
3. **数据库名隔离** — 同一 Postgres 里不同 database，各自独立互不可见
4. **独立 reset** — `pnpm reset` 只 drop/recreate apps 的库，`pnpm reset:sms` 只动 sms 的库

### 端口总表

| 服务 | apps (TokenJoy) | sms | 说明 |
|------|----------------|-----|------|
| **Postgres** | 5510 (共用) | 5510 (共用) | 一个容器，不同 database |
| **Redis** | 6310 (共用) | 6310 (共用) | 一个容器，不同 db number |
| **Backend (Go)** | 8010 | 8020 | |
| **Frontend (Vite)** | 5173 | 5174 | |
| **NewAPI (apps)** | 3010 | — | |
| **NewAPI (sms)** | — | 3020 | |

> 两个 NewAPI 实例各自连不同 database，互不干扰。

### 数据库隔离

一个 Postgres 容器内共 6 个 database（含默认库 `tokenjoy`，init 脚本额外创建 5 个）：

| Database 名 | 所属产品 | 用途 | 连接用户 |
|-------------|---------|------|---------|
| `tokenjoy` | apps | apps backend 主库 | tokenjoy |
| `newapi` | apps | apps NewAPI 数据 | tokenjoy |
| `logs` | apps | apps NewAPI 日志 | tokenjoy |
| `sms` | sms | sms backend 主库 | sms |
| `sms_newapi` | sms | sms NewAPI 数据 | sms |
| `sms_logs` | sms | sms NewAPI 日志 | sms |

> apps 侧三个库用 Postgres 用户 `tokenjoy`，sms 侧三个库用 Postgres 用户 `sms`，开发/生产通用，天然隔离。

**为什么 sms 的 NewAPI 库叫 `sms_newapi` / `sms_logs` 而不是 `newapi` / `logs`？**
— 因为 apps 已经占用了 `newapi` 和 `logs` 这两个名字。同一 Postgres 里 database 名必须唯一。加前缀彻底消除歧义。

### Redis 隔离

共用一个 Redis 实例，通过 db number 隔离：

| 产品 | Redis DB | 用途 |
|------|----------|------|
| apps NewAPI | db 0 | NewAPI session/cache |
| sms NewAPI | db 1 | SMS NewAPI session/cache |
| apps backend | db 2 | 如果将来需要 |
| sms backend | db 3 | 如果将来需要 |

Redis 连接字符串加 `/<db>` 后缀即可选择 db number：`redis://127.0.0.1:6310/1`

---

## 统一 docker-compose.yml（根目录）

```yaml
# mytokenjoy/docker-compose.yml — 开发环境统一基础设施
services:
  postgres:
    image: postgres:17-alpine
    command: ['postgres', '-c', 'max_locks_per_transaction=1024', '-c', 'max_connections=300']
    environment:
      POSTGRES_USER: tokenjoy
      POSTGRES_PASSWORD: tokenjoy
      POSTGRES_DB: tokenjoy
    ports:
      - '5510:5432'
    volumes:
      - dev_pg:/var/lib/postgresql/data
      - ./scripts/postgres-init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U tokenjoy -d tokenjoy']
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - '6310:6379'
    healthcheck:
      test: ['CMD', 'redis-cli', 'ping']
      interval: 5s
      timeout: 5s
      retries: 5

  newapi-apps:
    build:
      context: ./apps/newapi
      dockerfile: Dockerfile
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    ports:
      - '127.0.0.1:3010:3000'
    extra_hosts:
      - 'host.docker.internal:host-gateway'
    environment:
      SQL_DSN: postgresql://tokenjoy:tokenjoy@postgres:5432/newapi?sslmode=disable
      LOG_SQL_DSN: postgresql://tokenjoy:tokenjoy@postgres:5432/logs?sslmode=disable&search_path=newapi
      REDIS_CONN_STRING: redis://redis:6379/0
      SESSION_SECRET: tokenjoy-dev-session-secret
      SYNC_FREQUENCY: 60
      MANAGEMENT_WEBHOOK_URL: http://host.docker.internal:8010/api/internal/webhooks/newapi-log
      MANAGEMENT_WEBHOOK_SECRET: tokenjoy-webhook-secret

  newapi-sms:
    image: calciumion/new-api:latest
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    ports:
      - '127.0.0.1:3020:3000'
    extra_hosts:
      - 'host.docker.internal:host-gateway'
    environment:
      SQL_DSN: postgresql://sms:sms@postgres:5432/sms_newapi?sslmode=disable
      LOG_SQL_DSN: postgresql://sms:sms@postgres:5432/sms_logs?sslmode=disable&search_path=newapi
      REDIS_CONN_STRING: redis://redis:6379/1
      SESSION_SECRET: sms-dev-session-secret
      SYNC_FREQUENCY: 60

volumes:
  dev_pg:
```

### scripts/postgres-init/01-create-all-dbs.sh

```bash
#!/bin/bash
set -euo pipefail

# 创建 sms 用户
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-EOSQL
    CREATE USER sms WITH PASSWORD 'sms';
EOSQL

# 创建所有 database（apps 侧 owner=tokenjoy，sms 侧 owner=sms）
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-EOSQL
    -- apps 侧
    CREATE DATABASE newapi;
    CREATE DATABASE logs;

    -- sms 侧
    CREATE DATABASE sms OWNER sms;
    CREATE DATABASE sms_newapi OWNER sms;
    CREATE DATABASE sms_logs OWNER sms;
EOSQL

# 为各 database 安装所需 extensions
for db in tokenjoy newapi logs sms sms_newapi sms_logs; do
  psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "$db" <<-EOSQL
      CREATE EXTENSION IF NOT EXISTS ltree;
EOSQL
done

# 创建日志库的 schema
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "logs" <<-EOSQL
    CREATE SCHEMA IF NOT EXISTS newapi;
EOSQL

psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "sms_logs" <<-EOSQL
    CREATE SCHEMA IF NOT EXISTS newapi AUTHORIZATION sms;
EOSQL
```

---

## 环境变量配置

### apps/backend/.env.development

```env
DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5510/tokenjoy?sslmode=disable
PORT=8010
CORS_ORIGINS=http://localhost:5173
NEW_API_BASE_URL=http://127.0.0.1:3010
LOG_DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5510/logs?sslmode=disable
REDIS_URL=redis://127.0.0.1:6310/2
# ... 其余不变
```

### sms/backend/.env.development

```env
DATABASE_URL=postgres://sms:sms@127.0.0.1:5510/sms?sslmode=disable
PORT=8020
CORS_ORIGINS=http://localhost:5174
NEWAPI_BASE_URL=http://localhost:3020
REDIS_URL=redis://127.0.0.1:6310/3
JWT_SECRET=sms-dev-secret
# ... 其余不变
```

### sms/frontend/vite.config.ts

```ts
export default defineConfig({
  server: {
    port: 5174,
    proxy: {
      '/api': { target: 'http://127.0.0.1:8020' },
    },
  },
})
```

---

## 统一 pnpm 命令

### pnpm-workspace.yaml

```yaml
packages:
  - 'apps/*'
  - 'sms/*'
  - 'packages/*'
allowBuilds:
  esbuild: true
  msw: true
```

### 根 package.json scripts

```jsonc
{
  "scripts": {
    // ─── 基础设施 ───
    "infra": "docker compose up -d --wait",
    "infra:down": "docker compose down",

    // ─── apps（客户侧） ───
    "start": "bash scripts/dev.sh start",
    "reset": "bash scripts/dev.sh reset",

    // ─── sms（内部运营） ───
    "start:sms": "bash scripts/dev-sms.sh start",
    "reset:sms": "bash scripts/dev-sms.sh reset",

    // ─── 全部 ───
    "start:all": "concurrently -n apps,sms -c blue,cyan \"pnpm start\" \"pnpm start:sms\"",
    "reset:all": "bash scripts/dev.sh reset && bash scripts/dev-sms.sh reset",

    // ─── 通用 ───
    "build": "pnpm -F @tokenjoy/frontend build",
    "build:sms": "pnpm -F @sms/frontend build",
    "lint": "pnpm -r --parallel lint",
    "test": "bash scripts/dev.sh test",
    "test:sms": "pnpm -F @sms/frontend test"
  }
}
```

### scripts/dev-sms.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

cmd_start() {
  cleanup() { pkill -P $$ 2>/dev/null || true; sleep 0.5; pkill -9 -P $$ 2>/dev/null || true; }
  trap cleanup EXIT INT TERM

  # 确保共享基础设施已起
  docker compose -f "${ROOT}/docker-compose.yml" up postgres redis newapi-sms -d --wait

  concurrently --kill-others-on-fail --kill-signal SIGTERM -n sms-be,sms-fe -c blue,green \
    "pnpm -F @sms/backend start" \
    "pnpm -F @sms/frontend dev"
}

cmd_reset() {
  # 只 drop/recreate sms 相关的 database，不影响 apps
  docker compose -f "${ROOT}/docker-compose.yml" up postgres -d --wait

  psql "postgres://tokenjoy:tokenjoy@127.0.0.1:5510/postgres" <<-EOSQL
    DROP DATABASE IF EXISTS sms;
    DROP DATABASE IF EXISTS sms_newapi;
    DROP DATABASE IF EXISTS sms_logs;
    CREATE DATABASE sms OWNER sms;
    CREATE DATABASE sms_newapi OWNER sms;
    CREATE DATABASE sms_logs OWNER sms;
EOSQL

  # 重建 extensions 和 schema
  for db in sms sms_newapi sms_logs; do
    psql "postgres://tokenjoy:tokenjoy@127.0.0.1:5510/$db" <<-EOSQL
      CREATE EXTENSION IF NOT EXISTS ltree;
EOSQL
  done

  psql "postgres://tokenjoy:tokenjoy@127.0.0.1:5510/sms_logs" <<-EOSQL
    CREATE SCHEMA IF NOT EXISTS newapi AUTHORIZATION sms;
EOSQL

  pnpm -F @sms/backend seed
  echo -e "\nSMS reset complete. Next: pnpm start:sms"
}

case "${1:-}" in
  start) cmd_start ;;
  reset) cmd_reset ;;
  *) echo "usage: scripts/dev-sms.sh <start|reset>" >&2; exit 1 ;;
esac
```

### scripts/dev.sh reset（调整思路）

apps 的 reset 同理只 drop/recreate apps 相关的 3 个 database（`tokenjoy`、`newapi`、`logs`），不动 sms 的。不能再用 `docker compose down -v` 粗暴删 volume。

---

## 冲突排查清单

### 运行时隔离（已确认无冲突）

| 维度 | 状态 | 说明 |
|------|------|------|
| 端口 | ✅ 无冲突 | apps xx10 段、sms xx20 段，完全分离 |
| Database 名 | ✅ 无冲突 | sms 侧用 `sms`/`sms_newapi`/`sms_logs` 前缀 |
| Redis | ✅ 无冲突 | 不同 db number（0/1/2/3） |
| Go module | ✅ 无冲突 | `apps/backend` → module 路径与 `sms/backend` 完全独立 |
| npm 包名 | ✅ 无冲突 | `@tokenjoy/frontend` vs `@sms/frontend` |
| Docker 容器 | ✅ 无冲突 | 统一一份 compose，service name 明确区分 |
| Vite HMR websocket | ✅ 无冲突 | 不同端口（5173 vs 5174）各自独立 ws 连接 |
| .env 文件 | ✅ 无冲突 | 各在自己目录下（apps/backend/.env.development vs sms/backend/.env.development） |
| Go air（热重载）| ✅ 无冲突 | 各自 .air.toml 在自己目录，watch 自己的文件 |
| pnpm filter | ✅ 无冲突 | `-F @tokenjoy/*` vs `-F @sms/*`，名字不重叠 |
| DB 用户 | ✅ 无冲突 | apps 侧用 `tokenjoy`，sms 侧用 `sms`，owner 隔离互不可见 |

### pnpm 命令隔离

**无冲突。** 设计规则：

- 根 package.json 裸命令（`start`、`reset`、`build`、`test`）默认操作 apps（TokenJoy 客户侧）
- sms 的命令一律带 `:sms` 后缀（`start:sms`、`reset:sms`、`build:sms`、`test:sms`）
- 全局操作带 `:all` 后缀（`start:all`、`reset:all`）
- `lint` 是唯一一个同时跑两边的（`pnpm -r --parallel lint`），因为代码质量不分产品
- pnpm workspace filter：apps 用 `@tokenjoy/*`，sms 用 `@sms/*`，包名不重叠

### docs 目录隔离

**就近放置，文档跟着产品走。**

```
mytokenjoy/
├── apps/
│   ├── docs/              ← TokenJoy 产品文档（plan/、adr/、todos/ 等）
│   ├── frontend/
│   ├── backend/
│   └── ...
├── sms/
│   ├── docs/              ← SMS 产品文档
│   ├── frontend/
│   ├── backend/
│   └── ...
├── docs/                  ← 仓库级/跨产品文档（本 ADR、整体 Roadmap）
└── ...
```

**规则：**
- apps 专属文档放 `apps/docs/`（现有 `docs/` 下的内容迁入）
- sms 专属文档放 `sms/docs/`
- 仅跨产品文档留在根 `docs/`（如本策略文档、整体 Roadmap）
- 好处：产品边界完整，未来 sms 若拆仓直接搬走整个 `sms/` 目录即可

### .kiro/ 配置隔离

**skills 完全重复，不冲突；steering 需要合并更新。**

分析：

| 文件 | 状态 | 处理方式 |
|------|------|---------|
| `.kiro/skills/*` | ✅ 两边一模一样 | 保留一份即可（都是通用技能如 golang-pro、react-patterns） |
| `.kiro/steering/chinese.md` | ✅ 相同 | 保留一份 |
| `.kiro/steering/migration.md` | ✅ 相同 | 保留一份 |
| `.kiro/steering/ponytail.md` | ✅ 相同 | 保留一份 |
| `.kiro/steering/project-structure.md` | ⚠️ 需合并 | 扩展内容以覆盖 sms 目录规则 |

**project-structure.md 合并方案：**

在现有 steering 规则中追加 sms 段：

```markdown
## SMS (sms/) 文件放置

### 后端 (sms/backend/)
- 结构同 apps/backend：cmd/ 只放入口，internal/ 放业务
- Go module: sms/backend（独立，不引用 apps/backend）

### 前端 (sms/frontend/)
- 结构同 apps/frontend：features/、components/ui/、routes/ 等
- 包名 @sms/frontend
- 禁止与 apps/frontend 互相 import

### 跨产品约束
- apps/ 和 sms/ 之间禁止 Go import
- apps/ 和 sms/ 之间禁止 TypeScript import
- 共享类型/契约只能放 packages/contracts/
- 跨产品通信只能通过 HTTP API
```

**迁移时动作：**
1. 删除 sms 里的 `.kiro/` 目录（不带入，用根目录的统一配置）
2. 更新根 `.kiro/steering/project-structure.md` 追加 sms 段

---

## 迁移步骤

1. **复制 sms 代码** — 将 `tokenjoy-sms/apps/{frontend,backend,newapi}` 复制到 `mytokenjoy/sms/`
2. **创建根 docker-compose.yml** — 替代原来 `apps/newapi/docker-compose.yml` 和 sms 各自的 compose
3. **创建 scripts/postgres-init/01-create-all-dbs.sh** — 统一初始化所有 database
4. **调整端口** — apps backend 8080→8010、apps newapi 3000→3010；sms backend 8081→8020、sms newapi 3000→3020、sms frontend 5173→5174
5. **调整 sms database 名** — `newapi`→`sms_newapi`、`logs`→`sms_logs`
6. **调整 Redis 连接** — 端口改 6310，各加 db number
7. **pnpm-workspace.yaml** — 添加 `'sms/*'`
8. **添加 scripts/dev-sms.sh** — 如上
9. **改造 scripts/dev/reset.sh** — 从 `down -v` 改为 SQL drop/recreate 指定 database
10. **pnpm install** — 一次安装
11. **验证** — `pnpm infra && pnpm reset && pnpm start` 确认 apps 正常；`pnpm reset:sms && pnpm start:sms` 确认 sms 正常；`pnpm start:all` 确认并行无冲突

### 不需要做的事

- 不合并 Go module — 两个后端各自独立
- 不统一前端依赖版本 — pnpm workspace hoisting 自动处理
- 不做 migration — 项目未上线
- 不合并 schema — 两个产品领域完全不同，共用表零收益

---

## 核心规则

1. **两个产品，两个边界** — `apps/` 是客户的，`sms/` 是内部的
2. **Local ↔ SaaS** — 同一前后端，运行时 `SupportSaas` flag 分流
3. **跨产品通信仅 HTTP API** — 不共享 DB、不共享 domain 实现代码
4. **共享层放 packages/** — 任何需要跨产品共享的契约、类型放这里
5. **制品交付** — 客户只拿 `apps/` 的 image，`sms/` 永不出门
6. **独立演进** — sms 不阻塞客户侧发版，反之亦然
7. **基础设施共用** — 开发环境一份 docker-compose，端口段隔离，database 名隔离

---

## 跨产品通信

```
Apps Backend (8010)  ──GET  /api/v1/pricing/latest──→  SMS Backend (8020)
                     ←── { version, model_ratio, completion_ratio } ──┘

Apps Backend (8010)  ──POST /api/v1/instances/register──→  SMS Backend (8020)
                     ←── { instance_id, ... } ────────────┘
```

接口极少，耦合极低。两边独立开发、独立部署。

---

## 构建产物

| 产物 | 来源 | 交付 |
|------|------|------|
| `tokenjoy/frontend` | `apps/frontend/` | 客户 |
| `tokenjoy/backend` | `apps/backend/` | 客户 |
| `tokenjoy/newapi-local` | `apps/newapi/` | 客户 |
| `tokenjoy/sms-ui` | `sms/frontend/` | 仅内部 |
| `tokenjoy/sms-backend` | `sms/backend/` | 仅内部 |
| `tokenjoy/newapi-sms` | `sms/newapi/` | 仅内部 |

---

## 决策理由

| 决策 | 为什么 |
|------|--------|
| SMS 独立后端 | 领域完全不同（SRM vs Key 管理），零共享 |
| 同一 monorepo | 原子提交、共享 CI、共享 packages |
| `sms/` 而非 `platform/` | 命名与包名一致（@sms/frontend, sms/backend） |
| Local/SaaS 不拆 | UI 重叠 90%，一个 flag 足够 |
| 共用 Postgres/Redis | 省资源，database owner 隔离（tokenjoy vs sms 用户）+ 端口段分离 |
| 端口段分离 | 不占默认端口，避免和其他项目冲突 |
| SQL drop/create 而非 volume 删除 | 允许独立 reset 各自产品而不影响对方 |
