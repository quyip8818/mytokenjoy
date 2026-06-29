# TokenJoy 文档索引

Monorepo：`apps/frontend`（React）+ `apps/backend`（Go）。本地联调在仓库根目录执行 `pnpm start`。

## 文档地图

| 文档                                           | 读者        | 内容                                                  |
| ---------------------------------------------- | ----------- | ----------------------------------------------------- |
| [TokenJoy-PRD.md](./TokenJoy-PRD.md)           | 产品 / 全员 | 需求、用户故事、数据模型                              |
| [Frontend-API契约.md](./Frontend-API契约.md)   | 前后端      | **81 个 REST 端点**、类型、鉴权、错误约定（权威来源） |
| [Frontend-开发指南.md](./Frontend-开发指南.md) | 前端        | 目录结构、路由、API DI、页面 Hook、测试               |
| [Backend-设计.md](./Backend-设计.md)           | 后端        | 分层、Store、配置、看板用量、维护要点                 |
| [Backend-存储架构.md](./Backend-存储架构.md)   | 后端        | Postgres 表、domain_snapshot、拆分演进与实体关系      |
| [Backend-待实现.md](./Backend-待实现.md)       | 后端 / 产品 | PRD 与实现的差距、推荐实施顺序                        |
| [Backend-test.md](./Backend-test.md)           | 后端        | 测试目录、运行方式、编写规范                          |

## 权威来源优先级

1. API 路径与 JSON 形状 → [Frontend-API契约.md](./Frontend-API契约.md)
2. 前端类型 → `apps/frontend/src/api/types/`
3. 后端类型 → `apps/backend/internal/domain/types/`
4. 业务规则 → 各 domain `Service` 实现

## 常用命令

```bash
pnpm install    # 安装依赖
pnpm start      # 并发 backend :8080 + frontend（Vite 代理 /api）
pnpm verify     # lint + test + build（PR 前）
pnpm test       # 前端 Vitest + 后端 go test
```
