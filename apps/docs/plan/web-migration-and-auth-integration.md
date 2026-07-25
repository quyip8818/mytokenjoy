# 官网迁入 Monorepo + 认证集成

## 概述

将 `tokenjoy-web`（官网）迁入 monorepo 为 `apps/web/`，官网登录/注册通过 iframe 嵌入 App 的认证页面实现。官网本身保持纯展示型轻量 SPA，不引入任何认证逻辑。

---

## 架构决策

### 为什么不做 packages/auth-popup 共享包

| 因素 | 共享包方案 | iframe 方案 |
|------|-----------|-------------|
| 官网依赖 | 需引入 Radix/Select/AvatarPicker + 认证状态机 | 零认证依赖 |
| 维护成本 | 两端消费同一包，UI 变更需兼容测试 | 认证 UI 只在 App 维护一份 |
| 包内 UI | 需自带 Dialog/Input/Button/Select（远超 200 行） | 不需要 |
| 官网体积 | 增加 ~50KB（Radix + 认证逻辑） | 增加 0 |
| App 认证改版 | 需同步发包、两端验证 | 官网零改动 |

结论：官网追求轻量，认证逻辑 100% 留在 App 端，通过 iframe 嵌入提供不离开官网的体验。

---

## 迁移后仓库结构

```
mytokenjoy/
├── apps/
│   ├── web/                     ← 官网（marketing site）
│   │   ├── src/
│   │   │   ├── content/         ← 文案数据（nav、hero、sections 等）
│   │   │   ├── pages/           ← Home（单页 + 懒加载 HomeRest）
│   │   │   ├── sections/        ← Hero / Challenges / Solutions / Capabilities / ...
│   │   │   └── shared/          ← Navbar、Logo、Footer、icons、BrandIcons
│   │   ├── public/              ← 静态资源（logo、favicon 等）
│   │   ├── package.json         ← @tokenjoy/web
│   │   ├── vite.config.ts
│   │   └── tailwind.config.js   ← Tailwind v3（独立于 apps/frontend 的 v4）
│   │
│   ├── frontend/                ← 不变（SaaS 管理后台）
│   ├── backend/                 ← 不变
│   └── docs/                    ← 不变
│
├── sms/                         ← 不变
├── packages/                    ← 不变（contracts 等）
└── pnpm-workspace.yaml          ← 不需改（已有 'apps/*'）
```

官网放 `apps/web/`，命中已有的 `apps/*` workspace pattern，无需修改 `pnpm-workspace.yaml`。

---

## 认证集成方案：iframe + postMessage

### 流程

```
用户在官网点击"登录"或"免费试用"
  ↓
官网弹出 modal，内嵌 <iframe src="https://app.tokenjoy.com/auth/embed?mode=register">
  ↓
App 渲染 /auth/embed 页面（AuthCard，无 Dialog 包裹，无 App chrome）
  ↓
iframe 加载完成后发送 postMessage({ type: 'auth:ready' }) 到 parent
  ↓
官网收到 auth:ready，确认 iframe 正常工作
  ↓
用户完成认证，Cookie 写入 .tokenjoy.com 域
  ↓
App iframe 内 postMessage({ type: 'auth:success' }, targetOrigin) 到 parent
  ↓
官网收到消息，校验 origin，关闭 modal，跳转 https://app.tokenjoy.com
```

### App 端组件拆分：AuthCard

当前 `AuthPopup` 使用 Radix `<Dialog>` 包裹。iframe embed 场景不需要 Dialog overlay/portal 行为。
需要从 `AuthPopup` 中提取核心表单逻辑为 `AuthCard` 组件：

- `AuthCard`：纯认证表单（tabs + 各 step 表单），无 Dialog、无 overlay
- `AuthPopup`：`<Dialog>` + `<AuthCard />`，保持现有 App 内弹窗行为不变

```typescript
// apps/frontend/src/features/auth/components/auth-card.tsx
// 从 auth-popup.tsx 提取的核心表单逻辑，不含 Dialog 包裹

interface AuthCardProps {
  defaultMode?: AuthMode
  onSuccess?: () => void
}

export function AuthCard({ defaultMode = 'login', onSuccess }: AuthCardProps) {
  // 现有 AuthPopup 的全部表单状态和 step 逻辑移到这里
  // ...
}
```

```typescript
// apps/frontend/src/features/auth/components/auth-popup.tsx（改造后）
import { AuthCard } from './auth-card'

export function AuthPopup({ open, defaultMode, closable, onSuccess, onClose }: AuthPopupProps) {
  return (
    <Dialog open={open} onOpenChange={...}>
      <DialogContent ...>
        <AuthCard defaultMode={defaultMode} onSuccess={onSuccess} />
      </DialogContent>
    </Dialog>
  )
}
```

### App 端：新增 /auth/embed 路由

