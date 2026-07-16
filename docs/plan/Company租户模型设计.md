# Company 租户模型设计

> 本文定义 `companies` 表的 `type` 和 `status` 两个维度，以及 SaaS / 非 SaaS 两种部署模式下的运作方式。

---

## 1. 两个维度

| 维度 | 职责 | 特点 |
|------|------|------|
| `type` | 描述租户的**本质**（它是什么） | 创建后不变 |
| `status` | 描述租户的**当前生命周期状态** | 会流转 |

两者正交：任何 type 的 company 都可以处于任何 status。

---

## 2. type 枚举

| 值 | 含义 | 部署形态 | 典型生命周期 |
|------|------|---------|-------------|
| `standard` | SaaS 正式付费客户 | SaaS | 永久 |
| `trial` | SaaS 注册试用（有账号，限时） | SaaS | 到期 → 降级或冻结 |
| `demo` | 匿名体验（无账号，Cookie 绑定） | SaaS | 30 天无活动 → 自动清理 |
| `selfhosted` | 私有化部署企业 | 非 SaaS | 永久，单实例唯一 |
| `testing` | 开发 / CI 自动化测试 | 开发环境 | 不清理 |

### 当前实现状态

| type | 状态 |
|------|------|
| `selfhosted` | 已实现（现有私有化模式） |
| `testing` | 已实现（开发 seed company） |
| `demo` | 本次实现 |
| `standard` | Placeholder，SaaS 注册流程时实现 |
| `trial` | Placeholder，远期 |

### type 的确定时机

- `selfhosted`：非 SaaS 部署时 bootstrap 创建，或 `/setup` wizard 初始化
- `testing`：开发环境 seed 脚本创建
- `demo`：`POST /auth/demo/create` 创建
- `standard`：SaaS 注册流程创建（远期）
- `trial`：SaaS 试用注册创建（远期）

type 一旦确定不允许修改（`demo` → `standard` 的"升级"通过创建新 company + 数据迁移实现，而非直接改 type）。

---

## 3. status 枚举

| 值 | 含义 | 影响 |
|------|------|------|
| `active` | 正常运行 | 全功能可用 |
| `suspended` | 冻结 | 只读（GET/HEAD/OPTIONS 放行，写入请求返回 403）；Gateway API 调用被拒绝 |

### status 流转

```
创建 → active
         │
    平台管理员操作 / 欠费 / 到期
         │
         ▼
     suspended
         │
    恢复 / 续费 / 管理员操作
         │
         ▼
       active
```

### status 的影响面（现有代码）

| 检查点 | 逻辑 |
|--------|------|
| `CompanyReadOnlyMiddleware` | `status == suspended` → 阻止所有写入请求（POST/PUT/PATCH/DELETE），返回 403 |
| `IsGatewayBlocked()` | `status != active` → Gateway precheck 拒绝 API 调用 |
| `ForEachActiveCompany()` | 只遍历 `status == active` 的 company（定时任务） |

---

## 4. 部署模式

### 4.1 非 SaaS 模式（私有化）

配置：`SUPPORT_SAAS=false`

```
┌─────────────────────────────────────────────────┐
│  单实例部署                                      │
│                                                  │
│  companies 表只有 1 条记录（type=selfhosted）     │
│  + 可选 1 条 testing（开发环境）                  │
│                                                  │
│  Company 解析方式：                              │
│  CompanyResolve middleware 使用 LocalCompanyID    │
│  作为隐式租户（无需 JWT 携带 company_id）         │
│                                                  │
│  无 /platform 路由                               │
│  无多租户管理                                    │
└─────────────────────────────────────────────────┘
```

**关键行为**：
- `LocalCompanyID`（默认 2）作为唯一租户
- 用户登录无需指定 company，middleware 自动注入
- 无 platform operator 管理后台
- 无 Demo 功能（`DEMO_ENABLED` 默认 false）

### 4.2 SaaS 模式

配置：`SUPPORT_SAAS=true`

```
┌─────────────────────────────────────────────────┐
│  多租户 SaaS 部署                                │
│                                                  │
│  companies 表有多条记录                          │
│  type 可为 standard / trial / demo               │
│                                                  │
│  Company 解析方式：                              │
│  CompanyResolve middleware 从 JWT claims 中      │
│  提取 company_id（无 LocalCompanyID 兜底）       │
│                                                  │
│  有 /platform 路由（平台运营管理后台）            │
│  有 Demo 入口（DEMO_ENABLED 控制）               │
└─────────────────────────────────────────────────┘
```

