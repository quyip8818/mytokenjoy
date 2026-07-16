# Demo 流程实施方案

> **范围**：本文仅覆盖 Demo 试用流程的完整实现。正式注册（SaaS `/register`）、私有化 Setup（`/setup`）、成员邀请等见 [登录注册方案设计.md](./登录注册方案设计.md)（placeholder，后续补充）。

---

## 1. 产品概述

用户在登录页点击 **"免费试用"**，无需注册，直接进入一个独立的 Demo 环境体验完整平台。

### 核心体验

```
/login 页面
    │
    ├── [免费试用] 按钮
    │       │
    │       ▼
    │   选择初始化方式（二选一）
    │   ┌─────────────────────────────────────────┐
    │   │  ① 使用示例数据    ② 导入您的数据       │
    │   │     （立即体验）      （CSV 上传）        │
    │   └─────────────────────────────────────────┘
    │       │
    │       ▼
    │   进入完整平台（带 Demo Banner）
    │   右上角有 [删除 Demo] 按钮
    │
    └── [登录] （正式用户，placeholder）
```

---

## 2. 用户流程

### 2.1 入口：点击"免费试用"

```
┌────────────────────────────────────────┐
│            登录 TokenJoy               │
│                                        │
│  [邮箱/手机号登录 - placeholder]       │
│                                        │
│  ──────────── 或 ────────────          │
│                                        │
│  [✨ 免费试用]                          │
│                                        │
└────────────────────────────────────────┘
```

点击后进入 Demo 初始化选择页。**无需手机号/邮箱**（首版匿名 Demo，通过 Cookie 绑定）。

### 2.2 选择初始化方式

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│            选择如何开始体验                           │
│                                                      │
│   ┌─────────────────┐    ┌─────────────────┐        │
│   │                 │    │                 │        │
│   │   📊 示例数据    │    │   📁 导入数据    │        │
│   │                 │    │                 │        │
│   │  预设组织架构    │    │  上传 CSV 创建  │        │
│   │  预算 & 模型     │    │  您自己的组织   │        │
│   │  即刻体验全功能  │    │                 │        │
│   │                 │    │                 │        │
│   │  [立即开始]     │    │  [上传 CSV]     │        │
│   └─────────────────┘    └─────────────────┘        │
│                                                      │
│   ┌──────────────────────────────────────────┐      │
│   │  🔗 第三方同步                            │      │
│   │  ┌────────┐ ┌────────┐ ┌────────┐       │      │
│   │  │ 飞书   │ │ 钉钉   │ │ 企微   │       │      │
│   │  │ (灰色) │ │ (灰色) │ │ (灰色) │       │      │
│   │  └────────┘ └────────┘ └────────┘       │      │
│   │  Demo 模式下暂不支持第三方同步            │      │
│   └──────────────────────────────────────────┘      │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 2.3 选项 ① 使用示例数据

点击 "立即开始"：
1. 后端创建 Demo Company + 灌入 seed 数据
2. 签发 JWT Cookie
3. 直接跳转到 Dashboard

### 2.4 选项 ② 导入您的数据

