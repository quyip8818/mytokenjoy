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

**纯 JWT Cookie 绑定**（不需要手机号、不需要 slug）

- 创建 Demo 时签发 JWT（与正式登录一致，claims 含 `company_id`）
- 用户下次访问若 JWT 有效 → 直接进入对应 Demo
- JWT 过期 → 回到登录页，需重新创建 Demo
- 清除 Cookie / 换设备 → Demo 仍存在（30 天未活动才清理），但用户无法找回

> **为什么不做恢复**：Demo 目的是降低试用门槛，快速体验。丢了重建一个即可，数据不珍贵。正式使用时走注册流程。

### 3.2 Company 类型

`companies` 表新增 `type` 列（枚举），详见 [Company 租户模型设计](./Company租户模型设计.md)：

| 值 | 含义 | 部署形态 | 生命周期 |
| --- | --- | --- | --- |
| `standard` | SaaS 正式付费客户 | SaaS | 永久 |
| `trial` | SaaS 注册试用（有账号，限时） | SaaS | 到期降级/冻结 |
| `demo` | 匿名体验（无账号） | SaaS | 30 天无活动清理 |
| `selfhosted` | 私有化部署企业 | 私有化（`SupportSaas=false`） | 永久，单实例唯一 |
| `testing` | 开发/CI 测试 | 开发环境 | 不清理 |

本次实现 `demo`。`standard` / `trial` 为 placeholder，`selfhosted` / `testing` 对应现有 company。

Company ID 分配方式：复用现有 `CreateCompany` 的应用层 ID 分配逻辑。

Demo 数量上限通过 `SELECT COUNT(*) FROM companies WHERE type = 'demo'` 检查。

### 3.3 后端 API

| 方法 | 路径 | Body | 响应 | 说明 |
| --- | --- | --- | --- | --- |
| POST | `/auth/demo/create` | `{ mode: "seed" \| "csv", csvData?: string }` | `{ memberId, companyId }` | 创建 Demo |
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
}
```

逻辑：
1. 检查 `DEMO_ENABLED`；检查 Demo 租户总数 < `DEMO_MAX_TENANTS`（`SELECT COUNT(*) WHERE type = 'demo'`，加 advisory lock 防并发超发）
2. 创建 Company（`type='demo'`）
3. `mode=seed`：灌入完整 seed（组织 + 预算 + 模型 + 调用记录）
4. `mode=csv`：解析 CSV → 创建部门树 + 成员 + 灌入基础配置（模型列表）
5. 取超管 → 签发 JWT → Set-Cookie
6. 返回

**错误响应**：
- Demo 未启用 → `404`
- 上限已满 → `503 { message: "试用名额已满，请稍后再试" }`
- CSV 格式错误 → `422 { message: "CSV 解析失败：第 3 行缺少姓名字段" }`
- CSV 行数超限 → `422 { message: "最多支持 500 行" }`

#### `DELETE /auth/demo`

逻辑：
1. 从 Session 取 company_id
2. 验证 `type='demo'`（防止误删正式企业）
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
    // 分批删除，每批最多 10 个，避免长事务锁表
    // 通过 members 表的最近登录时间判断活跃度：
    // SELECT c.id FROM companies c
    // WHERE c.type = 'demo'
    //   AND NOT EXISTS (
    //     SELECT 1 FROM members m
    //     WHERE m.company_id = c.id
    //       AND m.last_login_at > NOW() - INTERVAL '30 days'
    //   )
    // LIMIT 10
    // 逐个 CASCADE DELETE
}
```

清理判断依据：Demo company 下所有成员的最近登录时间均超过 30 天，则认为该 Demo 已废弃。不需要在 companies 表维护额外的活跃字段。

选择在 session 逻辑中更新而非 middleware：
- 每次页面加载/刷新才触发一次，频率合理（不是每个 API 请求都写 DB）
- 不需要改全局 middleware
- UPDATE 失败不影响响应

### 3.6 Gateway（Demo 下）

Demo 租户的 Gateway 调用走 Mock LLM：
- Gateway precheck 阶段通过 platform key 查到 company，若 `type='demo'`，precheck 在响应中标记 `isDemoCompany`
- Gateway `Director` 根据该标记将请求 rewrite 到 `DEMO_MOCK_LLM_URL`（默认 `http://127.0.0.1:8765`），而非正式 NewAPI
- Mock LLM 返回模拟 response，正常走 ingest 写 ledger（看板能看到数据）
- Demo seed 灌入初始钱包余额（50000 points），确保扣费有余量；csv 模式灌入 10000 points

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
│   └── use-demo-session.ts       — Demo 状态检测
├── components/
│   ├── demo-setup-page-shell.tsx — 选择页面 UI
│   ├── demo-csv-upload.tsx       — CSV 上传 + 预览
│   ├── demo-banner.tsx           — 顶部 Banner
│   └── demo-delete-dialog.tsx    — 删除确认弹窗
└── lib/
    └── csv-parser.ts             — CSV 前端解析

