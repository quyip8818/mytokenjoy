# 前端本地联调与演示数据

本文说明开发阶段如何用**真实 Go API** 跑通前端，同时借助**后端种子数据**验证各业务流程，并解释 `Unexpected token '<', "<!doctype "... is not valid JSON` 一类错误的成因。

相关文档：[Frontend-开发指南](./Frontend-开发指南.md) · [Frontend-API契约](./Frontend-API契约.md) · [Backend-设计](./Backend-设计.md)

---

## 1. 错误现象

浏览器控制台或 React Query 报错：

```
Unexpected token '<', "<!doctype "... is not valid JSON
```

### 1.1 含义

前端 `api/client.ts` 的 `request()` 在 **HTTP 2xx** 时会执行 `res.json()`。若响应体是 HTML（通常以 `<!doctype html>` 开头），JSON 解析就会失败。

### 1.2 常见原因（按频率）

| 原因           | 表现                         | 说明                                                                           |
| -------------- | ---------------------------- | ------------------------------------------------------------------------------ |
| **后端未启动** | 代理返回 502 JSON 或启动失败 | `pnpm start:frontend` 会等待 `:8080/healthz`；请用 `pnpm start` 或先起 backend |
| **路径错误**   | 请求打到静态资源而非 `/api`  | 检查 Network 里 URL 是否为 `/api/session` 等                                   |

### 1.3 与「未登录」的区别

未登录时，**代理配置正确且后端在跑**的情况下：

- `GET /api/session` 应返回 **401** + JSON：`{ "message": "..." }`
- `SessionGate` 会跳转 `/login`，**不会**出现 `<!doctype` 解析错误

因此：看到 `<!doctype` 时，优先检查代理与后端进程，而不是先怀疑鉴权。

---

## 2. 推荐方案：真实 API + 后端 Seed 演示数据

开发联调的演示数据来自 **Go 后端 Postgres + `internal/store/seed/`**，通过真实 HTTP 返回，与生产路径一致。

```
浏览器  →  fetch {BASE_URL}/api/*
         →  Vite dev/preview 同域反代（默认 → :8080）
         →  Go backend /api/*
         →  Postgres（空库首次启动由 seed 初始化）
```

优点：

- 走真实 `request()`、Cookie、权限中间件与 JSON 契约
- 组织、预算、密钥、看板、审计等域均有完整演示数据
- 写操作持久化于 Postgres（重启 backend 数据保留）

---

## 3. 快速开始

### 3.1 一键启动（推荐）

在仓库根目录（需 Docker，用于 Postgres）：

```bash
pnpm install
pnpm start
```

`pnpm start` 会先启动 Postgres（`apps/newapi/docker-compose.yml`），再并发 backend（`:8080`）与 frontend（Vite，默认 `:5173`），并等待 `http://127.0.0.1:8080/healthz` 就绪后再起前端。

### 3.2 单独起前端

根目录 `pnpm start:frontend` 会等待 backend `healthz` 后再启动 Vite。也可手动：

```bash
pnpm start:postgres   # 或 docker compose -f apps/newapi/docker-compose.yml up postgres -d --wait
cd apps/backend && DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable go run ./cmd/server   # 终端 1
pnpm start:frontend   # 终端 2（仓库根目录）
```

可复制 [`apps/backend/.env.example`](../apps/backend/.env.example) 为 `.env` 并 `export $(cat .env | xargs)`。

可选：复制 `apps/frontend/.env.development.example` 为 `.env.development` 以覆盖 `VITE_API_PROXY_TARGET`（默认 `http://127.0.0.1:8080`）。

Vite 在 **dev 与 preview** 均将 `{BASE_URL}/api` 反代到 Go，见 [`vite-api-proxy.ts`](../apps/frontend/vite-api-proxy.ts)。

### 3.3 Dev 登录

1. 打开 `http://localhost:5173/login`
2. 在 **Dev backend sign-in** 面板选择成员
3. 页面写入 Cookie `tokenjoy_session_member={memberId}` 并跳转首页