点击 "上传 CSV"：

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│   ← 返回                                            │
│                                                      │
│   导入组织架构                                       │
│                                                      │
│   1. 下载 CSV 模板                                   │
│      [📥 下载模板]                                    │
│                                                      │
│   2. 填写后上传                                      │
│      ┌──────────────────────────────────┐            │
│      │  拖拽文件到此处 或 点击上传       │            │
│      └──────────────────────────────────┘            │
│                                                      │
│   ─────────────────────────────────                  │
│   预览（上传后显示）：                               │
│   ┌──────────────────────────────────────┐           │
│   │ 姓名    │ 邮箱          │ 部门       │           │
│   │ 张三    │ zhang@co.com  │ 技术部     │           │
│   │ 李四    │ li@co.com     │ 产品部     │           │
│   │ ...     │               │            │           │
│   └──────────────────────────────────────┘           │
│                                                      │
│   共 12 条记录，将创建 3 个部门                      │
│                                                      │
│   [开始体验]                                         │
│                                                      │
│   或者 [使用示例数据直接开始]                        │
│                                                      │
└──────────────────────────────────────────────────────┘
```

点击 "开始体验"：
1. 后端创建 Demo Company（空白，无 seed）
2. 执行 CSV 导入（部门 + 成员）
3. 灌入基础配置数据（模型列表、预算模板）
4. 签发 JWT Cookie
5. 跳转 Dashboard

**CSV 导入后的补充提示**（进入 Dashboard 后 Toast）：
> "已导入您的组织架构。平台含预设模型，预算待分配。可在设置页配置，或使用模拟消耗测试看板。"

### 2.5 CSV 模板

文件名：`tokenjoy-组织架构模板.csv`

```csv
姓名,邮箱,手机号,部门,角色
张三,zhang@example.com,13800001111,技术部,超级管理员
李四,li@example.com,13800002222,技术部/后端组,普通成员
王五,wang@example.com,13800003333,产品部,普通成员
```

规则：
- 部门用 `/` 表示层级（自动创建不存在的部门）
- 角色可选：超级管理员 / 组织管理员 / 普通成员（默认）
- 邮箱和手机号均为选填（Demo 下不发通知，无需真实联系方式）
- 第一行为表头，自动识别
- 角色映射：中文 → 内部名（"超级管理员" → `super_admin`、"组织管理员" → `org_admin`、"普通成员" → `member`）

### 2.6 进入平台后

Demo 环境与正式环境功能一致，但有以下标识：

**顶部 Banner**（纯提示，无 CTA）：
```
┌──────────────────────────────────────────────────────────────┐
│ 🎯 试用环境 · 数据将在 30 天无活动后自动清理                 │
└──────────────────────────────────────────────────────────────┘
```

**右上角菜单增加**：
```
┌────────────┐
│ 当前：Demo │
│ ────────── │
│ 删除 Demo  │  ← 红色文字
└────────────┘
```

**Demo ID 恢复提示**（创建成功后 Toast，仅出现一次）：
```
您的 Demo ID：demo-38571
记住此 ID，可在其他设备通过"恢复 Demo"找回您的数据。
```

### 2.7 删除 Demo

点击 "删除 Demo"：

```
┌────────────────────────────────────────┐
│                                        │
│   确定删除 Demo 环境？                 │
│                                        │
│   删除后所有数据将立即清除，            │
│   无法恢复。                           │
│                                        │
│   [取消]          [确认删除]           │
│                                        │
└────────────────────────────────────────┘
```

确认后：
1. 后端 CASCADE DELETE Demo 租户全部数据
2. 清除 Cookie
3. 跳转回 `/login`

---

## 3. 技术方案

### 3.1 Demo 身份绑定

**首版：Cookie 匿名绑定**（不需要手机号）

- 创建 Demo 时签发 JWT（与正式登录一致）
- 同时设置 `tokenjoy_demo_id` Cookie（company_id）
- 用户下次访问若 JWT 有效 → 直接进入
- JWT 过期但 `tokenjoy_demo_id` 存在 → 尝试重新签发（验证 company 仍存在）

**Demo ID 恢复**：
- 每个 Demo 的 slug 即为 Demo ID（如 `demo-38571`）
- `/login` 页面增加小入口："已有 Demo？[输入 Demo ID]"
- `POST /auth/demo/recover { demoSlug }` → 校验存在 + 签 JWT

> **为什么不要手机号**：Demo 的目的是降低试用门槛。要手机号会流失大量潜在用户。正式注册时再要。

### 3.2 Company ID 段

| ID 范围 | 用途 |
| --- | --- |
| `1 ~ 99` | 开发/测试（seed company_id=1） |
| `100 ~ 99_999` | Demo 租户 |
| `100_000+` | 正式企业（placeholder，后续实现） |

### 3.3 后端 API

| 方法 | 路径 | Body | 响应 | 说明 |
| --- | --- | --- | --- | --- |
| POST | `/auth/demo/create` | `{ mode: "seed" \| "csv", csvData?: string }` | `{ memberId, companyId, demoSlug }` | 创建 Demo |
| POST | `/auth/demo/recover` | `{ demoSlug }` | `{ memberId, companyId }` | 通过 Demo ID 恢复 |
| DELETE | `/auth/demo` | — | `void` | 删除当前 Demo（需 Session） |
| GET | `/auth/demo/csv-template` | — | CSV file | 下载模板 |

#### `POST /auth/demo/create`

```typescript
interface DemoCreateRequest {
  mode: "seed" | "csv"
  csvData?: string  // mode=csv 时，Base64 编码的 CSV 内容
}

