# SMS 前端 shadcn UI 迁移计划

> 破坏性迁移。一步到位对齐 apps 的 shadcn 组件体系。

## 现状

### apps/frontend（目标参考）

- shadcn CLI 配置 (`components.json`) ✅
- 35 个 UI 组件（`components/ui/`）
- 样式：Tailwind v4 + CSS 变量 + `cn()` 工具函数
- 无手写原生 HTML 表单/弹窗
- 无 barrel export（各组件单独 import）

### sms/frontend（当前）

- 无 `components.json`（shadcn CLI 无法工作）
- 11 个 UI 组件（手写，部分对齐 shadcn API）
- CSS tokens 已配好（index.css 有完整 shadcn 变量）
- `lib/utils.ts` 有 `cn()` ✅
- 页面大量使用原生 HTML：`<input className="input">`、`<select>`、手写 modal `<div>`
- 总计 ~2500 行页面代码需要迁移（含 detail.tsx 和 system/ 页面）

## 迁移步骤

### 1. 添加 components.json

直接从 apps/frontend 复制，保持完全一致：

```json
{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "base-nova",
  "rsc": false,
  "tsx": true,
  "tailwind": {
    "config": "",
    "css": "src/index.css",
    "baseColor": "neutral",
    "cssVariables": true,
    "prefix": ""
  },
  "iconLibrary": "lucide",
  "rtl": false,
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils",
    "ui": "@/components/ui",
    "lib": "@/lib",
    "hooks": "@/hooks"
  },
  "menuColor": "default",
  "menuAccent": "subtle",
  "registries": {}
}
```

### 2. 添加缺失的 shadcn 组件

从 apps 复制或用 CLI 生成，sms 需要的组件：

| 组件 | sms 已有 | apps 有 | 动作 |
|------|----------|---------|------|
| button | ✅ | ✅ | 保留 |
| input | ✅ | ✅ | 保留 |
| dialog | ✅ | ✅ | 保留 |
| badge | ✅ (含 StatusBadge) | ✅ | 保留 |
| select | ✅ (原生 wrapper) | ✅ (radix) | **替换为 radix select** |
| pagination | ✅ | ✅ | 保留（业务组件） |
| tooltip | ✅ | ✅ | 保留 |
| password-input | ✅ | ✅ | 保留 |
| table | ❌ | ✅ | **添加** |
| card | ❌ | ✅ | **添加** |
| label | ❌ | ✅ | **添加** |
| textarea | ❌ | ✅ | **添加** |
| tabs | ❌ | ✅ | **添加** |
| separator | ❌ | ✅ | **添加** |
| sheet (侧抽屉) | ❌ | ✅ | **添加** |
| dropdown-menu | ❌ | ✅ | **添加** |
| skeleton | ❌ | ✅ | **添加** |
| alert-dialog | ❌ | ✅ | **添加** |
| progress | ❌ | ✅ | **添加** |
| confirm-action-dialog | ❌ | ✅ | **添加（封装 AlertDialog）** |

添加命令（在 `sms/frontend/` 下）：
```bash
pnpm dlx shadcn@latest add table card label textarea tabs separator sheet dropdown-menu skeleton alert-dialog progress
```

`confirm-action-dialog` 从 apps 复制（业务封装组件）。

### 3. 删除手写组件 / 替换为 shadcn 版本

| 当前文件 | 动作 |
|----------|------|
| `components/ui/field.tsx` | 删除，用 `<Label>` + 直接 `<Input>` |
| `components/ui/select.tsx` | 删除，用 shadcn `<Select>` (radix) |
| `components/ui/empty-state.tsx` | 保留（业务组件，apps 也有） |
| `components/ui/badge.tsx` / `StatusBadge` | 保留（业务组件） |
| `components/ui/pagination.tsx` | 保留（业务组件） |
| `components/ui/password-input.tsx` | 保留（apps 也有同名组件） |

### 4. 页面迁移（从简到复杂）

#### 4.1 dashboard-page.tsx (~140 行)

| 当前 | 替换为 |
|------|--------|
| `<div className="rounded-lg border bg-white p-5">` | `<Card>` + `<CardContent>` |
| loading 时无占位 | `<Skeleton>` |

#### 4.2 system/weights.tsx (~90 行，直接写在 route 文件)