`GET /api/session` **始终需要**有效 Cookie（或 Bearer）；这与 `APP_PROFILE=demo` 下部分 GET 接口可匿名读取并不矛盾。

### 3.4 自检清单

```bash
# 1. 后端健康
curl -s http://127.0.0.1:8080/healthz

# 2. 未登录 session → 401 JSON（不是 HTML）
curl -s -i http://127.0.0.1:8080/api/session

# 3. 带 Cookie → 200 JSON
curl -s http://127.0.0.1:8080/api/session \
  -H 'Cookie: tokenjoy_session_member=m-admin'
```

浏览器 DevTools → Network：确认 `/api/session` 的 Response 为 JSON，且 Request Headers 含 Cookie。

---

## 4. 用不同身份验证流程

### 4.1 登录页成员与后端 Seed 对齐

`config/dev-members.ts` 中的成员 ID 与 `apps/backend/internal/store/seed/ids.go` 一致，切换身份即可测 RBAC：

| 登录选项            | memberId    | 典型用途                   |
| ------------------- | ----------- | -------------------------- |
| 管理员 · 超级管理员 | `m-admin`   | 全权限、组织/预算/密钥管理 |
| 李四 · 组织管理员   | `m-2`       | 组织管理、预算审批         |
| 张三 · API 调用者   | `m-1`       | 我的密钥、个人额度         |
| 孙审计 · 只读审计员 | `m-auditor` | 审计只读                   |
| 周八 · 普通成员     | `m-pure`    | 最小权限基线               |

顶栏 **Switch member**（仅 `import.meta.env.DEV`）可快速换回 `/login`。

### 4.2 演示数据覆盖范围

Seed 目录 `apps/backend/internal/store/seed/` 包含例如：

- 部门树、成员、角色与权限
- 预算树、预算组、告警规则
- 平台密钥、供应商密钥、审批单
- 模型列表与路由规则
- 看板费用/用量（`usage_buckets` 由 seed 灌入）
- 操作日志、调用日志（`data/*.json`）

默认 `APP_PROFILE=demo`：多数 **GET 读接口** 在 demo 模式下可不挂 Session（见 `PublicOrReadRoutes`）；**写接口** 仍需 Session + 权限。完整联调仍建议先 Dev 登录。

### 4.3 环境变量微调演示

| 变量             | 默认         | 作用                            |
| ---------------- | ------------ | ------------------------------- |
| `DEMO_TODAY`     | `2026-06-19` | 看板/用量时间轴锚点             |
| `SIMULATE_DELAY` | `true`       | 数据源 test/import 模拟延迟     |
| `APP_PROFILE`    | `demo`       | `prod` 时所有读接口也需 Session |

示例：

```bash
DEMO_TODAY=2026-06-01 DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable go run ./cmd/server
```

---

## 5. 其他场景的「假数据」

### 5.1 单元测试 / Hook 测试（不启 backend）

使用 `tests/utils.tsx` 的 **`createMockApis()`** 注入 `AppApis`，配合 `renderHookWithProviders` / `TestSessionProvider`：

```typescript
const apis = createMockApis({
  dashboardApi: {
    getCostSummary: vi.fn().mockResolvedValue({
      /* ... */
    }),
  },
})

renderHookWithProviders(() => useCostDashboardPage(apis), { apis })
```

- 不经过 `fetch`，不依赖 Vite 代理
- 适合验证页面 Hook、权限门控、workflow 编排
- 静态 fixture 放在 `apps/frontend/tests/fixtures/`

### 5.2 仅 UI 预览、无 API

若暂时无法起 backend，可：

1. 为单个页面写 Story / 测试，注入 `createMockApis`
2. 或临时在页面 Hook 传入 `injectedApis`（与生产 `defaultApis` 二选一，勿提交）

---

## 6. 请求链路速查

```
页面 Hook
  useApis() / useInjectedQuery
    → AppApis.{domain}Api.*
      → client.request('/{path}')
        → fetch(`${API_BASE_PATH}${path}`)
          → 同域反代 → Go :8080（dev / preview / 生产 nginx）
          → 若 /api 误走 SPA fallback → 返回 HTML（应检查边缘路由配置）
```