interface DemoCreateResponse {
  memberId: string
  companyId: number
  companySlug: string   // 即 Demo ID，如 "demo-38571"
}
```

逻辑：
1. 检查 `DEMO_ENABLED`；检查 Demo 租户总数 < `DEMO_MAX_TENANTS`
2. 分配 company_id（Demo 段 100~99999）
3. 创建 Company（`is_demo=true`, slug=`demo-{id}`）
4. `mode=seed`：灌入完整 seed（组织 + 预算 + 模型 + 调用记录）
5. `mode=csv`：解析 CSV → 创建部门树 + 成员 + 灌入基础配置（模型列表）
6. 取超管 → 签发 JWT → Set-Cookie
7. 返回

**错误响应**：
- Demo 未启用 → `404`
- 上限已满 → `503 { message: "试用名额已满，请稍后再试" }`
- CSV 格式错误 → `422 { message: "CSV 解析失败：第 3 行缺少姓名字段" }`
- CSV 行数超限 → `422 { message: "最多支持 500 行" }`

#### `POST /auth/demo/recover`

```typescript
interface DemoRecoverRequest {
  demoSlug: string  // 如 "demo-38571"
}
```

逻辑：
1. 查 `companies` WHERE `slug = demoSlug AND is_demo = true`
2. 不存在 → `404 { message: "Demo 不存在或已过期" }`
3. 存在 → 取超管成员 → 签发 JWT → 更新 `last_active_at`
4. 返回

#### `DELETE /auth/demo`

逻辑：
1. 从 Session 取 company_id
2. 验证 `is_demo=true`（防止误删正式企业）
3. CASCADE DELETE 该 company 全部数据
4. 清除 Session Cookie
5. 返回 204

### 3.4 Seed 数据内容

`mode=seed` 灌入的数据（精简版，取自现有 snapshot）：

| 数据 | 内容 |
| --- | --- |
| 组织架构 | 3 部门（技术/产品/市场）+ 8 名成员 |
| 角色 | 5 个预设角色 |
| 预算 | 企业总预算 50000 + 部门分配 |
| 模型 | 6 个模型（GPT-4o, Claude, DeepSeek 等） |
| API Key | 2 个示例 Platform Key |
| 调用记录 | 近 7 天模拟消费数据（看板有东西看） |
| 超限策略 | 80%/90% 预警阈值 |

`mode=csv` 灌入的基础配置（无调用数据）：

| 数据 | 内容 |
| --- | --- |
| 组织架构 | 从 CSV 解析 |
| 角色 | 5 个预设角色 |
| 预算 | 企业总预算 10000（未分配） |
| 模型 | 6 个模型 |
| API Key | 无 |
| 调用记录 | 无 |

### 3.5 Demo 清理

River periodic job，每日执行：

```go
func CleanExpiredDemos(ctx context.Context, st store.Store) {
    // DELETE FROM companies
    // WHERE is_demo = true
    //   AND id BETWEEN 100 AND 99999
    //   AND last_active_at < NOW() - INTERVAL '30 days'
    // CASCADE
}
```

`last_active_at` 专用字段，仅在以下时机更新：
- Demo 创建时
- JWT 有效登录进入 Demo 时
- `/auth/demo/recover` 恢复时

不会被 migration 或其他 company 更新误触。

### 3.6 Gateway（Demo 下）

Demo 租户的 Gateway 调用走 `dev-mock-llm`：
- 后端根据 `company.is_demo` 判断
- 代理目标切换为 `DEMO_MOCK_LLM_URL`（默认 `http://127.0.0.1:8765`）
- 返回模拟 response，正常走 ingest 写 ledger（看板能看到数据）

