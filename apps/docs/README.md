# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）+ `apps/newapi`（NewAPI）+ `apps/dev-mock-llm`（本地 ingest 测试上游）；共享契约 `packages/contracts`。本地联调：`pnpm start`（Postgres + Redis + NewAPI + backend + frontend + mock）。

---

## 文档地图

### 后端架构

| 文档                                                                   | 内容                                                                                      |
| ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| [Backend-架构.md](./Backend-架构.md)                                   | **后端核心入口**：分层、请求链、命名、Gateway、看板、模块化、NewAPI 集成、Handler 开发约定 |
| [Backend-存储架构.md](./Backend-存储架构.md)                           | 双库表结构、域关系、Store 映射、ID 与额度术语                                             |
| [Backend-Ingest架构.md](./Backend-Ingest架构.md)                       | 入账全链路：通信、日志共享、对齐、同事务 consumed 写入                                    |
| [Backend-预算.md](./Backend-预算.md)                                   | 双轴预算、分配规则、Rebalance、Overrun、入账累计、开发者扩展指南                          |
| [Backend-计费模式.md](./Backend-计费模式.md)                           | point + lot、币种/PPU、冻结展示、事实/投影边界                                            |
| [Backend-离线任务.md](./Backend-离线任务.md)                           | Ingest + River 两条线、10 kind、入队与 Worker                                             |
| [Backend-配置架构.md](./Backend-配置架构.md)                           | 配置加载、生产契约、空库引导、Clock                                                       |
| [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)               | 业务时钟、双轨 period、护栏                                                               |
| [Backend-NewAPI-多租户钥匙代建.md](./Backend-NewAPI-多租户钥匙代建.md) | NewAPI Token 归属与多租户方案                                                             |

### 前端

| 文档                         | 内容                                            |
| ---------------------------- | ----------------------------------------------- |
| [Frontend.md](./Frontend.md) | 前端架构、API 契约、路由、联调、页面体系设计规范 |

### 横切能力

| 文档                                                       | 内容                                                         |
| ---------------------------------------------------------- | ------------------------------------------------------------ |
| [auth-system.md](./auth-system.md)                         | 认证架构、成员状态机、邀请注册流程、Member/User 数据边界     |
| [middleware.md](./middleware.md)                            | Middleware 链、Rate Limiting、Timeout                        |
| [权限管理.md](./权限管理.md)                               | Identity JWT + PDP、RBAC、manifest 契约                      |
| [platform-permission-isolation.md](./platform-permission-isolation.md) | Platform 权限纵深防御：Session/Router/Middleware 三层隔离 |
| [Notification.md](./Notification.md)                       | 多渠道通知系统、数据模型规范、已实现通知事件、UI 规范        |
| [approval-system.md](./approval-system.md)                 | 统一审批引擎：Engine + Handler 模式、4 种审批类型            |
| [命名规范与一致性治理.md](./命名规范与一致性治理.md)       | 全链路命名规则、领域映射表                                   |

### 产品与规划

| 文档                                                   | 内容                          |
| ------------------------------------------------------ | ----------------------------- |
| [PRD.md](./PRD.md)                                     | 产品需求（只读权威）          |
| [PRD-差距分析.md](./PRD-差距分析.md)                   | PRD vs 实现差距（按 US 分析） |
| [Roadmap.md](./Roadmap.md)                             | 产品差距状态简表              |
| [plan/未实现与优化方向.md](./plan/未实现与优化方向.md) | 各领域待做/可优化项汇总       |

### 工程问题

| 文档                                             | 内容                              |
| ------------------------------------------------ | --------------------------------- |
| [problems.md](./problems.md)                     | 代码 bug / 技术债清单             |
| [架构优化建议.md](./架构优化建议.md)             | 并发一致性、Gateway 语义、通知降级等架构改进建议 |
| [安全审查.md](./安全审查.md)                     | 安全风险清单（含已核实无问题项）  |
| [未完成功能清单.md](./未完成功能清单.md)         | 全仓库未完成项汇总（按优先级/业务域） |

### 本地开发

| 文档                                           | 内容                                      |
| ---------------------------------------------- | ----------------------------------------- |
| [本地开发-启动优化.md](./本地开发-启动优化.md) | **SSOT**：命令契约、端口表、L0–L2、决策树 |

### 子目录

| 路径             | 说明                 |
| ---------------- | -------------------- |
| [plan/](./plan/) | 工程计划与待实现设计 |

---

## 契约优先级

1. API 路径与 JSON → [Frontend.md](./Frontend.md) §5 + `apps/frontend/src/api/types/`
2. 后端类型 → `apps/backend/internal/domain/types/`
3. 业务规则 → 各 domain `Service` 实现
4. 预算扩展 → [Backend-预算.md](./Backend-预算.md) §14
5. 产品差距 → [Roadmap.md](./Roadmap.md)
6. 未实现/优化 → [plan/未实现与优化方向.md](./plan/未实现与优化方向.md)

---

## 常用命令

命令速查表（含端口、reset 详解、就绪层级）以 [本地开发-启动优化.md](./本地开发-启动优化.md) 为 **SSOT**，此处仅列最常用的几条：

```bash
pnpm install

pnpm start            # 启动 apps，默认 local 模式
pnpm start:saas       # 启动 apps saas 模式
pnpm start:all        # 并行启动 local + saas
pnpm reset            # 重置 apps 库（默认 local，可传 local/saas）
pnpm infra            # 启动两套 Docker 基础设施（saas + local）

pnpm test             # apps 全量测试（frontend + backend）
pnpm test:integration # apps 后端集成测试
pnpm test:e2e         # apps 前端 E2E
pnpm lint             # apps + web + sms 全量 lint

# sms 独立命令见 scripts/dev-sms.sh：bash scripts/dev-sms.sh start / reset
```

**没有** `pnpm verify`、`pnpm verify gate`、`pnpm lint:sms` 这些命令，`package.json` 未定义。
