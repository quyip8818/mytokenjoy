# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）+ `apps/newapi`（NewAPI）；共享契约 `packages/contracts`。本地联调：`pnpm start`。

## 权威来源

| 用途 | 文档 |
| --- | --- |
| 工程待办（上线前 fix / 功能 / 门禁） | **[plan.md](./plan.md)** |
| 架构现状 | [Backend.md](./Backend.md) 及子文档、[工程收口.md](./工程收口.md)、[Frontend.md](./Frontend.md) |
| 产品差距 | [Roadmap.md](./Roadmap.md)、[PRD-差距分析.md](./PRD-差距分析.md) |
| 产品需求（只读权威） | [PRD.md](./PRD.md) |

## 文档地图

| 文档 | 读者 | 内容 |
| --- | --- | --- |
| **[plan.md](./plan.md)** | 研发 / 架构 | **上线前 backlog 唯一入口** |
| [PRD.md](./PRD.md) | 产品 / 全员 | 需求与验收标准 |
| [PRD-差距分析.md](./PRD-差距分析.md) | 产品 / 研发 | PRD vs 实现差距 |
| [Roadmap.md](./Roadmap.md) | 全员 | 产品差距简表 |
| [Frontend.md](./Frontend.md) | 前端 / 前后端 | 前端架构、API 契约 |
| [Backend.md](./Backend.md) | 后端 | 索引：SaaS、运行、Gateway、Keys、Seed、测试 |
| [Backend-配置架构.md](./Backend-配置架构.md) | 后端 / 运维 | 配置、生产契约、空库引导、Clock |
| [Backend-架构.md](./Backend-架构.md) | 后端 / 架构 | 分层、请求链、命名约定、Gateway、看板读路径 |
| [Backend-存储架构.md](./Backend-存储架构.md) | 后端 / DBA | 双库表、域关系、Store 与 ID 约定 |
| [Backend-计费模式.md](./Backend-计费模式.md) | 后端 / 计费 | point + lot、钱包、wallet_sync |
| [Backend-预算.md](./Backend-预算.md) | 后端 / 计费 | 双轴、Ingest、Rebalance、Overrun |
| [Backend-Ingest架构.md](./Backend-Ingest架构.md) | 后端 / 联调 | 入账全链路：通信、日志共享、对齐与优化 |
| [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md) | 后端 / 架构 | 业务时钟、双轨 period、护栏 |
| [工程收口.md](./工程收口.md) | 研发 / 架构 | 后端、前端、NewAPI 待收口项（按优先级） |
| [权限管理.md](./权限管理.md) | 后端 / 前端 / 架构 | Identity JWT + PDP |

### 归档笔记（非权威 backlog）

| 路径 | 说明 |
| --- | --- |
| [reviews/](./reviews/) | 安全评估等一次性笔记 |

## 契约优先级

1. API 路径与 JSON → [Frontend.md](./Frontend.md) §5 + `apps/frontend/src/api/types/`
2. 后端类型 → `apps/backend/internal/domain/types/`
3. 业务规则 → 各 domain `Service` 实现
4. 工程待办 → [plan.md](./plan.md)
5. 产品差距 → [Roadmap.md](./Roadmap.md)

## 常用命令

```bash
pnpm install          # 安装依赖
pnpm start            # Postgres + backend :8080 + frontend :5173
pnpm start:postgres   # 仅起 PostgreSQL（跑后端测试前必须）
pnpm verify           # lint + test + build + backend build:check（PR 前）
pnpm test             # 前端 Vitest + 后端 go test（需 PostgreSQL）
pnpm test:e2e         # 前端 Playwright E2E
pnpm start:newapi      # 完整 NewAPI 栈（Postgres + Redis + new-api）
```

## CI

`.github/workflows/ci.yml`：`verify` job（含 postgres service，执行 `pnpm verify`）。
