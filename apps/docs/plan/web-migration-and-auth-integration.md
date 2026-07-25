# 官网迁入 Monorepo + AuthPopup 集成

## 概述

将 `tokenjoy-web`（官网）迁入 monorepo 为 `web/`，通过 `packages/auth-popup` 共享认证弹窗组件，实现"官网点击登录 → 弹出 AuthPopup → 认证后跳转 SaaS 后台"的完整流程。

---

## 迁移后仓库结构

```
mytokenjoy/
├── web/                         ← 官网（marketing site）
│   ├── src/
│   │   ├── pages/               ← 页面（Home）
│   │   ├── sections/            ← 首页各 section 组件
│   │   ├── shared/              ← 官网共享组件（Navbar、Footer、Logo）
│   │   └── content/             ← 文案数据（CMS-like 结构化内容）
│   ├── package.json             ← @tokenjoy/web
│   ├── vite.config.ts
│   └── tailwind.config.js       ← 官网使用 Tailwind v3（独立于 apps）
│
├── packages/
│   ├── contracts/               ← 已有：permission codegen
│   └── auth-popup/              ← 新增：跨应用共享的认证弹窗组件
│       ├── src/
│       │   ├── index.ts         ← 导出 AuthPopup + types
│       │   ├── auth-popup.tsx   ← 主组件（Dialog + 状态机）
│       │   ├── api-client.ts    ← 轻量 fetch wrapper（自带，不依赖 App 的 client）
│       │   └── steps/           ← 各认证步骤组件
│       │       ├── login-phone-code.tsx
│       │       ├── login-email-pw.tsx
│       │       ├── register-phone.tsx
│       │       ├── register-info.tsx
│       │       ├── select-company.tsx
│       │       └── reset-password.tsx
│       ├── package.json         ← @tokenjoy/auth-popup
│       └── tsconfig.json
│
├── apps/                        ← 不变
├── sms/                         ← 不变
└── pnpm-workspace.yaml          ← 加 'web'
```

---

## packages/auth-popup 设计

### 定位

独立的认证 UI 组件包，可在以下场景使用：
- **官网**（`web/`）：CTA 按钮 → 弹出注册/登录 → 成功后跳转 App
- **App**（`apps/frontend/`）：SessionGate 未登录时 → 弹出登录 → 成功后进入系统
- **App 401**：session 过期 → 弹出重新登录 → 恢复

### 接口

```typescript
interface AuthPopupProps {
  open: boolean
  defaultMode?: 'login' | 'register'
  apiBase?: string               // 默认 '/api'；官网传绝对 URL 如 'https://api.tokenjoy.com'
  closable?: boolean             // false = 不可关闭（App SessionGate 强制登录）
  onSuccess?: () => void         // 认证成功回调
  onClose?: () => void           // 关闭回调（closable=true 时有效）
}
```

### 关键约束

| 约束 | 原因 |
|------|------|
| 零路由依赖 | 不调 navigate、不读 location，纯受控组件 |
| 自带 API client | 纯 fetch + `credentials: 'include'`，不依赖 App 的 React Query |
| 零 UI 库依赖 | 内部自带极简 Dialog/Input 或使用 Radix headless（peer dep） |
| peer dep: react | 消费方提供 React |

### 从 apps/frontend 提取

当前 `apps/frontend/src/features/auth/components/auth-popup.tsx` 已实现完整的认证状态机。迁移步骤：

1. 将 `auth-popup.tsx` + 相关 steps 移入 `packages/auth-popup/src/`
2. 替换对 `@/components/ui/*` 的依赖 → 包内自带或 peer dep Radix
3. 替换对 `@/api/auth` 的依赖 → 包内 `api-client.ts`（纯 fetch）
4. `apps/frontend` 改为 `import { AuthPopup } from '@tokenjoy/auth-popup'`

---

## 官网集成

### Navbar 改造