```typescript
// apps/frontend/src/routes/auth/embed.tsx
import { useEffect } from 'react'
import { useSearchParams } from 'react-router'
import { AuthCard } from '@/features/auth'

const PARENT_ORIGIN = import.meta.env.VITE_WEB_ORIGIN || 'https://www.tokenjoy.com'

export default function AuthEmbedPage() {
  const [params] = useSearchParams()
  const mode = params.get('mode') === 'register' ? 'register' : 'login'

  // iframe 加载完成后通知父窗口
  useEffect(() => {
    if (window.parent !== window) {
      window.parent.postMessage({ type: 'auth:ready' }, PARENT_ORIGIN)
    }
  }, [])

  const handleSuccess = () => {
    if (window.parent !== window) {
      window.parent.postMessage({ type: 'auth:success' }, PARENT_ORIGIN)
    } else {
      // 直接访问时跳转 Dashboard
      window.location.href = '/'
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <AuthCard defaultMode={mode} onSuccess={handleSuccess} />
    </div>
  )
}
```

关键约束：
- `/auth/embed` 是公开路由，在 `App.tsx` 中与 `/login` 平级注册（在 SessionGate 之外）
- 渲染时不加载 App 的导航、侧边栏等 chrome
- `postMessage` 的 `targetOrigin` 必须指定为官网域名，禁止使用 `'*'`

### App.tsx 路由注册

```typescript
// apps/frontend/src/App.tsx 新增路由（与 LoginPage 平级）
const AuthEmbedPage = lazy(() => import('@/routes/auth/embed'))

// 在 <Routes> 中：
<Route
  path="auth/embed"
  element={
    <Suspense fallback={<RouteFallback />}>
      <AuthEmbedPage />
    </Suspense>
  }
/>
```

### 官网端：Navbar 集成

```tsx
// apps/web/src/shared/Navbar.tsx
import { useState, useEffect, useRef } from 'react'

const AUTH_EMBED_URL = import.meta.env.VITE_AUTH_EMBED_URL || 'http://localhost:5173/auth/embed'
const APP_ORIGIN = import.meta.env.VITE_APP_ORIGIN || 'http://localhost:5173'
const APP_URL = import.meta.env.VITE_APP_URL || 'http://localhost:5173'

export function Navbar() {
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login')
  const [iframeReady, setIframeReady] = useState(false)
  const [iframeError, setIframeError] = useState(false)
  const dialogRef = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    if (!authOpen) {
      setIframeReady(false)
      setIframeError(false)
      return
    }

    // 打开 modal
    dialogRef.current?.showModal()

    // 超时检测：5s 内没收到 auth:ready 则视为加载失败
    const timeout = setTimeout(() => {
      if (!iframeReady) setIframeError(true)
    }, 5000)

    const handler = (e: MessageEvent) => {
      // 严格校验来源
      if (e.origin !== APP_ORIGIN) return

      if (e.data?.type === 'auth:ready') {
        setIframeReady(true)
        setIframeError(false)
      }
      if (e.data?.type === 'auth:success') {
        dialogRef.current?.close()
        setAuthOpen(false)
        window.location.href = APP_URL
      }
    }
    window.addEventListener('message', handler)
    return () => {
      clearTimeout(timeout)
      window.removeEventListener('message', handler)
    }
  }, [authOpen, iframeReady])

  const closeModal = () => {
    dialogRef.current?.close()
    setAuthOpen(false)
  }

  return (
    <nav>
      {/* ... 导航内容 ... */}
      <button onClick={() => { setAuthMode('login'); setAuthOpen(true) }}>登录</button>
      <button onClick={() => { setAuthMode('register'); setAuthOpen(true) }}>免费试用</button>

      <dialog
        ref={dialogRef}
        className="backdrop:bg-black/50 bg-transparent p-0 m-auto rounded-2xl"
        onClose={() => setAuthOpen(false)}
      >
        <div className="relative w-[480px] h-[640px] rounded-2xl overflow-hidden bg-white shadow-2xl">
          <button
            onClick={closeModal}
            className="absolute top-3 right-3 z-10 text-gray-400 hover:text-gray-600"
            aria-label="关闭"
          >
            ✕
          </button>
          {iframeError ? (
            <div className="flex h-full items-center justify-center text-center p-6">
              <div>
                <p className="text-gray-600 mb-4">加载失败，请重试</p>
                <button
                  onClick={() => { setIframeError(false); setIframeReady(false) }}
                  className="text-primary underline"
                >
                  重新加载
                </button>
              </div>
            </div>
          ) : (
            <iframe
              src={`${AUTH_EMBED_URL}?mode=${authMode}`}
              className="w-full h-full border-0"
              title="TokenJoy 认证"
            />
          )}
        </div>
      </dialog>
    </nav>
  )
}
```

### postMessage 协议