**关键行为**：
- JWT 必须携带 `company_id`，无隐式租户
- 登录时前端传 `companyId` 参数
- Platform operator 可管理所有租户（创建、冻结、充值）
- Demo 可选开启

### 4.3 模式对比

| 能力 | 非 SaaS（`selfhosted`） | SaaS |
|------|-------------------|------|
| 租户数量 | 1 | 多个 |
| Company 解析 | `LocalCompanyID` 兜底 | JWT 必须携带 |
| Platform 管理后台 | 无 | 有 |
| Demo 入口 | 无 | 可选 |
| 多 company 切换 | 不支持 | 支持 |
| 注册/邀请 | 仅邀请（`/setup` wizard） | 注册 + 邀请 |

---

## 5. 各 type 的行为差异

| 行为 | `selfhosted` | `testing` | `demo` | `standard` | `trial` |
|------|---------|-----------|--------|-----------|---------|
| Gateway 代理目标 | 正式 NewAPI | dev-mock-llm | mock LLM | 正式 NewAPI | 正式 NewAPI |
| 钱包 / 充值 | 全功能 | 全功能 | 固定余额，充值禁用 | 全功能 | 有限额度 |
| 通知投递 | 正常 | 不投递 | 不投递 | 正常 | 正常 |
| 自动清理 | 不清理 | 不清理 | 30 天无活动清理 | 不清理 | 到期冻结 |
| 第三方数据源 | 全功能 | 全功能 | 禁用 | 全功能 | 全功能 |
| 可被删除 | 不可 | 可（手动） | 可（用户自助 + 自动） | 不可 | 不可 |

---

## 6. 数据库 DDL

直接在 `schema.sql` 中修改 companies 表定义（无历史数据，无需 migration）：

```sql
CREATE TABLE IF NOT EXISTS companies (
    id                        BIGINT PRIMARY KEY,
    name                      TEXT NOT NULL,
    type                      TEXT NOT NULL DEFAULT 'selfhosted'
                              CHECK (type IN ('standard', 'trial', 'demo', 'selfhosted', 'testing')),
    status                    TEXT NOT NULL DEFAULT 'active',
    root_dept_id              TEXT,
    -- ... 其余现有列不变
);
```

Demo 清理不依赖 companies 表额外字段，而是通过成员表的最近登录时间判断活跃度。

bootstrap 时：
- 非 SaaS 模式创建的 company 写入 `type='selfhosted'`
- 开发 seed company 写入 `type='testing'`

---

## 7. 代码影响（type 引入后）

| 位置 | 变更 |
|------|------|
| `schema.sql` companies 表 | 加 `type` 列定义 |
| `store.Company` struct | 加 `Type string` 字段 |
| `store.CompanyRepository` | 无需新方法，`Create` 时写入 type 即可 |
| `ctxcompany.Info` | 加 `Type string`（供 middleware / service 判断） |
| `CompanyResolve` middleware | `ResolveCompanyContext` 已返回完整 context，type 自然带入 |
| `CreateCompany` / `CreateCompanyRequest` | 加 `Type` 参数（默认 `standard`，Demo 传 `demo`） |
| `ensureBootstrapCompany` | INSERT 加上 type 列（selfhosted / testing） |
| `AuthzSvc.GetSessionContext` | session 响应增加 `companyType` 字段 |
| Gateway precheck | 检查 type，Demo 租户路由到 mock LLM |
| 清理 job | 按 `type='demo'` + 成员最近登录时间清理 |
| 前端 `AppSession` | 加 `companyType` 字段 |

---

## 8. 设计决策记录

**Q：为什么 type 和 status 分开，不用一个字段？**

A：职责不同。type 是"它是什么"（本质，不变），status 是"它现在怎么样"（状态，会流转）。一个 demo company 也可能被 suspended（比如检测到滥用），一个 standard company 也可能从 active 变为 suspended 再恢复。混在一起会导致状态组合爆炸。

**Q：为什么不拆表（demo_companies / prod_companies）？**

A：所有 type 的 company 共享相同的业务数据模型（部门、成员、预算、模型、key、调用记录），子表全部 FK 到 `companies.id`。拆表会导致 FK 和 CASCADE DELETE 变复杂，收益为零。行业标准做法是同表 + type 列。

**Q：type 能不能改？**

A：设计上不允许直接修改。`demo` → `standard` 的"升级"是创建新 company + 迁移数据的流程，不是简单 UPDATE type。这样确保各 type 的生命周期策略（清理、计费）不会出现不一致的中间态。