### 3.7 Demo 下功能限制

| 功能 | Demo 行为 |
| --- | --- |
| API 调用（Gateway） | Mock LLM，正常扣费记录，看板可见 |
| 预算 / 模型 / Key 管理 | 全功能 |
| 组织架构 CRUD | 全功能 |
| 审计列表 / 导出 | 全功能 |
| 看板 | 全功能（seed 模式有数据，csv 模式需模拟调用后可见） |
| 充值 / 钱包 | 显示固定余额，充值按钮禁用并提示"试用环境不支持充值" |
| 邀请成员 | 可发起但不真实投递（Toast 提示"试用环境不发送通知"） |
| 数据源配置（飞书/钉钉/企微） | 入口可见但表单禁用，提示"试用环境暂不支持第三方同步" |
| 通知（邮件/短信） | 不投递；站内通知正常 |

---

## 4. 前端实现

### 4.1 新增路由

| 路由 | 组件 | 说明 |
| --- | --- | --- |
| `/demo/setup` | DemoSetupPage | 选择初始化方式 |

### 4.2 文件结构

```
features/demo/
├── index.ts
├── hooks/
│   ├── use-demo-setup-page.ts    — 初始化选择页逻辑
│   ├── use-demo-session.ts       — Demo 状态检测
│   └── use-demo-recover.ts       — Demo ID 恢复逻辑
├── components/
│   ├── demo-setup-page-shell.tsx — 选择页面 UI
│   ├── demo-csv-upload.tsx       — CSV 上传 + 预览
│   ├── demo-banner.tsx           — 顶部 Banner
│   ├── demo-delete-dialog.tsx    — 删除确认弹窗
│   └── demo-recover-dialog.tsx   — Demo ID 恢复弹窗
└── lib/
    └── csv-parser.ts             — CSV 前端解析

api/demo.ts                       — demoApi（create/recover/delete/template）
```

### 4.3 Demo Banner 集成

在 `AdminLayout` 中：

```tsx
{session.isDemo && <DemoBanner />}
```

### 4.4 右上角菜单

在 Header 用户菜单中：

```tsx
{session.isDemo && (
  <DropdownMenuItem variant="destructive" onClick={openDeleteDialog}>
    删除 Demo
  </DropdownMenuItem>
)}
```

### 4.5 登录页改造

`/login` 页面增加 "免费试用" 按钮和 Demo 恢复入口：

```tsx
<Button variant="outline" onClick={() => navigate('/demo/setup')}>
  ✨ 免费试用
</Button>

<p className="text-sm text-muted-foreground">
  已有 Demo？<Button variant="link" onClick={openRecoverDialog}>输入 Demo ID 恢复</Button>
</p>
```

正式登录表单暂时保留现有邮箱密码（placeholder，后续改手机号）。

---

## 5. 数据库变更

```sql
-- companies 表新增列
ALTER TABLE companies ADD COLUMN is_demo BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE companies ADD COLUMN last_active_at TIMESTAMPTZ;  -- Demo 专用活跃标记

-- 清理用索引
CREATE INDEX idx_companies_demo_cleanup
    ON companies(last_active_at)
    WHERE is_demo = TRUE;
```

无需 `demo_tenants` 映射表（首版匿名 Cookie 绑定，不绑手机号；通过 Demo slug 恢复）。

---

## 6. 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `DEMO_ENABLED` | `false` | 是否开放 Demo 入口 |
| `DEMO_TTL_DAYS` | `30` | 无活动清理天数 |
| `DEMO_MAX_TENANTS` | `1000` | 同时存在上限 |
| `DEMO_MOCK_LLM_URL` | `http://127.0.0.1:8765` | Demo Gateway 代理目标 |

前端：
| 变量 | 说明 |
| --- | --- |
| `VITE_DEMO_ENABLED` | 控制 "免费试用" 按钮显示 |

---

## 7. Session 响应变更