```tsx
// web/src/shared/Navbar.tsx（改造后）
import { useState } from 'react'
import { AuthPopup } from '@tokenjoy/auth-popup'

export function Navbar({ content }: NavbarProps) {
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login')

  const handleLogin = () => { setAuthMode('login'); setAuthOpen(true) }
  const handleCTA = () => { setAuthMode('register'); setAuthOpen(true) }

  return (
    <nav>
      {/* ... 现有导航 ... */}
      <button onClick={handleLogin}>{content.loginLabel}</button>
      <button onClick={handleCTA}>{content.ctaLabel}</button>

      <AuthPopup
        open={authOpen}
        defaultMode={authMode}
        apiBase="https://api.tokenjoy.com"
        closable={true}
        onSuccess={() => window.location.href = 'https://app.tokenjoy.com'}
        onClose={() => setAuthOpen(false)}
      />
    </nav>
  )
}
```

### 效果

用户在官网点击"登录"或"免费试用"：
1. 当前页面不跳转，背景不变
2. 弹出 AuthPopup（Dialog 遮罩 + 认证卡片）
3. 完成认证后，Cookie 写入 `.tokenjoy.com` 域
4. `onSuccess` 回调跳转到 `app.tokenjoy.com`
5. App 加载时检测到 Cookie，直接进入 Dashboard

---

## Cookie 跨域方案

```
tokenjoy.com（父域）
├── www.tokenjoy.com    — 官网（web/）
├── app.tokenjoy.com    — SaaS 管理后台（apps/frontend/）
└── api.tokenjoy.com    — 后端 API（apps/backend/）
```

| 配置项 | 开发环境 | 生产环境 |
|--------|---------|---------|
| Cookie Domain | （空，默认 localhost） | `.tokenjoy.com` |
| CORS Origins | `http://localhost:5175` | `https://www.tokenjoy.com, https://app.tokenjoy.com` |
| SameSite | Lax | Lax |
| Secure | false | true |

后端环境变量 `COOKIE_DOMAIN` 控制 Set-Cookie 的 domain 字段。

---

## 本地开发

### 端口分配

| 服务 | 端口 |
|------|------|
| 官网 (web/) | 5175 |
| Apps frontend | 5173 |
| Apps backend | 8010 |

### 开发流程

```bash
pnpm start web           # 启动官网 dev server (port 5175)
pnpm start               # 启动 apps（backend + frontend + mock）
```

官网本地联调时，AuthPopup 的 `apiBase` 指向 `http://localhost:8010`（通过 env 或 vite proxy）。

### pnpm-workspace.yaml

```yaml
packages:
  - 'apps/*'
  - 'sms/*'
  - 'web'
  - 'packages/*'
```

---

## App 端集成

### SessionGate 改造

`apps/frontend` 的 SessionGate 检测到无 session 时：
1. 渲染静态 Fake Dashboard 背景（已有，见 `login-popup-design.md`）
2. 打开 `<AuthPopup open={true} closable={false} />`
3. 认证成功 → `refreshSession()` → 进入系统

### 401 拦截

```typescript
// api/client.ts
if (response.status === 401 && refreshFailed) {
  authPopupControl.open('login')  // 弹出重新登录
}
```

---

## 实施步骤

| 步骤 | 内容 |
|------|------|
| 1 | 复制 `tokenjoy-web` 到 `web/`，包名改为 `@tokenjoy/web`，加入 workspace |
| 2 | 创建 `packages/auth-popup` 骨架，从 `apps/frontend/features/auth/` 提取核心组件 |
| 3 | `packages/auth-popup` 内实现独立 API client（fetch + credentials:include） |
| 4 | `apps/frontend` 改为消费 `@tokenjoy/auth-popup`（替换本地 auth-popup.tsx） |
| 5 | `web/` 的 Navbar 集成 AuthPopup（替换原来的 `<a href>` 跳转） |
| 6 | 后端加 `COOKIE_DOMAIN` + CORS 白名单配置 |
| 7 | 根 package.json 加 `start web` / `build:web` 命令 |
| 8 | 本地联调验证：官网登录 → Cookie 写入 → App 免登 |

---

## 不做的事

- 不做 SSR/SSG — 官网是纯客户端 SPA，搜索引擎用 prerender 服务处理
- 不统一 Tailwind 版本 — 官网保持 v3（视觉独立），apps 用 v4
- 不做 UMD bundle — 官网本身是 React，直接 import 包即可
- 不合并官网到 apps/frontend — 官网是独立产品，独立部署，视觉完全不同