| 方向 | 消息 | 含义 |
|------|------|------|
| iframe → parent | `{ type: 'auth:ready' }` | iframe 加载完成，认证表单就绪 |
| iframe → parent | `{ type: 'auth:success' }` | 认证完成，可跳转 |
| iframe → parent | `{ type: 'auth:close' }` | 用户请求关闭（预留） |

安全约束：
- App iframe 发送 postMessage 时 **必须** 指定 `targetOrigin`（不允许 `'*'`），通过 `VITE_WEB_ORIGIN` 环境变量配置
- 官网接收消息时 **必须** 校验 `e.origin === APP_ORIGIN`，丢弃不匹配的消息
- 握手机制：官网等待 `auth:ready`，5 秒超时展示错误提示

---

## Cookie 方案

```
tokenjoy.com（父域）
├── www.tokenjoy.com    — 官网（apps/web/）
├── app.tokenjoy.com    — SaaS 管理后台（apps/frontend/）
└── api.tokenjoy.com    — 后端 API（apps/backend/）
```

| 配置项 | 开发环境 | 生产环境 |
|--------|---------|---------|
| Cookie Domain | （空，默认 localhost） | `.tokenjoy.com` |
| SameSite | Lax | Lax |
| Secure | false | true |

开发环境说明：官网（localhost:5175）和 App（localhost:5173）同为 localhost，Cookie 不区分端口，因此开发环境 Cookie 天然共享，无需额外配置。

生产环境说明：`www`、`app`、`api` 同属 `.tokenjoy.com`，是 **same-site**，`SameSite=Lax` 对同站请求没有限制。需要处理的是 CORS（cross-origin），不是 cookie 传递。

### CORS 配置

后端需要在 CORS 白名单中加入官网域名：

| 环境 | 允许的 Origins |
|------|---------------|
| 开发 | `http://localhost:5173`, `http://localhost:5175` |
| 生产 | `https://app.tokenjoy.com`, `https://www.tokenjoy.com` |

### iframe 安全头（部署层配置）

App 端需要允许官网通过 iframe 嵌入 `/auth/embed`。

**重要**：由于 App 是 SPA（单一 `index.html`），无法在应用代码中按路由返回不同响应头。iframe 安全头需要在**部署/反向代理层**配置（Nginx、CloudFront、Vercel headers 等）。

Nginx 示例：
```nginx
# /auth/embed 路径允许被官网嵌入
location /auth/embed {
    add_header Content-Security-Policy "frame-ancestors https://www.tokenjoy.com" always;
    try_files $uri /index.html;
}

# 其他路径禁止被嵌入
location / {
    add_header Content-Security-Policy "frame-ancestors 'none'" always;
    try_files $uri /index.html;
}
```

注意：`X-Frame-Options: ALLOW-FROM` 已被现代浏览器废弃（Chrome/Edge 从未支持），只需使用 `Content-Security-Policy: frame-ancestors`。

---

## 本地开发

### 端口分配

| 服务 | 端口 |
|------|------|
| 官网 (apps/web/) | 5175 |
| Apps frontend | 5173 |
| Apps backend | 8010 |

### 环境变量 (apps/web/.env.development)

```env
VITE_AUTH_EMBED_URL=http://localhost:5173/auth/embed
VITE_APP_ORIGIN=http://localhost:5173
VITE_APP_URL=http://localhost:5173
```

### 环境变量 (apps/frontend/.env.development 新增)

```env
VITE_WEB_ORIGIN=http://localhost:5175
```

### 根 package.json 脚本（新增）

```jsonc
{
  "scripts": {
    "start:web": "pnpm -F @tokenjoy/web start",
    "build:web": "pnpm -F @tokenjoy/web build",
    "lint:web": "pnpm -F @tokenjoy/web lint"
  }
}
```

### 开发流程

```bash
pnpm start:web    # 启动官网 dev server (port 5175)
pnpm start        # 启动 apps（backend + frontend）
```

官网 iframe 指向 `localhost:5173/auth/embed`，本地联调时需要两个 dev server 同时运行。

---

## 依赖版本对齐

迁入 monorepo 前，`tokenjoy-web` 核心依赖升级到与 monorepo 一致。

### 必须对齐（跟随 monorepo 当前版本）

| 依赖 | tokenjoy-web 当前 | 目标 | 原因 |
|------|-------------------|------|------|
| react / react-dom | 18.3 | 与 monorepo 一致 | 统一 React 版本，避免 pnpm hoist 冲突 |
| typescript | 5.8 | 与 monorepo 一致 | project references 统一 |
| vite | 6.3 | 与 monorepo 一致 | 统一避免 pnpm hoist 冲突 |
| lucide-react | 0.511 | 与 monorepo 一致 | 0.x → 1.x breaking（图标重命名） |

### 保持独立

