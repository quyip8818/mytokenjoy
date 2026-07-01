# TokenJoy 前端开发指南

`apps/frontend` 架构说明与开发规范。API 契约见 [Frontend-API契约.md](./Frontend-API契约.md)；常用命令见 [README.md](./README.md#常用命令)。

---

## 1. 相关文档

| 文档                                                                              | 职责                  |
| --------------------------------------------------------------------------------- | --------------------- |
| [Frontend-API契约.md](./Frontend-API契约.md)                                      | REST 路径、类型、鉴权 |
| [Backend-设计.md](./Backend-设计.md)                                              | Go 服务、联调方式     |
| [TokenJoy-PRD.md](./TokenJoy-PRD.md)                                              | 产品需求              |
| [`.cursor/rules/frontend-structure.mdc`](../.cursor/rules/frontend-structure.mdc) | AI 速查摘要           |

---

## 2. 技术栈

React 19、React Router 7、Vite 8、TypeScript 5.x、Tailwind CSS 4、Base UI / shadcn、TanStack Table、Recharts、react-hook-form、Zustand、TanStack Query、Vitest。CI 使用 Node 24、Go 1.24。

路径别名：`@/` → `src/`，`@tests/` → `tests/`。

---

## 3. 运行时

```
main.tsx → App.tsx
└─ AdminLayout
   ├─ ApiProvider + QueryProvider
   ├─ AuthSessionProvider + AuthUnauthorizedBridge + SessionNavigationBridge
   └─ WorkflowProvider → Sidebar / Header / Outlet / WorkflowPanelStack
```

- 根目录 `pnpm start`：backend `:8080` + Vite；同域 `/api` 反代到 Go（dev 与 preview 均启用，见 `vite-api-proxy.ts`）
- Dev 登录：`/login` 选成员 → `tokenjoy_session_member` cookie → `GET /session`
- 首页 `/` 按权限跳转 `HOME_PATH_CANDIDATES`

---

## 4. 目录结构

```
apps/frontend/
├── tests/                  Vitest（镜像 src）
└── src/
    ├── config/             routes.ts（ROUTE_DEFINITIONS 单源）、nav.ts、app.ts、dev-members.ts
    ├── routes/{domain}/    页面 + hooks/ + components/
    ├── components/         ui/、layout/、auth/、{domain}/
    ├── features/           session/、workflow/、query/
    ├── api/                client + 域 API + types/
    ├── hooks/              跨域 Hook
    └── lib/                纯函数、权限、labels
```

**分层：** `routes/` 对应页面域；`components/{domain}/` 跨页复用；`features/` 独立 Provider 能力；`api/` + `lib/` 无 UI。

---

## 5. 路由

[`config/routes.ts`](../apps/frontend/src/config/routes.ts) 以 **`ROUTE_DEFINITIONS`** 为唯一手写源，派生 `ROUTES`、`APP_ROUTES`、`NAV_GROUP_LAYOUT` 等。

**新增页面：** 在 `ROUTE_DEFINITIONS` 加一条 → 创建 `{page}.tsx` + `hooks/use-{page}-page.ts`。`pnpm lint` 的 `check-conventions` 会校验。

当前 **16** 个业务页：dashboard（cost、usage）、org（3）、budget（3）、keys（4）、models（2）、audit（2）。

---

## 6. API 层

- `api/client.ts`：`request()`、`ApiError`、`buildQuery()`；`credentials: 'include'`
- `app-apis.ts`：`AppApis` + `defaultApis`（14 个命名空间）
- 生产：`AdminLayout` 注入 `defaultApis`；测试：`createMockApis()` + `injectedApis`

改 API 须同步：`api/{domain}.ts`、`api/types/`、契约文档、`query-keys.ts`（读操作）。

**SaaS 未接入：** 后端已实现 `auth` / `billing` / `platform` API（契约 §10），但 `AppApis` 仅 14 个命名空间；控制台无 `/platform/login`、`/invite/accept`、充值页。

---

## 7. 页面架构

**薄页面 → 页面 Hook → 展示组件**

- Hook：`useApis()`、`useInjectedQuery` + `queryKeys`、workflow 编排；返回扁平 view model，无 JSX
- 组件：props 受控，不直接 `import { xxxApi }`
- Workflow 成功后：`useWorkflowRefresh(refresh)` 或 `invalidateQueries`

---

## 8. Workflow

`features/workflow/`：Zustand store + 侧滑栈。新增：`workflows/{name}.tsx` + payload + `definitions/{domain}.ts` 注册。复杂表单用 `defineDelegateWorkflow`。

---

## 9. 测试

`tests/setup.ts`、`createMockApis()`、`renderHookWithProviders()`；不依赖 backend 进程。静态 fixture 在 `tests/fixtures/`。

优先级：`lib/` 纯函数 → Hook → 页面 Hook（`injectedApis`）。

---

## 10. 代码放置

| 代码       | 位置                                       |
| ---------- | ------------------------------------------ |
| 页面入口   | `routes/{domain}/{page}.tsx`               |
| 页面逻辑   | `routes/{domain}/hooks/use-{page}-page.ts` |
| 单页 UI    | `routes/{domain}/components/`              |
| 跨页 UI    | `components/{domain}/`                     |
| Primitive  | `components/ui/`                           |
| HTTP / DTO | `api/{domain}.ts`、`api/types/`            |
| 纯逻辑     | `lib/`                                     |

禁止硬编码路由字符串（用 `ROUTES.*`）；禁止 `components/ui` 含业务语义。

---

## 11. 工具链

| 命令          | 作用                       |
| ------------- | -------------------------- |
| `pnpm start`  | backend + frontend         |
| `pnpm lint`   | ESLint + check-conventions |
| `pnpm test`   | Vitest                     |
| `pnpm build`  | tsc + vite build           |
| `pnpm verify` | lint + test + build + backend build:check |

---

## 12. 生产同域部署

生产环境使用 **同域反向代理**：`/api/*` 转发到 Go，页面路径才走 SPA `index.html` fallback。

示例配置：[`deploy/nginx.conf.example`](../deploy/nginx.conf.example)。`location /api/` 必须写在 `try_files ... /index.html` **之前**，否则 `/api` 会被静态托管误回退为 HTML。

---

## 13. PR 自检

- [ ] 新页面只改 `ROUTE_DEFINITIONS` 一条
- [ ] 页面从 `use-*-page` 取数，无内联 `useApis`
- [ ] 新 API：契约 + `api/` + `queryKeys` + 后端 handler
- [ ] `pnpm lint` 与 `pnpm test` 通过
