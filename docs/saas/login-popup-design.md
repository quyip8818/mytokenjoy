# AuthPopup 跨域认证方案

> AuthPopup 是独立的认证流程组件，可嵌入 App、官网、任何子域。  
> 官网和 App 分开部署在不同子域名，通过父域 Cookie 共享 session。

---

## 1. 部署拓扑

```
tokenjoy.com（父域）
├── www.tokenjoy.com    — 官网（静态站，Vite SSG / Next.js）
├── app.tokenjoy.com    — SaaS 管理后台（React SPA）
└── api.tokenjoy.com    — 后端 API（Go）
```

Cookie Domain 设为 `.tokenjoy.com`，所有子域共享 session。

---

## 2. 核心流程

### 2.1 从官网注册 → 进入 App

```
用户在 www.tokenjoy.com 点击 "免费试用"
  → 弹出 AuthPopup（mode=register）
  → SMS 验证 → 设置密码 + 公司名 → 提交
  → API (api.tokenjoy.com) Set-Cookie domain=.tokenjoy.com
  → onSuccess 回调：window.location.href = 'https://app.tokenjoy.com'
  → App 加载时 Cookie 已在 → 直接进入 Dashboard
```

### 2.2 从 App 登录（常规）

```
用户访问 app.tokenjoy.com
  → SessionGate 检测无 session → 渲染 Fake UI + 打开 AuthPopup
  → SMS 登录 → API Set-Cookie domain=.tokenjoy.com
  → onSuccess：refreshSession() → 进入 Dashboard
```

### 2.3 401 Session 过期

```
用户操作中 API 返回 401
  → refresh 失败 → 打开 AuthPopup（覆盖当前页）
  → 重新登录 → invalidateQueries() → 继续使用
```

---

## 3. Cookie 跨域方案

### 3.1 后端 Set-Cookie 配置

```go
http.SetCookie(w, &http.Cookie{
    Name:     "tokenjoy_session_member",
    Value:    accessToken,
    Domain:   cfg.CookieDomain,  // 生产: ".tokenjoy.com"
    Path:     "/",
    HttpOnly: true,
    Secure:   true,              // HTTPS only
    SameSite: http.SameSiteLaxMode,
    MaxAge:   sessionTTLSec,
})
```

环境变量：

| 变量 | 开发环境 | 生产环境 |
|------|---------|---------|
| `COOKIE_DOMAIN` | （空，默认 localhost） | `.tokenjoy.com` |

### 3.2 CORS 配置

```go
cors.Options{
    AllowedOrigins:   []string{
        "https://www.tokenjoy.com",
        "https://app.tokenjoy.com",
    },
    AllowCredentials: true,
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
}
```

`AllowCredentials: true` 允许跨域请求携带/设置 Cookie。

### 3.3 前端 fetch 配置

```typescript
// 官网和 App 的 API client 都需要
fetch(url, {
  credentials: 'include',  // ← 关键：跨域带 Cookie
  headers: { 'Content-Type': 'application/json' },
})
```

---

## 4. AuthPopup 组件设计

### 4.1 组件接口

```typescript
interface AuthPopupProps {
  open: boolean
  defaultMode?: 'login' | 'register'
  apiBase?: string                  // 默认 '/api'，官网传 'https://api.tokenjoy.com'
  closable?: boolean                // false = 不可关闭（App 内未登录时）
  onSuccess?: () => void            // 认证成功回调
  onClose?: () => void              // 关闭回调
}
```

### 4.2 内部状态机

```
┌─────────────────────────────────────────────┐
│                                             │
│  Tab: [登录] [注册]                          │
│                                             │
│  登录 Tab:                                  │
│    phone_verify → enter (成功)              │
│                → select_company (多企业)    │
│                → choose (有邀请)            │
│                → not_found (提示切注册)     │
│                                             │
│  注册 Tab:                                  │
│    phone_verify → info_step (密码+公司名)   │
│                → already_registered (提示)  │
│    info_step   → success (创建完成)         │
│                                             │
│  企业选择 / 邀请选择:                        │
│    select → success                         │
│                                             │
└─────────────────────────────────────────────┘
```

### 4.3 视觉规格

