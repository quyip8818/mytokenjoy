# SMS 前端样式对齐 Apps 设计文档

## 目标

将 sms/frontend 的布局与视觉风格对齐到 apps/frontend 的水准，统一品牌感。

---

## 现状对比

| 维度 | apps/frontend | sms/frontend | 差距 |
|------|--------------|--------------|------|
| **Sidebar 宽度** | 可折叠 `w-56` ↔ `w-14`，动画过渡 | 固定 `w-56`，无折叠 | 缺折叠能力 |
| **Sidebar Logo** | 显示 `logo.png`，折叠后变 toggle 按钮 | 纯文字 "SMS" | 品牌感弱 |
| **NavGroup 折叠** | 支持组折叠/展开 + localStorage 持久化 | 无折叠，列表平铺 | 缺少交互层 |
| **NavItem 选中态** | `bg-primary/8 text-primary` + icon 加粗 | 同（已对齐） | ✓ |
| **折叠态 Tooltip** | 使用 shadcn Tooltip 显示菜单名称 | 不存在 | 缺 |
| **Badge** | 支持 `badgeKey` 动态数字红点 | 不需要（业务无审批） | 跳过 |
| **Header 高度** | `h-14` | `h-14` | ✓ |
| **Header 左侧** | 显示当前页面标题 | 无内容 | 缺页面标题 |
| **Header 右侧** | CompanyChip + UserAvatar + 通知铃铛 | 文字用户名 + 退出按钮 | 需要升级 |
| **主内容区** | `p-8` + `overflow-auto` | `p-8` + `overflow-auto` + `bg-muted/30` | 接近 |
| **Design Tokens** | 完整 indigo 品牌色系 + 自定义 shadow + sidebar token + dark mode | 默认 shadcn neutral hsl 色 | 色彩体系差距大 |
| **字体** | Inter Variable | 系统字体 | 不一致 |
| **Provider 层** | SidebarLayoutProvider + WorkflowProvider | 无 layout provider | 缺折叠状态管理 |

---

## 对齐方案

### 1. Design Tokens — 替换 `sms/frontend/src/index.css`

将 sms 的 CSS 变量系统替换为 apps 同款。关键改动：

```css
@import 'tailwindcss';
@import 'tw-animate-css';
@import 'shadcn/tailwind.css';
@import '@fontsource-variable/inter';

@custom-variant dark (&:is(.dark *));

@theme inline {
  --font-heading: 'Inter Variable', sans-serif;
  --font-sans: 'Inter Variable', sans-serif;
  /* 复制 apps/frontend/src/index.css 完整 @theme inline 块 */
  /* 包括 sidebar tokens、shadow tokens */
}
```

直接复制 apps 的 `:root` 和 `.dark` 变量。primary 色保持 `#635bff` 统一品牌。

### 2. Sidebar 折叠能力

从 apps 移植以下文件（适配 sms 业务路由）：

| 新增/修改文件 | 来源 |
|---|---|
| `sms/frontend/src/components/layout/sidebar-layout-constants.ts` | 新建，同 apps |
| `sms/frontend/src/components/layout/sidebar-layout-context.ts` | 新建，同 apps |
| `sms/frontend/src/components/layout/sidebar-layout-provider.tsx` | 新建，同 apps |
| `sms/frontend/src/components/layout/use-sidebar-layout.ts` | 新建，同 apps |
| `sms/frontend/src/components/layout/sidebar.tsx` | 重写 |

Sidebar 重写要点：
- `SidebarHeader` 组件：展开时显示 "SMS" logo（或文字），折叠后只显示 toggle 按钮
- `SidebarGroup` 组件：支持组标题点击折叠 + ChevronDown 动画
- `SidebarNavItem` 组件：折叠态显示 icon + Tooltip
- 宽度 transition: `transition-[width] duration-200 ease-in-out`
- localStorage key 改为 `sms.sidebar.collapsed` 避免冲突

### 3. Header 升级

改造 `sms/frontend/src/components/layout/header.tsx`：

```
┌─────────────────────────────────────────────────────────────┐
│  [当前页面标题]                         [用户名] [退出]      │
└─────────────────────────────────────────────────────────────┘
```

具体改动：
- 左侧增加 `h1` 显示当前路由标题（从 `ROUTE_DEFINITIONS` 派生 `ROUTE_TITLES` map）
- 右侧用户区改为 chip 样式（`rounded-md border px-2.5 py-1.5`）
- 退出按钮保留，样式对齐 apps 的 ghost button 风格

**不需要**：CompanyChip（SMS 无多公司概念）、通知铃铛（SMS 无通知系统）。

### 4. AdminLayout 包裹 Provider

```tsx
export function AdminLayout() {
  return (
    <SidebarLayoutProvider>
      <div className="flex h-screen bg-background">
        <Sidebar />
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
          <Header />
          <main className="flex min-h-0 flex-1 flex-col overflow-hidden p-8">
            <div className="min-h-0 flex-1 overflow-auto">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
      <Toaster theme="light" />
    </SidebarLayoutProvider>
  )
}
```

### 5. 路由配置增加 `navGroupCollapsed`

给 sms 的 `ROUTE_DEFINITIONS` 增加 `navGroupCollapsed` 字段（"系统设置" 默认折叠）。

### 6. 字体安装

```bash
pnpm add @fontsource-variable/inter -w --filter sms-frontend
```

---

## 不做的事

| 跳过 | 原因 |
|------|------|
| Dark mode 切换按钮 | SMS 内部系统，浅色够用 |
| Badge 红点 | SMS 无审批流 |
| 通知系统 | SMS 无通知需求 |
| WorkflowProvider | SMS 无 workflow 面板 |
| TrialBanner | SMS 无试用概念 |
| 权限系统对齐 | SMS 用简单 role-based，无需 permission key 体系 |

---

## 实施顺序

1. **index.css** — 替换 tokens + 安装字体（立即可见效果）
2. **sidebar-layout-provider 系列** — 新建 4 个文件
3. **sidebar.tsx 重写** — 折叠 + group collapse
4. **header.tsx 改造** — 页面标题 + chip 样式
5. **admin-layout.tsx** — 包裹 Provider + 调整 main 结构

预计改动 ~8 个文件，新建 4 个文件。
