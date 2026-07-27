# Deep Link 生成器设计

> 模块化 URL 拼装工具，让路由跳转类型安全、可维护。

---

## 现状问题

```ts
// 1. 手动拼字符串，typo 无感知
navigate('/me/settings?tab=notifications')

// 2. getActionUrl 里大量裸字符串
return `/keys/platform?highlight=${payload.keyID}`

// 3. ROUTES 对象只有 path，不支持 query params
export const ROUTES = { mySettings: '/me/settings', ... }
```

问题：路径改了 → grep 全量搜 → 忘一处就 404。

---

## 同行方案对比

| 方案 | 代表 | 优点 | 缺点 |
|------|------|------|------|
| TanStack Router search params | Vercel/TanStack | 编译期完全类型安全，自动 parse/serialize | 需要换路由库，重构成本极高 |
| react-router-typesafe-routes | fenok/github | 跟 react-router 兼容，schema-based | 额外依赖，学习成本 |
| 手写 route builder 对象 | Next.js 社区常见 | 零依赖，项目控制力强 | 需要手动维护 |
| 常量 + query helper 函数 | Linear/Stripe 内部 | 最小方案，渐进式 | 没有编译期路径校验 |

---

## 推荐方案：路由常量 + typed link builder

不引入新依赖，在现有 `ROUTES` 基础上扩展一个 `link()` 工具函数。

### 设计

```ts
// src/lib/link.ts

import { ROUTES } from '@/config/routes'

type QueryParams = Record<string, string | number | boolean | undefined>

/**
 * 构建带 query params 的路由链接。
 * 
 * @example
 * link(ROUTES.mySettings, { tab: 'notifications' })
 * // → '/me/settings?tab=notifications'
 * 
 * link(ROUTES.keysPlatform, { highlight: keyID })
 * // → '/keys/platform?highlight=xxx'
 * 
 * link(ROUTES.mySettings)
 * // → '/me/settings'
 */
export function link(path: string, params?: QueryParams): string {
  if (!params) return path
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === '') continue
    search.set(key, String(value))
  }
  const qs = search.toString()
  return qs ? `${path}?${qs}` : path
}
```

### 使用方式

```ts
// Before
navigate('/me/settings?tab=notifications')

// After
navigate(link(ROUTES.mySettings, { tab: 'notifications' }))
```

```ts
// Before (getActionUrl)
return `/keys/platform?highlight=${payload.keyID}`

// After
return link(ROUTES.keysPlatform, { highlight: payload.keyID as string })
```

### 优势

1. **路径不再硬编码** — 全部走 `ROUTES.xxx`，改一处全局生效
2. **query params 自动过滤 undefined** — 不会产生 `?foo=undefined`
3. **零依赖** — 纯函数，30 行代码
4. **渐进式迁移** — 不需要一次改完，新代码用 `link()`，旧代码可以慢慢迁
5. **TypeScript 补全** — `ROUTES.` 有所有路由 key 的自动补全

### 不做什么

- 不做路径参数替换（`/users/:id` → `/users/123`）—— 当前项目路由全是静态路径
- 不做编译期 query params 类型校验 —— ROI 低，运行时 undefined 过滤已经够了
- 不引入新路由库 —— react-router 够用

---

## 实现范围

| 文件 | 改动 |
|------|------|
| `src/lib/link.ts` | 新增 `link()` 函数 |
| `notification-center.tsx` | navigate 改用 `link()` |
| `notification-inbox.tsx` | navigate 改用 `link()` |
| `get-action-url.ts` | 全部改用 `link()` + `ROUTES` |
| `header.tsx` | navigate 改用 `link()` |

总改动量：~10 行新代码 + ~15 处引用替换。