| 配置项                    | 位置                       | 作用                                       |
| ------------------------- | -------------------------- | ------------------------------------------ |
| `API_BASE_PATH`           | `config/app.ts`            | `{BASE_URL}/api`                           |
| `VITE_API_PROXY_TARGET`   | `.env.development`（可选） | 覆盖反代目标，默认 `http://127.0.0.1:8080` |
| `tokenjoy_session_member` | Dev 登录写入               | Session 成员 ID                            |
| `SESSION_COOKIE`          | `config/auth.ts`           | Cookie 名常量                              |

---

## 7. 故障排查

### 7.1 API 返回 HTML 或非 JSON

1. `curl http://127.0.0.1:8080/healthz` 必须成功
2. 使用 `pnpm start` 或 `pnpm start:frontend`（会等待 backend）
3. Network 中确认 `/api/*` 的 `Content-Type` 为 `application/json`
4. 生产环境检查 nginx：`location /api/` 必须在 SPA `try_files` 之前（见 `deploy/nginx.conf.example`）

### 7.2 401 / 一直跳转登录

1. 访问 `/login` 选择成员
2. Application → Cookies 检查 `tokenjoy_session_member`
3. memberId 须存在于 seed（如 `m-admin`），否则 backend 返回 401

### 7.3 200 但页面空白或 schema 错误

1. 对比响应与 `api/types/`、`api/schemas/session.ts`
2. 后端字段须 **camelCase**，与契约一致

### 7.4 CORS 问题

本地默认 `CORS_ORIGINS=http://localhost:5173`。若 Vite 用了其他端口，启动 backend 时设置：

```bash
CORS_ORIGINS=http://localhost:5173,http://127.0.0.1:5173 go run ./cmd/server
```

---

## 8. 自定义 / 扩展演示数据

| 目标                 | 做法                                                                                                                                         |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| 改成员、部门、密钥等 | 编辑 `internal/store/seed/*.go` 或 `internal/store/seed/data/*.json`                                                                         |
| 改看板日期范围       | 调整 `DEMO_TODAY` 与 `internal/store/seed/usage.go`                                                                                          |
| 重新灌 seed          | `docker compose -f apps/newapi/docker-compose.yml down -v && pnpm start:postgres`，或清空 `domain_snapshot` + `usage_buckets` 后重启 backend |
| 对接真实 NewAPI      | `NEW_API_ENABLED=true` 及相关变量（Relay 场景，非管理面 seed）                                                                               |

改 seed 后：若需重新初始化，按上表重置数据库再重启 backend。

---

## 9. 模式对比

| 模式                 | 命令                                            | 数据来源                              | 适用场景                     |
| -------------------- | ----------------------------------------------- | ------------------------------------- | ---------------------------- |
| **全栈联调（推荐）** | 根目录 `pnpm start` + Dev 登录                  | Postgres + seed（持久化）             | 验证完整业务流程、权限、契约 |
| **前端 + 代理**      | backend 手动起 + frontend + `.env.development`  | 同上                                  | 只调试前端时                 |
| **Vitest**           | `pnpm -F @tokenjoy/frontend test`               | `createMockApis` / fixtures           | Hook、组件、回归             |
| **生产构建预览**     | `pnpm build` + `pnpm preview`（backend 运行中） | 后端 seed；preview 已配置 `/api` 反代 |

---

## 10. 小结

- `<!doctype` JSON 错误 ≈ **`/api` 没有打到 Go 后端**，先配代理、起 backend，再谈登录。
- 开发阶段「假数据」的正式来源是 **`apps/backend/internal/store/seed/`**，经 Postgres 持久化后通过真实 API 返回。
- 验证不同角色：Dev 登录切换成员；验证逻辑：Vitest + `createMockApis()`。
- 日常最省事：`pnpm start` → `/login` 选身份 → 按业务域点一遍。