- 容器：`glass-card rounded-2xl` + `shadow-[0_10px_50px_rgba(139,92,246,0.12)]`
- 宽度：`max-w-md`（448px）
- 遮罩：`bg-black/20 backdrop-blur-sm`（让背景可见但不抢焦点）
- 品牌色：`brand-500 #8B5CF6`（按钮、Tab active）
- Tab 切换动画：下划线滑动 + 内容 fade

### 4.4 Popup 卡片 Wireframe

```
┌──────────────────────────────────────┐
│                                      │
│        TokenJoy                      │
│    企业 AI 管理平台                  │
│                                      │
│   ┌────────┐  ┌────────┐            │
│   │  登录  │  │  注册  │            │
│   └════════┘  └────────┘            │
│                                      │
│   手机号                             │
│   [+86] [_______________]            │
│                                      │
│   验证码                             │
│   [________]  [获取验证码]           │
│                                      │
│   [═══════ 登录 ═══════]             │
│                                      │
└──────────────────────────────────────┘
```

---

## 5. 官网集成

### 5.1 作为 npm 包（官网也是 React）

```tsx
// 官网 tokenjoy-web 代码
import { AuthPopup } from '@tokenjoy/auth-popup'

function HeroCTA() {
  const [open, setOpen] = useState(false)

  return (
    <>
      <button onClick={() => setOpen(true)}>免费试用</button>
      <AuthPopup
        open={open}
        defaultMode="register"
        apiBase="https://api.tokenjoy.com"
        closable={true}
        onSuccess={() => {
          window.location.href = 'https://app.tokenjoy.com'
        }}
        onClose={() => setOpen(false)}
      />
    </>
  )
}
```

### 5.2 npm 包结构（packages/auth-popup）

```
packages/auth-popup/
├── package.json
├── src/
│   ├── index.ts                    — 导出 AuthPopup + types
│   ├── auth-popup.tsx              — 主组件
│   ├── internal/
│   │   ├── login-tab.tsx
│   │   ├── register-phone-step.tsx
│   │   ├── register-info-step.tsx
│   │   ├── company-select-step.tsx
│   │   ├── invite-select-step.tsx
│   │   └── sms-countdown.ts
│   └── api-client.ts              — 轻量 fetch wrapper（不依赖 App 的 client）
└── tsconfig.json
```

关键：这个包**自带 API client**（纯 fetch + credentials:include），不依赖 App 的 React Query / api 层。

### 5.3 作为独立 JS bundle（官网非 React 时的备选）

```html
<script src="https://app.tokenjoy.com/auth-popup.umd.js"></script>
<script>
  TokenJoyAuth.open({
    mode: 'register',
    apiBase: 'https://api.tokenjoy.com',
    onSuccess: () => location.href = 'https://app.tokenjoy.com'
  })
</script>
```

构建为 UMD bundle（Vite library mode），挂载到 `window.TokenJoyAuth`。

---

## 6. App 内集成

### 6.1 全局 Context

```typescript
// features/auth/auth-popup-context.tsx
const AuthPopupContext = createContext<AuthPopupControl>(...)

// App 内任何地方
const { open } = useAuthPopup()
open('login')
```

### 6.2 SessionGate 改造

```tsx
function SessionGate({ children }) {
  const { session, loading } = useSession()
  const { open, isOpen } = useAuthPopup()

  if (loading) return <RouteFallback />
  if (!session) {
    if (!isOpen) open('login')
    return <FakeDashboardBackground />  // 静态背景，不是 children
  }
  return children
}
```

### 6.3 401 拦截

```typescript
// api/client.ts
if (response.status === 401 && refreshFailed) {
  authPopupControl.open('login')
}

// 登录成功后
onSuccess: () => {
  queryClient.invalidateQueries()  // 全量重新 fetch
}
```

---

## 7. `/login` 路由页（Fake UI 背景）

App 内的 `/login` 路由是一个纯静态视觉页面，AuthPopup 浮在上面。

### 7.1 结构