**先迁移到 `features/system/weights-page.tsx`**（route 文件只做导出），然后：

| 当前 | 替换为 |
|------|--------|
| `<button>` | `<Button>` |
| `<input type="range">` + `<input type="number">` | 保留原生（range+number 联动比 radix Slider 更直接） |
| 手写 loading | `<Skeleton>` |
| card div | `<Card>` |

#### 4.3 evaluations-page.tsx (~260 行)

| 当前 | 替换为 |
|------|--------|
| `<table>` 手写 | `<Table>` 组件 |
| `<input>` / `<select>` | `<Input>` / `<Select>` |
| `<button>` | `<Button>` |
| 手写 modal div | `<Dialog>` / `<DialogContent>` |
| `confirm()` | `<ConfirmActionDialog>` |
| `<input type="range">` + number 联动 | 保留原生（比 radix Slider 更适合精确数值输入） |
| 右侧预览面板 | `<Card>` |

#### 4.4 suppliers-page.tsx (~210 行)

| 当前 | 替换为 |
|------|--------|
| `<table>` 手写 | `<Table>` 组件 |
| `<input className="input">` | `<Input>` |
| `<select>` | `<Select>` |
| `<button onClick={openCreate}>` | `<Button>` |
| 手写 modal div | `<Dialog>` / `<DialogContent>` |
| `confirm()` | `<ConfirmActionDialog>` |

#### 4.5 orders-page.tsx (~210 行)

同 suppliers 模式。

#### 4.6 contracts-page.tsx (~320 行)

| 当前 | 替换为 |
|------|--------|
| 同上表格+表单模式 | 同上 |
| 详情侧抽屉（手写 div） | `<Sheet>` / `<SheetContent>` |
| 附件区域 | `<Card>` 包裹 |
| `confirm()` | `<ConfirmActionDialog>` |

#### 4.7 suppliers/detail.tsx (~310 行)

| 当前 | 替换为 |
|------|--------|
| 手写 Tab 按钮 | shadcn `<Tabs>` + `<TabsList>` + `<TabsContent>` |
| 5 个 tab 内表格 | `<Table>` |
| 联系人弹窗 | `<Dialog>` |
| 信息头 div | `<Card>` |
| `confirm()` | `<ConfirmActionDialog>` |

#### 4.8 users-page.tsx (~190 行)

| 当前 | 替换为 |
|------|--------|
| `<table>` | `<Table>` |
| `<button>` | `<Button>` |
| `<input>` / `<select>` | `<Input>` / `<Select>` |
| 手写 modal div | `<Dialog>` |
| `confirm()` | `<ConfirmActionDialog>` |

### 5. 删除 barrel export

`components/ui/index.ts` 删除。所有页面改为直接 import：
```ts
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
```

对齐 apps 的风格（apps 无 barrel export）。

### 6. 更新 components/layout/

保持不变（sidebar、header、page-shell 是 sms 自己的 layout，不需要和 apps 对齐）。

## 规则（对齐 apps 约定）

1. `components/ui/` — 只放 shadcn 组件和无业务语义的原子 UI
2. 页面组件不直接用原生 HTML 表单元素（`<input>`、`<select>`、`<button>`）
3. 弹窗必须用 `<Dialog>` 或 `<AlertDialog>`，禁止 `window.confirm()`
4. 表格必须用 `<Table>` 组件
5. 所有样式通过 `cn()` 组合，不用 `className="input"` 等自定义 class
6. 无 barrel export，每个组件单独 import

## 预估

| 步骤 | 工作量 |
|------|--------|
| 1-2. 配置 + 添加组件 | 5 分钟 |
| 3. 删除旧组件 | 5 分钟 |
| 4.1 dashboard | 10 分钟 |
| 4.2 weights | 15 分钟 |
| 4.3 evaluations | 25 分钟 |
| 4.4 suppliers | 20 分钟 |
| 4.5 orders | 20 分钟 |
| 4.6 contracts | 25 分钟 |
| 4.7 detail | 30 分钟 |
| 4.8 users | 15 分钟 |
| 5. 清理 barrel export | 5 分钟 |
| **总计** | **~3 小时** |

## 验证

```bash
cd sms/frontend
pnpm build      # TypeScript 编译通过
pnpm lint       # eslint 通过
```
