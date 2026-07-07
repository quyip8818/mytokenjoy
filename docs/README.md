# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）+ `apps/newapi`（Relay 本地栈）。本地联调：`pnpm start`。

## 文档地图

| 文档                                                   | 读者               | 内容                                                  |
| ------------------------------------------------------ | ------------------ | ----------------------------------------------------- |
| **[plan.md](./plan.md)**                               | 研发 / 架构        | **工程 backlog 唯一入口**                             |
| [TokenJoy-PRD.md](./TokenJoy-PRD.md)                   | 产品 / 全员        | 需求与验收标准（权威 PRD）                            |
| [Frontend.md](./Frontend.md)                           | 前端 / 前后端      | 前端架构、**API 契约**（82+11 端点；17 业务页）       |
| [Backend.md](./Backend.md)                             | 后端               | 索引：SaaS、运行、Relay、Keys 约束、测试              |
| [Backend-架构.md](./Backend-架构.md)                   | 后端 / 架构        | 分层、请求链、Relay、看板读路径                       |
| [Backend-存储.md](./Backend-存储.md)                   | 后端 / DBA         | 36 表、ER、合并表、ID 约定                            |
| [Backend-预算.md](./Backend-预算.md)                   | 后端 / 计费        | 双轴、Ingest、Rebalance、Overrun                      |
| [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md) | 后端 / 联调        | NewAPI 入账方案 B 架构现状                            |
| [权限管理.md](./权限管理.md)                           | 后端 / 前端 / 架构 | Identity JWT + PDP 实现与验收 Gate                    |
| [Roadmap.md](./Roadmap.md)                             | 全员               | 产品差距（未实现能力）                                |

### 归档与历史

| 文档 | 说明 |
| --- | --- |
| [archive/](./archive/) | 已合并计划的全文（前端架构、发布清单、Keys 规格、MSW 等） |
| [PRD.md](./PRD.md) | **历史副本**，权威 PRD 见 [TokenJoy-PRD.md](./TokenJoy-PRD.md) |
| [MSW去除与后端对齐计划.md](./MSW去除与后端对齐计划.md) | 已完成，指向 archive + plan |
| [前端架构优化与模块化建议.md](./前端架构优化与模块化建议.md) | 已合并，指向 Frontend + plan |
| [下一步工作清单.md](./下一步工作清单.md) | 已合并，指向 Frontend + plan |
| [清理兼容与死代码-下一步.md](./清理兼容与死代码-下一步.md) | 已合并，指向 Backend + plan |

## 权威来源优先级

1. API 路径与 JSON → [Frontend.md](./Frontend.md) §5 + `apps/frontend/src/api/types/`
2. 后端类型 → `apps/backend/internal/domain/types/`
3. 业务规则 → 各 domain `Service` 实现
4. 工程待办 → [plan.md](./plan.md)
5. 产品差距 → [Roadmap.md](./Roadmap.md)

## 常用命令

```bash
pnpm install          # 安装依赖
pnpm start            # Postgres + backend :8080 + frontend :5173
pnpm verify           # lint + test + build + backend build:check（PR 前）
pnpm test             # 前端 Vitest + 后端 go test
pnpm test:integration # 后端 Postgres 集成测试
pnpm test:e2e         # 前端 Playwright E2E
pnpm start:relay      # 完整 NewAPI 栈（Postgres + Redis + new-api）
```

## CI

`.github/workflows/ci.yml` 三 job 并行：`verify`、`frontend-e2e`、`backend-integration`。