| 依赖 | 保持版本 | 原因 |
|------|---------|------|
| tailwindcss | v3.4 | 官网视觉独立，v4 配置模式不兼容 |
| postcss / autoprefixer | 保持 | 配合 Tailwind v3 |

### apps/web/package.json

```json
{
  "name": "@tokenjoy/web",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "start": "vite --port 5175",
    "build": "tsc -b && vite build",
    "lint": "eslint .",
    "preview": "vite preview --port 5175"
  },
  "dependencies": {
    "clsx": "^2.1.1",
    "lucide-react": "^1.20.0",
    "react": "^19.2.6",
    "react-dom": "^19.2.6",
    "tailwind-merge": "^3.6.0"
  },
  "devDependencies": {
    "@eslint/js": "^10.0.1",
    "@types/react": "^19.2.14",
    "@types/react-dom": "^19.2.3",
    "@vitejs/plugin-react": "^6.0.1",
    "autoprefixer": "^10.4.21",
    "eslint": "^10.3.0",
    "eslint-config-prettier": "^10.1.8",
    "eslint-plugin-react-hooks": "^7.1.1",
    "eslint-plugin-react-refresh": "^0.5.2",
    "globals": "^17.6.0",
    "postcss": "^8.5.3",
    "tailwindcss": "^3.4.17",
    "typescript": "~6.0.2",
    "typescript-eslint": "^8.59.2",
    "vite": "^8.0.12",
    "vite-tsconfig-paths": "^5.1.4"
  }
}
```

无 react-router、无 zustand、无 tanstack-query、无 Radix、无状态管理库。官网就是一个 landing page。

> 注：以上具体版本号以迁移时 monorepo 实际使用的版本为准，不必提前锁定。

---

## App 端改动清单

| 改动 | 描述 |
|------|------|
| 提取 AuthCard 组件 | 从 AuthPopup 中提取核心表单逻辑为 AuthCard（无 Dialog 包裹） |
| AuthPopup 改为组合 | AuthPopup = Dialog + AuthCard，行为不变 |
| 新增路由 `/auth/embed` | 公开页面，渲染 AuthCard + postMessage 通知 |
| App.tsx 路由注册 | `/auth/embed` 与 `/login` 平级，不走 SessionGate，不加载 App shell |
| features/auth/index.ts | 导出 AuthCard |
| CORS 白名单 | 加入 `https://www.tokenjoy.com` |
| 后端 COOKIE_DOMAIN | 环境变量控制 Set-Cookie domain（生产 `.tokenjoy.com`） |
| 新增环境变量 | `VITE_WEB_ORIGIN`（apps/frontend），控制 postMessage targetOrigin |

---

## 部署层改动清单

| 改动 | 描述 |
|------|------|
| CSP frame-ancestors | 反向代理层对 `/auth/embed` 返回 `frame-ancestors https://www.tokenjoy.com` |
| 默认 frame-ancestors | 其他路径返回 `frame-ancestors 'none'` |
| 官网独立部署 | apps/web/ 独立构建、独立部署到 www.tokenjoy.com |

---

## 实施步骤

| 步骤 | 内容 |
|------|------|
| 1 | 从 AuthPopup 提取 AuthCard 组件（纯表单，无 Dialog），AuthPopup 改为 Dialog + AuthCard 组合 |
| 2 | App.tsx 新增 `/auth/embed` 路由，与 `/login` 平级注册（SessionGate 之外） |
| 3 | AuthEmbedPage 渲染 AuthCard，加载完发 `auth:ready`，认证成功发 `auth:success`（指定 targetOrigin） |
| 4 | 后端加 COOKIE_DOMAIN 配置 + CORS 白名单 |
| 5 | 部署配置：反向代理对 `/auth/embed` 添加 `frame-ancestors` CSP 头 |
| 6 | 复制 `tokenjoy-web` 到 `apps/web/`，包名改为 `@tokenjoy/web` |
| 7 | 升级 apps/web/ 核心依赖（与 monorepo 对齐），验证构建通过 |
| 8 | 官网 Navbar 集成：dialog.showModal() + iframe + postMessage 监听（含 origin 校验 + 超时处理） |
| 9 | 根 package.json 加 `start:web` / `build:web` / `lint:web` |
| 10 | 本地联调验证：官网弹窗登录 → auth:ready 握手 → 认证完成 → Cookie 写入 → 跳转 App 免登 |

---

## 不做的事

- 不做 `packages/auth-popup` 共享包 — iframe 方案更轻、维护成本更低
- 不做 SSR/SSG — 官网纯 SPA，SEO 用 prerender 服务
- 不统一 Tailwind 版本 — 官网 v3，App v4，视觉独立
- 不合并官网到 apps/frontend — 独立产品、独立部署、视觉完全不同
- 不在官网引入路由库 — 单页不需要