```tsx
export default function LoginPage() {
  return (
    <div className="relative min-h-screen overflow-hidden hero-ambient">
      {/* 浮动光斑 */}
      <FloatingBlobs />
      {/* 网格 */}
      <div className="absolute inset-0 bg-light-grid opacity-50" />
      {/* 品牌栏 */}
      <BrandHeader />
      {/* 假 Dashboard 卡片（模糊） */}
      <div className="relative mx-auto mt-12 max-w-5xl px-8 opacity-50 blur-[2px]">
        <FakeDashboardCards />
      </div>
      {/* AuthPopup */}
      <AuthPopup open={true} defaultMode="login" closable={false} />
    </div>
  )
}
```

### 7.2 FakeDashboardCards

```
┌──────────┐ ┌──────────┐ ┌──────────┐
│ 模型用量 │ │ 预算余额 │ │ 活跃 Key │
│  12,580  │ │ ¥12,580  │ │   23 个  │
│  ▇▅▇▆▇▇▅ │ │ ████░░░░ │ │  4 项目  │
└──────────┘ └──────────┘ └──────────┘
┌──────────────────────────────────────┐
│ 今日调用  1,247 次  +12% ↑           │
│ ▁▂▃▅▆▇▆▅▃▄▅▇█▇▆▅▄▃▂▃▄▅▆           │
└──────────────────────────────────────┘
```

纯 HTML/CSS 渲染，glass-card 风格，零 API 依赖。

---

## 8. 安全

| 项 | 措施 |
|---|---|
| Cookie Domain | `.tokenjoy.com` — 仅自己的子域可读 |
| SameSite | `Lax` — 顶级导航带 Cookie，第三方 POST 不带 |
| Secure | `true` — 仅 HTTPS |
| CORS 白名单 | 仅允许 `www.tokenjoy.com` + `app.tokenjoy.com` |
| credentials | `include` — 跨域请求带 Cookie |
| Open Redirect 防护 | `onSuccess` 由调用方硬编码，不从 URL 读取 |
| Protected 树隔离 | 未认证时 protected 组件不 mount、chunk 不加载 |

---

## 9. 本地开发

| 场景 | 配置 |
|------|------|
| App 单独开发 | `COOKIE_DOMAIN` 留空（默认 localhost）; AuthPopup 内嵌在 App |
| 官网 + App 联调 | 本地 hosts: `127.0.0.1 www.local.tokenjoy.com app.local.tokenjoy.com`; `COOKIE_DOMAIN=.local.tokenjoy.com` |
| API base | App: `/api`（proxy）; 官网: `http://localhost:8080` 或上述子域 |

---

## 10. 实施步骤

| 阶段 | 内容 | 影响范围 |
|------|------|---------|
| P0 | 后端：`COOKIE_DOMAIN` 环境变量 + Set-Cookie domain | backend config |
| P0 | 后端：CORS 配置（允许官网 origin + credentials） | backend middleware |
| P0 | `packages/auth-popup` 包骨架 + AuthPopup 组件 | 新 package |
| P0 | App `login.tsx` 改为 Fake UI 背景 + AuthPopup | routes/auth/login.tsx |
| P1 | AuthPopup 内部：login tab + register 两步 | auth-popup 包 |
| P1 | App SessionGate 改造（popup + fake background） | features/session |
| P1 | App 401 拦截器集成 | api/client.ts |
| P2 | 官网集成：CTA → AuthPopup → 跳转 App | tokenjoy-web |
| P2 | 独立 JS bundle 构建（UMD，备选） | vite library mode |
| P3 | 删除旧 `/register`、`/onboard` 独立路由 | cleanup |

---

## 11. 设计约束

| 约束 | 说明 |
|------|------|
| AuthPopup 零路由依赖 | 不调 navigate、不读 location，纯受控组件 |
| AuthPopup 自带 API client | 不依赖 App 的 axios/React Query，纯 fetch |
| 官网不需要完整 App 依赖 | 只引 `@tokenjoy/auth-popup` 一个包 |
| Cookie 只设父域 | 生产 `.tokenjoy.com`，开发留空 |
| Fake UI 纯静态 | 零 API、零 session、纯 CSS/HTML |
| onSuccess 硬编码目标 | 官网写死 `app.tokenjoy.com`，不从 URL 参数读（防 open redirect） |
