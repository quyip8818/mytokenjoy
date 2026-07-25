# 仓库架构总览

两个独立产品，三个部署形态，一个代码仓库。

## 产品定位

| 目录 | 产品 | 用户 | 核心职责 | 交付 |
|------|------|------|---------|------|
| `apps/` | TokenJoy | 客户管理员 | LLM API Key 管理、预算控制、组织架构、审计、Gateway | 客户私有化 / SaaS |
| `apps/web/` | 官网 | 公众访客 | 产品介绍、注册入口 | 公开部署 www.tokenjoy.com |
| `sms/` | SMS（供应商管理系统） | TJ 内部运营团队 | 供应商管理、模型目录、合同、定价发布、评测 | 仅内部部署 |

## apps/ — 客户侧产品（TokenJoy）

面向客户管理员的 LLM API 管控平台。Local（私有化）和 SaaS 共用同一套前后端代码，运行时通过 `SUPPORT_SAAS` flag 分流。

### apps/frontend/ — 管理后台 SPA
- 技术栈：React 19 + TypeScript + Vite + TanStack Query + Tailwind CSS v4
- 包名：`@tokenjoy/frontend`，端口 5173
- 领域模块（features/）：account、approval、audit、auth、billing、budget、dashboard、keys、models、mydashboard、notifications、org、query、session、workflow

### apps/backend/ — Go API 服务
- 技术栈：Go + PostgreSQL + Redis + BullMQ worker
- Module：`github.com/tokenjoy/backend`，端口 8010
- 内部结构：domain/（核心领域逻辑）、http/（handler）、store/（数据持久化）、worker/（后台任务）、adapter/、infra/、pkg/
- 领域模块（domain/）：approval、audit、billing、budget、company、dashboard、gateway、grants、keys、models、notification、org、usage 等

### apps/newapi/ — NewAPI Gateway
- Docker 构建 + 配置脚本
- 作为 LLM 请求的反向代理网关，端口 3010

### apps/web/ — 产品官网
- 技术栈：React 19 + Tailwind CSS v3（注意：不是 v4）+ Vite，无路由库/状态管理
- 包名：`@tokenjoy/web`，端口 5175
- 纯展示型轻量 SPA，页面由 sections 组合：Hero、Capabilities、Solutions、DeploymentModes、QuotaControl 等
- 认证集成：登录/注册通过 iframe 嵌入 App 的 `/embed.html`，postMessage 协议 `{ type: 'auth:success' }` 通知跳转
- 联调需同时运行 `pnpm start`（App）+ `pnpm start:web`（官网）

### apps/dev-mock-llm/ — 本地模拟 LLM 上游
- 开发环境用，模拟 OpenAI 兼容接口

## sms/ — 内部运营产品（供应商管理系统）

面向 TJ 运营团队的内部系统，管理 LLM 供应商、模型目录、合同定价。制品永远不出门。

### sms/frontend/ — 运营管理 SPA
- 技术栈：React 19 + TypeScript + Vite + TanStack Query + Tailwind CSS v4
- 包名：`@sms/frontend`，端口 5174
- 领域模块（features/）：contracts、dashboard、evaluations、models、orders、suppliers、users、query、session

### sms/backend/ — Go API 服务
- 技术栈：Go + PostgreSQL + Redis
- Module：`sms/backend`（独立，不引用 apps/backend），端口 8020
- 领域模块（domain/）：auth、contract、dashboard、evaluation、model、order、supplier、user、newapisync

### sms/newapi/ — NewAPI Gateway
- 独立于 apps/newapi，端口 3020

## packages/ — 跨产品共享层

| 包 | 用途 |
|----|------|
| `packages/contracts/` | 跨产品共享契约（permission codegen 等） |

## 核心规则

1. **两个产品，两个边界** — `apps/` 是客户的，`sms/` 是内部的，代码互不 import
2. **Local ↔ SaaS** — 同一前后端，运行时 `SUPPORT_SAAS` flag 分流
3. **跨产品通信仅 HTTP API** — 不共享 DB、不共享 domain 代码
4. **共享层放 packages/** — 跨产品共享的契约、类型放这里
5. **制品隔离** — 客户只拿 `apps/` 的 image，`sms/` 永不交付客户
6. **apps/web 是独立轻量应用** — 不依赖 apps/frontend 的任何代码，技术栈更简单

## 开发端口速查

| 服务 | apps | sms | web |
|------|------|-----|-----|
| Backend | 8010 | 8020 | — |
| Frontend | 5173 | 5174 | 5175 |
| NewAPI | 3010 | 3020 | — |
| Postgres | 5510（共用容器，不同 database） |||
| Redis | 6310（共用容器，不同 db number） |||

## 跨产品通信

```
Apps Backend (8010)  ──GET /api/v1/pricing/latest──→  SMS Backend (8020)
```

接口极少，耦合极低。两边独立开发、独立部署。