api/demo.ts                       — demoApi（create/delete/template）
```

### 4.3 Demo Banner 集成

在 `AdminLayout` 中：

```tsx
{session.companyType === 'demo' && <DemoBanner />}
```

### 4.4 右上角菜单

在 Header 用户菜单中：

```tsx
{session.companyType === 'demo' && (
  <DropdownMenuItem variant="destructive" onClick={openDeleteDialog}>
    删除 Demo
  </DropdownMenuItem>
)}
```

### 4.5 登录页改造

`/login` 页面增加 "免费试用" 按钮：

```tsx
<Button variant="outline" onClick={() => navigate('/demo/setup')}>
  ✨ 免费试用
</Button>
```

正式登录表单暂时保留现有邮箱密码（placeholder，后续改手机号）。

---

## 5. 数据库变更

直接修改 `schema.sql` 中 companies 表定义（无历史数据，无需 migration）：

```sql
-- companies 表加入 type 列
CREATE TABLE IF NOT EXISTS companies (
    id                        BIGINT PRIMARY KEY,
    name                      TEXT NOT NULL,
    type                      TEXT NOT NULL DEFAULT 'selfhosted'
                              CHECK (type IN ('standard', 'trial', 'demo', 'selfhosted', 'testing')),
    status                    TEXT NOT NULL DEFAULT 'active',
    -- ... 其余现有列不变
);
```

无需额外映射表。Demo 身份完全由 JWT Cookie 中的 company_id 承载。

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
interface AppSession {
  // ... 现有字段
  companyType: 'standard' | 'trial' | 'demo' | 'selfhosted' | 'testing'
}
```

后端 `AuthzSvc.GetSessionContext` 从 `companies.type` 读取，加入 session 响应。前端通过 `session.companyType === 'demo'` 判断。

---

## 8. 安全考量

| 风险 | 缓解 |
| --- | --- |
| Demo 被批量创建 | IP 限流（5 次/小时/IP）；`DEMO_MAX_TENANTS` 上限；创建时 advisory lock 防并发超发 |
| Demo DELETE 误操作 | 二次确认弹窗 + 只允许 `type='demo'` 的 company |
| Demo 中上传恶意 CSV | 限制文件大小（1MB）；行数上限（500）；严格列解析；错误逐行报告；后端 request body limit 2MB |
| Demo 数据泄露 | Demo 数据全为模拟/用户自传；无真实企业数据 |
| Cookie 伪造访问他人 Demo | JWT 签名校验；company_id 在 JWT claims 中 |

---

## 9. 实施任务清单

### 9.1 后端

- [ ] `schema.sql` 修改：companies 表加 `type` 列
- [ ] 独立 `demo` handler 包（`internal/http/handler/demo/`）
- [ ] `POST /auth/demo/create` handler（seed + csv 两种 mode，advisory lock 防并发超发）
- [ ] `DELETE /auth/demo` handler
- [ ] `GET /auth/demo/csv-template` — 返回模板文件
- [ ] CSV 解析逻辑（部门层级自动创建 + 成员插入 + 中文角色映射）
- [ ] Demo seed 函数（复用现有 snapshot，精简）
- [ ] Gateway precheck 识别 Demo 租户 → Director rewrite 到 mock LLM
- [ ] `AuthzSvc.GetSessionContext` 增加 `companyType` 字段
- [ ] Demo 清理 River job（每日，分批删除，基于成员最近登录时间）
- [ ] `DEMO_ENABLED` 配置守卫（未启用时不注册路由）
- [ ] Demo 创建 IP 限流中间件

### 9.2 前端

- [ ] `api/demo.ts` — demoApi（create / delete / template）
- [ ] `features/demo/` 完整 feature
- [ ] `/demo/setup` 路由 + 页面（选择方式 + CSV 上传 + 返回按钮）
- [ ] CSV 前端解析 + 预览表格 + 错误提示
- [ ] CSV 模板下载
- [ ] Demo Banner 组件（纯提示，无 CTA）
- [ ] Header 菜单增加"删除 Demo"
- [ ] 删除确认 Dialog
- [ ] `/login` 页增加 "免费试用" 按钮
- [ ] `AppSession` 类型增加 `companyType`
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
