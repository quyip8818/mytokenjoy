# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）+ `apps/newapi`（Relay 本地栈）。本地联调：`pnpm start`。

## 文档地图

| 文档                                                   | 读者               | 内容                                                  |
| ------------------------------------------------------ | ------------------ | ----------------------------------------------------- |
| [TokenJoy-PRD.md](./TokenJoy-PRD.md)                   | 产品 / 全员        | 需求与验收标准                                        |
| [Backend.md](./Backend.md)                             | 后端               | 索引：SaaS、运行、Relay、测试                         |
| [Backend-架构.md](./Backend-架构.md)                   | 后端 / 架构        | 分层、请求链、Relay、看板读路径                       |
| [Backend-存储.md](./Backend-存储.md)                   | 后端 / DBA         | 36 表、ER、合并表、ID 约定                            |
| [Backend-预算.md](./Backend-预算.md)                   | 后端 / 计费        | 双轴、Ingest、Rebalance、Overrun                      |
| [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md) | 后端 / 联调        | NewAPI 入账方案 B 现状与剩余缺口                      |
| [权限管理.md](./权限管理.md)                           | 后端 / 前端 / 架构 | **目标架构实现文档**（破坏性替换、WP-1–6、验收 Gate） |
| [Frontend.md](./Frontend.md)                           | 前端 / 前后端      | 前端现状与 **API 契约**（82+11 端点；17 业务页）      |
| [Roadmap.md](./Roadmap.md)                             | 全员               | 未实现能力与计划优化                                  |
| [下一步工作清单.md](./下一步工作清单.md)               | 架构 / 前后端      | 产品模型与 UI 演进架构、发布门禁、Phase 3 路线        |

## 权威来源优先级

1. API 路径与 JSON → [Frontend.md](./Frontend.md) §5 + `apps/frontend/src/api/types/`
2. 后端类型 → `apps/backend/internal/domain/types/`
3. 业务规则 → 各 domain `Service` 实现
4. 差距与计划 → [Roadmap.md](./Roadmap.md)

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
