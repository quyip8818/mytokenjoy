# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）+ `apps/newapi`（NewAPI）+ `apps/dev-mock-llm`（本地 ingest 测试上游）；共享契约 `packages/contracts`。本地联调：`pnpm start`（Postgres + Redis + NewAPI + backend + frontend + mock）。

## 权威来源

| 用途 | 文档 |
| --- | --- |
| 工程待办（上线前 fix / 功能 / 门禁） | **[plan.md](./plan.md)** |
| 架构现状 | [Backend.md](./Backend.md) 及子文档、[Backend-结构优化.md](./Backend-结构优化.md)、[Backend-模块化设计.md](./Backend-模块化设计.md)、[Backend-离线任务.md](./Backend-离线任务.md)（**as-built** 离线任务）、[工程收口.md](./工程收口.md)、[Frontend.md](./Frontend.md) |
| 架构目标 / 评审 | [架构终态设计.md](./架构终态设计.md)（**目标态，非 as-built**）、[架构评审-系统与数据模型.md](./架构评审-系统与数据模型.md) |
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
| [Backend-结构优化.md](./Backend-结构优化.md) | 后端 / 架构 | 结构基线与剩余分层债务 |
| [Backend-模块化设计.md](./Backend-模块化设计.md) | 后端 / 架构 | **目标态**：模块地图、`app/` 重组、分阶段 PR 切片 |
| [Backend-v1-Ingest链路优化.md](./Backend-v1-Ingest链路优化.md) | 后端 / 架构 | **v1 Gateway → Ingest** 性能项（G/I/P/R）+ 不牺牲热路径的 Lag 缩窗（§10） |
| [架构评审-系统与数据模型.md](./架构评审-系统与数据模型.md) | 架构 / DBA | 架构债与问题分析（含上线后 migration 建议） |
| [Backend-离线任务.md](./Backend-离线任务.md) | 后端 | **as-built**：Ingest + River 两条线、10 kind、入队与 Worker |
| [Backend-预算.md](./Backend-预算.md) | 后端 / 计费 | 双轴、异步投影、Rebalance、Overrun、Platform Key 执法链（**as-built**） |
| **[预算分配与扣减.md](./预算分配与扣减.md)** | 产品 / 研发 | **权威**：切蛋糕 vs 独立结算；三 scope 产品行为；personal 用尽 → 审批追加 |
| [Backend-存储架构.md](./Backend-存储架构.md) | 后端 / DBA | 双库表、域关系、Store 与 ID 约定（含 `platform_keys.scope`） |
| [Backend-计费模式.md](./Backend-计费模式.md) | 后端 / 计费 | point + lot、钱包、wallet_sync |
| [Backend-币种与入账全链路.md](./Backend-币种与入账全链路.md) | 后端 / 计费 | 币种与入账：现状 / 目标架构 |
| [Backend-Ingest架构.md](./Backend-Ingest架构.md) | 后端 / 联调 | 入账全链路：通信、日志共享、对齐与优化 |
| [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md) | 后端 / 架构 | 业务时钟、双轨 period、护栏 |
| [Backend-测试优化.md](./Backend-测试优化.md) | 后端 / 测试 | coverage + 速度优化 |
| [工程收口.md](./工程收口.md) | 研发 / 架构 | 后端、前端、NewAPI 待收口项（按优先级） |
| [权限管理.md](./权限管理.md) | 后端 / 前端 / 架构 | Identity JWT + PDP |
| [架构终态设计.md](./架构终态设计.md) | 架构 | **目标态**（非 as-built）：Gateway 性能、投影、执法分层 |

### 本地联调 / 手工测试

| 文档 | 读者 | 内容 |
| --- | --- | --- |
| **[本地开发-启动优化.md](./本地开发-启动优化.md)** | 研发 | **SSOT**：命令契约、L0–L2、决策树 |
| **[本地模式-模拟消耗Popup.md](./manual-testing/本地模式-模拟消耗Popup.md)** | 产品 / QA / 研发 | `local-test-model` 全链路 ingest 测试 |

### 归档笔记（非权威 backlog）

| 路径 | 说明 |
| --- | --- |
| [reviews/](./reviews/) | 安全评估等一次性笔记 |
| [todos/](./todos/) | 待定产品问题（如 [budget-wallet-limit.md](./todos/budget-wallet-limit.md)） |

## 契约优先级

1. API 路径与 JSON → [Frontend.md](./Frontend.md) §5 + `apps/frontend/src/api/types/`
2. 后端类型 → `apps/backend/internal/domain/types/`
3. 业务规则 → 各 domain `Service` 实现
4. 工程待办 → [plan.md](./plan.md)
5. 产品差距 → [Roadmap.md](./Roadmap.md)

## Backlog 分工

| 文档 | 写什么 | 不写什么 |
| --- | --- | --- |
| [plan.md](./plan.md) | 上线前 fix、联调门禁、发布验收 | 产品级长期 ❌（见 Roadmap） |
| [工程收口.md](./工程收口.md) | 架构/联调/边界类未完成项 | 日常功能 backlog |
| [Roadmap.md](./Roadmap.md) | PRD vs 实现差距状态 | 具体工程步骤 |
| [reviews/](./reviews/) | 一次性审计笔记 | 活跃 backlog |

## 常用命令

```bash
pnpm install
pnpm start            # 轻量：ensure-infra + backend + frontend + mock（见 本地开发-启动优化.md）
pnpm start:lite       # 无 NewAPI / mock
pnpm docker:reset     # 重初始化 L1a+L1b（alias: pnpm reset）
pnpm infra            # 仅 Docker 基础设施
pnpm infra postgres   # 仅 PG（跑后端测试前）
pnpm verify           # CI：lint + test + build
pnpm verify gate      # 通路冒烟
pnpm verify integration
pnpm test
pnpm test:e2e
pnpm infra attach     # 前台 attach NewAPI（调试）

cd apps/backend && make test-fast    # 仅 tests/pkg/...，无 Postgres
cd apps/backend && make test-unit    # 全量 go test（需 pnpm infra postgres）
```

全链路 ingest 手工测试（`local-test-model` + Popup）：见 [本地模式-模拟消耗Popup.md](./manual-testing/本地模式-模拟消耗Popup.md)。`pnpm start` 为全栈；`pnpm docker:reset` 会自动 bootstrap admin token；channel 失败时再跑 `setup-dev-mock-channel.sh`。

## CI

`.github/workflows/ci.yml`：`verify` job（含 postgres service，执行 `pnpm verify`）。
