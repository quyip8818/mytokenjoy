# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）+ `apps/newapi`（NewAPI）+ `apps/dev-mock-llm`（本地 ingest 测试上游）；共享契约 `packages/contracts`。本地联调：`pnpm start`（Postgres + Redis + NewAPI + backend + frontend + mock）。

---

## 文档地图

### 后端架构

| 文档 | 内容 |
| --- | --- |
| [Backend-架构.md](./Backend-架构.md) | **后端核心入口**：分层、请求链、命名、Gateway、看板、模块化设计、结构约束 |
| [Backend-存储架构.md](./Backend-存储架构.md) | 双库表结构、域关系、Store 映射、ID 与额度术语 |
| [Backend-Ingest架构.md](./Backend-Ingest架构.md) | 入账全链路：通信、日志共享、对齐、同事务 consumed 写入 |
| [Backend-预算.md](./Backend-预算.md) | 双轴预算、分配规则、Rebalance、Overrun、入账累计 |
| [Backend-计费模式.md](./Backend-计费模式.md) | point + lot、币种/PPU、冻结展示、事实/投影边界 |
| [Backend-离线任务.md](./Backend-离线任务.md) | Ingest + River 两条线、10 kind、入队与 Worker |
| [Backend-配置架构.md](./Backend-配置架构.md) | 配置加载、生产契约、空库引导、Clock |
| [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md) | 业务时钟、双轨 period、护栏 |
| [Backend-NewAPI-多租户钥匙代建.md](./Backend-NewAPI-多租户钥匙代建.md) | NewAPI Token 归属与多租户方案 |

### 前端

| 文档 | 内容 |
| --- | --- |
| [Frontend.md](./Frontend.md) | 前端架构、API 契约、路由、联调 |

### 横切能力

| 文档 | 内容 |
| --- | --- |
| [auth-system.md](./auth-system.md) | 认证架构：Session JWT + Refresh + Platform Key |
| [middleware.md](./middleware.md) | Middleware 链、Rate Limiting、Timeout |
| [权限管理.md](./权限管理.md) | Identity JWT + PDP、RBAC、manifest 契约 |
| [Notification.md](./Notification.md) | 多渠道通知系统 |

### 产品与规划

| 文档 | 内容 |
| --- | --- |
| [PRD.md](./PRD.md) | 产品需求（只读权威） |
| [PRD-差距分析.md](./PRD-差距分析.md) | PRD vs 实现差距（按 US 分析） |
| [Roadmap.md](./Roadmap.md) | 产品差距状态简表 |
| [预算分配与扣减.md](./预算分配与扣减.md) | **产品权威**：三 scope、切蛋糕 vs 独立结算 |
| [plan/未实现与优化方向.md](./plan/未实现与优化方向.md) | 各领域待做/可优化项汇总 |

### 架构演进与工程问题

| 文档 | 内容 |
| --- | --- |
| [架构演进.md](./架构演进.md) | 目标态设计 + 长期必改架构问题 |
| [problems.md](./problems.md) | 代码 bug / 技术债清单 |
| [工程收口.md](./工程收口.md) | 联调 / 上线门禁 / 边界未完成项 |

### 本地开发

| 文档 | 内容 |
| --- | --- |
| [本地开发-启动优化.md](./本地开发-启动优化.md) | **SSOT**：命令契约、L0–L2、决策树 |

### 子目录

| 路径 | 说明 |
| --- | --- |
| [adr/](./adr/) | 架构决策记录 |
| [plan/](./plan/) | 工程计划与设计文档 |
| [reviews/](./reviews/) | 安全评估等一次性笔记 |
| [todos/](./todos/) | 待定产品问题 |

---

## 契约优先级

1. API 路径与 JSON → [Frontend.md](./Frontend.md) §5 + `apps/frontend/src/api/types/`
2. 后端类型 → `apps/backend/internal/domain/types/`
3. 业务规则 → 各 domain `Service` 实现
4. 产品差距 → [Roadmap.md](./Roadmap.md)
5. 未实现/优化 → [plan/未实现与优化方向.md](./plan/未实现与优化方向.md)

---

## 常用命令

```bash
pnpm install
pnpm start            # 全栈：ensure-infra + backend + frontend + mock
pnpm start:lite       # 无 NewAPI / mock
pnpm docker:reset     # 重初始化（alias: pnpm reset）
pnpm infra            # 仅 Docker 基础设施
pnpm infra postgres   # 仅 PG（跑后端测试前）
pnpm verify           # CI：lint + test + build
pnpm verify gate      # 通路冒烟
pnpm test
pnpm test:e2e

cd apps/backend && make test-fast    # 仅 tests/pkg/...，无 Postgres
cd apps/backend && make test-unit    # 全量 go test（需 pnpm infra postgres）
```

## CI

`.github/workflows/ci.yml`：`verify` job（含 postgres service，执行 `pnpm verify`）。