`GET /session` 增加字段：

```typescript
interface SessionContext {
  // ... 现有字段
  isDemo: boolean          // 当前企业是否为 Demo
}
```

后端从 `companies.is_demo` 读取。

---

## 8. 安全考量

| 风险 | 缓解 |
| --- | --- |
| Demo 被批量创建 | IP 限流（5 次/小时/IP）；`DEMO_MAX_TENANTS` 上限 |
| Demo DELETE 误操作 | 二次确认弹窗 + 只允许 `is_demo=true` 的 company |
| Demo 中上传恶意 CSV | 限制文件大小（1MB）；行数上限（500）；严格列解析；错误逐行报告 |
| Demo 数据泄露 | Demo 数据全为模拟/用户自传；无真实企业数据 |
| Cookie 伪造访问他人 Demo | JWT 签名校验；company_id 在 JWT claims 中 |
| Demo ID 暴力猜测 | slug 为 `demo-{5位随机数字}`，组合空间 100k；可选加字母增大空间 |
| Demo recover 被滥用 | 恢复操作也受 IP 限流；仅返回 JWT 不暴露数据 |

---

## 9. 实施任务清单

### 9.1 后端

- [ ] DB Migration：`companies.is_demo` + `last_active_at` 列
- [ ] `POST /auth/demo/create` handler（seed + csv 两种 mode）
- [ ] `POST /auth/demo/recover` handler（Demo ID 恢复）
- [ ] `DELETE /auth/demo` handler
- [ ] `GET /auth/demo/csv-template` — 返回模板文件
- [ ] CSV 解析逻辑（部门层级自动创建 + 成员插入 + 中文角色映射）
- [ ] Demo seed 函数（复用现有 snapshot，精简）
- [ ] Gateway 对 `is_demo` 租户路由到 mock LLM
- [ ] Session handler 增加 `isDemo` 字段
- [ ] Demo 清理 River job（每日，基于 `last_active_at`）
- [ ] `DEMO_ENABLED` 配置守卫（未启用时 404）
- [ ] Demo 创建 IP 限流中间件

### 9.2 前端

- [ ] `api/demo.ts` — demoApi（create / recover / delete / template）
- [ ] `features/demo/` 完整 feature
- [ ] `/demo/setup` 路由 + 页面（选择方式 + CSV 上传 + 返回按钮）
- [ ] CSV 前端解析 + 预览表格 + 错误提示
- [ ] CSV 模板下载
- [ ] Demo Banner 组件（纯提示，无 CTA）
- [ ] Header 菜单增加"删除 Demo"
- [ ] 删除确认 Dialog
- [ ] Demo ID 创建成功 Toast
- [ ] `/login` 页增加 "免费试用" 按钮 + "恢复 Demo" 入口
- [ ] Demo 恢复 Dialog（输入 Demo ID）
- [ ] `SessionContext` 类型增加 `isDemo`
- [ ] Demo 下功能限制 UI（充值禁用、邀请 Toast、数据源灰色）

### 9.3 其他

- [ ] `dev-mock-llm` 确认可独立部署（Demo 生产环境）
- [ ] Demo 限流中间件

---

## 10. Placeholder：后续实现

以下功能在本次**不实现**，仅记录：

| 功能 | 说明 | 文档 |
| --- | --- | --- |
| SaaS 公开注册（`/register`） | 手机号 + 验证码 → 创建企业 | [登录注册方案设计.md](./登录注册方案设计.md) |
| 私有化 Setup（`/setup`） | 首次部署一次性 Wizard | 同上 |
| 手机号 + 验证码登录 | 替代邮箱密码为主认证 | 同上 |
| 邀请激活页（`/invite/accept`） | 成员自助设密码 | 同上 |
| Onboarding Workflow（正式版） | 正式注册后引导 | 同上 |
| Demo → 正式迁移 | 将 Demo 数据升级为正式企业 | 远期 |
| 手机号绑定 Demo | 替代 Cookie 匿名绑定 | 远期 |
| 忘记密码 / 重置密码 | P2 | 同上 |
