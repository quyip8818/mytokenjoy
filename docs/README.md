# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）+ `apps/newapi`（Relay 本地栈）。本地联调在仓库根目录执行 `pnpm start`。

## 文档地图

| 文档                                                               | 读者               | 内容                                                   |
| ------------------------------------------------------------------ | ------------------ | ------------------------------------------------------ |
| [TokenJoy-PRD.md](./TokenJoy-PRD.md)                               | 产品 / 全员        | 产品概述、功能清单、数据模型                           |
| [Backend-SaaS多租户架构.md](./Backend-SaaS多租户架构.md)           | 产品 / 后端 / 架构 | SaaS 多企业（Company / 成员 / 企业钱包）               |
| [NewAPI-SaaS多企业配置.md](./NewAPI-SaaS多企业配置.md)             | 运维 / 后端        | NewAPI 单集群多企业隔离与部署配置                      |
| [Frontend-API契约.md](./Frontend-API契约.md)                       | 前后端             | **82** 个企业面 REST 端点 + **11** 个 SaaS 端点        |
| [Frontend-开发指南.md](./Frontend-开发指南.md)                     | 前端               | 目录结构、路由、API DI、页面 Hook、测试                |
| [Frontend-本地联调与演示数据.md](./Frontend-本地联调与演示数据.md) | 前端               | 真实 API 联调、seed 演示数据、故障排查                 |
| [Backend-设计.md](./Backend-设计.md)                               | 后端               | 分层、Store、配置、看板用量                            |
| [Backend-存储架构.md](./Backend-存储架构.md)                       | 后端               | Postgres 44 张表、实体关系                             |
| [Backend-存储实体优化.md](./Backend-存储实体优化.md)               | 架构               | 实体收敛方向与优先级                                   |
| [Backend-预算运作.md](./Backend-预算运作.md)                       | 后端 / 架构        | 预算域、Ingest、Rebalance、企业钱包                    |
| [Backend-消耗数据SSOT对齐方案.md](./Backend-消耗数据SSOT对齐方案.md) | 架构 / 后端        | 消耗数据单一事实来源终极方案（ledger + 投影）          |
| [Backend-命名规范.md](./Backend-命名规范.md)                       | 全员               | 跨系统字段与术语权威命名（含 `newapi_wallet_user_id`） |
| [Backend-test.md](./Backend-test.md)                               | 后端               | 测试目录、运行方式、编写规范                           |

## 权威来源优先级

1. API 路径与 JSON 形状 → [Frontend-API契约.md](./Frontend-API契约.md)
2. 前端类型 → `apps/frontend/src/api/types/`
3. 后端类型 → `apps/backend/internal/domain/types/`
4. 业务规则 → 各 domain `Service` 实现

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

`.github/workflows/ci.yml` 三 job 并行：`verify`（lint + test + build）、`frontend-e2e`、`backend-integration`。
