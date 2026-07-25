# 跨包共享模块分析

分析对象：`apps/frontend`、`apps/web`、`sms/frontend`

## 现状

| 包 | 定位 | 关键依赖 |
|---|---|---|
| `@tokenjoy/frontend` | 主产品 SPA | react-query, radix-ui, cva, zustand, react-router, sonner |
| `@tokenjoy/web` | 营销/官网静态站 | react, clsx, tailwind-merge, lucide-react（极简） |
| `@sms/frontend` | SMS 产品 SPA | react-query, radix-ui, cva, zustand, react-router, sonner |

`apps/frontend` 与 `sms/frontend` 重叠度极高（相同技术栈、相同架构模式）。`apps/web` 是纯静态站，仅共享 `cn()` 工具函数。

---

## 可提取的共享包

### 1. `packages/ui-utils` — 样式工具（优先级：高，成本：极低）

**重复代码：** 三个包都有完全相同的 `cn()` 函数。

```ts
// apps/frontend/src/lib/utils.ts
// apps/web/src/shared/cn.ts
// sms/frontend/src/lib/utils.ts
import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'
export function cn(...inputs: ClassValue[]) { return twMerge(clsx(inputs)) }
```

**建议包内容：**
- `cn()` 函数
- 未来可放入其他纯工具（格式化金额、日期等，如果出现跨产品复用）

**依赖：** `clsx` + `tailwind-merge`（零 React 依赖）

---

### 2. `packages/query` — React Query 基础设施（优先级：高，成本：低）

**重复代码：** `apps/frontend` 和 `sms/frontend` 的 `features/query/` 几乎完全一样：

| 文件 | apps/frontend | sms/frontend | 差异 |
|---|---|---|---|
| `query-client.ts` | `createAppQueryClient` + `createTestQueryClient` | `createAppQueryClient` | sms 缺 test helper |
| `use-injected-query.ts` | 泛型 + 完整 options | 简化版（只有 enabled） | 接口宽窄不同 |
| `use-injected-mutation.ts` | invalidateKeys 支持函数 | invalidateKeys 只支持数组 | apps 更丰富 |
| `use-filtered-query.ts` | 泛型 F，返回 setFilter | 泛型 F extends {page?}，多 setPage/search | sms 多了分页快捷方法 |
| `query-provider.tsx` | 含 devtools | 无 devtools | devtools 可选 |

**建议包内容：**
- `createAppQueryClient` / `createTestQueryClient`
- `useInjectedQuery` — 取 apps/frontend 更完整的签名
- `useInjectedMutation` — 合并两端特性（invalidateKeys 支持数组 + 函数）
- `useFilteredQuery` — 合并两端特性（setPage/search 作为返回值可选）
- `QueryProvider` — devtools 通过参数控制

**泛型约束：** 包接受 `AppApis` 作为类型参数（每个产品传入自己的 API 类型），不依赖具体业务。

**依赖：** `@tanstack/react-query`, `react`

---

### 3. `packages/api-core` — HTTP 客户端内核（优先级：中，成本：中）

**重复代码：** `apps/frontend/src/api/client.ts` 与 `sms/frontend/src/api/client.ts` 结构相同：

| 概念 | apps/frontend | sms/frontend |
|---|---|---|
| `ApiError` class | ✅ (status + retryAfter) | ✅ (status only) |
| `request<T>()` | cookie auth + 401 refresh + event bus | Bearer token + 401 refresh + redirect |
| `buildQuery()` | ✅ | ✅（完全相同） |
| `api-context.ts` | `createContext<AppApis \| null>` | 完全相同 |
| `use-apis.ts` | `useApis` + `useInjectedApis` | 完全相同 |

**差异点：** 认证策略不同（cookie vs Bearer token），401 处理后行为不同（event bus vs redirect）。

**建议包内容：**
- `ApiError` class（统一，支持 retryAfter 可选）
- `buildQuery()` 纯函数
- `createApiContext<T>()` + `useApis<T>()` + `useInjectedApis<T>()` — 泛型化的 Context 工厂
- `createRequest(config)` — 工厂函数，接收配置：
  - `basePath: string`
  - `getAuthHeader?: () => Record<string, string>`
  - `onRefresh?: () => Promise<boolean>`
  - `on401?: () => void`
  - `on403?: (path: string) => void`

每个产品用自己的配置调用工厂，得到定制的 `request` 函数。

**依赖：** `react`（仅 Context 部分）

---

### 4. `packages/ui`（或扩展现有 `packages/contracts`）— 共享 UI 原子组件（优先级：低，成本：高）

**现状分析：**

两个 SPA 的 UI 组件（button、badge、input、dialog 等）模式相同（cva + radix + cn），但样式已经分叉：
- `apps/frontend` 用 shadcn/ui 最新版风格（data-slot、更复杂的变体）
- `sms/frontend` 用较早的 shadcn 风格（forwardRef、简单变体）

| 组件 | 共享程度 |
|---|---|
| Button | 模式相同，变体/样式完全不同 |
| Badge | 模式相同，变体不同 |
| Input | 模式相同，class 不同 |
| EmptyState | 接口完全不同 |
| Dialog | 模式相同 |

**结论：暂不建议抽取 UI 组件包。**

原因：
1. 两个产品视觉风格不同，强行统一会制造耦合
2. 组件数量和复杂度不对等（apps 35 个 vs sms 9 个）
3. 改动频率高，共享组件的版本协调成本 > 复制成本

如果未来两个产品设计系统统一，再考虑抽取。

---

## 不建议共享的部分

| 模块 | 原因 |
|---|---|
| UI 组件 | 两产品设计系统已分叉，强行合并得不偿失 |
| Session/Auth | 认证流程完全不同（多租户 RBAC vs 简单 JWT） |
| 路由配置 | 业务强相关，无复用价值 |
| apps/web 的页面组件 | 静态站，与 SPA 无交集 |

---

## 建议实施顺序

| 阶段 | 包 | 工作量 | 收益 |
|---|---|---|---|
| 1 | `packages/ui-utils` | ~1h | 消除最简单的重复，三包都受益 |
| 2 | `packages/query` | ~3h | 消除整套 query 基础设施重复，两 SPA 受益 |
| 3 | `packages/api-core` | ~4h | 消除 HTTP 客户端重复，但需要设计好工厂接口 |

阶段 1 和 2 可以一起做，ROI 最高。阶段 3 因为涉及认证策略差异，需要更仔细的接口设计。

---

## 包目录结构预览

```
packages/
├── contracts/          # 已有：权限/通知契约
├── ui-utils/
│   ├── package.json    # deps: clsx, tailwind-merge
│   └── src/
│       └── cn.ts
├── query/
│   ├── package.json    # deps: @tanstack/react-query, react
│   └── src/
│       ├── index.ts
│       ├── query-client.ts
│       ├── query-provider.tsx
│       ├── use-injected-query.ts
│       ├── use-injected-mutation.ts
│       └── use-filtered-query.ts
└── api-core/
    ├── package.json    # deps: react
    └── src/
        ├── index.ts
        ├── error.ts          # ApiError
        ├── build-query.ts    # buildQuery()
        ├── context.ts        # createApiContext, useApis, useInjectedApis
        └── create-request.ts # 工厂函数
```
