# 身份与租户数据模型

> 本文定义 TokenJoy 的三层数据模型：User（全局身份）、Member（企业角色）、Company（租户）。  
> 所有 schema 以本文为 SSOT。

---

## 1. User — 全局身份

| 表 | 职责 | 唯一性 |
|---|------|--------|
| `users` | 认证凭证、联系方式 | phone 全局唯一，email 全局唯一 |

### Schema

```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone         TEXT,
    email         TEXT,
    password_hash TEXT,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL AND phone != '';
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL AND email != '';
```

---

## 2. Member — 企业角色

| 表 | 职责 | 唯一性 |
|---|------|--------|
| `members` | 企业内部门、权限、状态 | (user_id, company_id) 唯一 |

### Schema

```sql
CREATE TABLE members (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id),
    company_id    UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    department_id UUID,
    roles         TEXT[] NOT NULL DEFAULT '{}',
    status        TEXT NOT NULL DEFAULT 'active',
    source        TEXT NOT NULL DEFAULT 'invite',
    employee_id   TEXT,
    job_title     TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, company_id)
);

CREATE INDEX idx_members_company ON members(company_id);
CREATE INDEX idx_members_user ON members(user_id);
```

### 字段归属

| 字段 | 属于 | 原因 |
|------|------|------|
| phone | user | 全局唯一身份标识 |
| email | user | 全局唯一身份标识 |
| password_hash | user | 改密码全局生效 |
| name | member | 同一人在不同公司可能用不同名字 |
| department_id | member | 企业内组织结构 |
| roles | member | 企业内权限 |
| status | 两者都有 | user.status=disabled 全局禁用；member.status=disabled 仅退出某企业 |
| employee_id | member | 企业内工号 |
| job_title | member | 企业内职位 |

---

## 3. 关系

```
user (1) ──→ (N) member ──→ (1) company
```

一个人 = 一个 user。一个人在一家公司 = 一条 member。

**规则**：

1. 一个 phone/email 只能对应一个 user
2. 一个 user 在同一家 company 只能有一条 member
3. 一个 user 可以属于多家 company（通过多条 member）
4. 登录认证 user，业务授权 member
5. 所有业务逻辑（权限/预算/Key/审计/通知）绑定 member_id

---

## 4. Company — 租户

### 4.1 两个维度

| 维度 | 职责 | 特点 |
|------|------|------|
| `type` | 描述租户的**本质**（它是什么） | 创建后不变（唯一例外：trial → standard） |
| `status` | 描述租户的**当前生命周期状态** | 会流转 |

两者正交：任何 type 的 company 都可以处于任何 status。

### 4.2 type 枚举

| 值 | 含义 | 部署形态 | 典型生命周期 |
|------|------|---------|-------------|
| `standard` | SaaS 正式付费客户 | SaaS | 永久 |
| `trial` | SaaS 免费试用（有账号，模拟资金） | SaaS | 永久（升级后变 standard） |
| `demo` | 匿名体验沙箱（保留，暂不实现） | SaaS | 30 天无活动 → 自动清理 |
| `selfhosted` | 私有化部署企业 | 非 SaaS | 永久，单实例唯一 |
| `testing` | 开发 / CI 自动化测试 | 开发环境 | 不清理 |

**确定时机**：

- `selfhosted`：非 SaaS 部署时 bootstrap 创建，或 `/setup` wizard 初始化
- `testing`：开发环境 seed 脚本创建
- `trial`：SaaS 注册流程创建（`POST /auth/register`）
- `standard`：平台开户创建，或 trial 付费升级
- `demo`：保留，暂不实现

### 4.3 status 枚举

| 值 | 含义 | 影响 |
|------|------|------|
| `active` | 正常运行 | 全功能可用 |
| `suspended` | 冻结 | 只读（GET/HEAD/OPTIONS 放行，写入返回 403）；Gateway API 调用被拒绝 |

**流转**：`active` ↔ `suspended`（平台管理员操作 / 欠费 / 恢复）

**检查点**：

| 位置 | 逻辑 |
|------|------|
| `CompanyReadOnlyMiddleware` | `status == suspended` → 阻止写入请求，返回 403 |
| `IsGatewayBlocked()` | `status != active` → Gateway precheck 拒绝 API 调用 |
| `ForEachActiveCompany()` | 只遍历 `status == active` 的 company（定时任务） |

### 4.4 Schema

```sql
CREATE TABLE IF NOT EXISTS companies (
    id                        UUID PRIMARY KEY,
    name                      TEXT NOT NULL,
    type                      TEXT NOT NULL DEFAULT 'selfhosted'
                              CHECK (type IN ('standard', 'trial', 'demo', 'selfhosted', 'testing')),
    status                    TEXT NOT NULL DEFAULT 'active',
    root_dept_id              UUID,
    newapi_wallet_company_id  BIGINT,
    authz_revision            BIGINT NOT NULL DEFAULT 0,
    billing_currency          CHAR(3) NOT NULL DEFAULT 'CNY',
    fifo_head_lot_id          UUID,
    wallet_remain_quota             BIGINT NOT NULL DEFAULT 0,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Trial 无到期机制，不需要 `trial_expires_at` 列。

---

## 5. 部署模式

### 5.1 非 SaaS 模式（私有化）

配置：`SUPPORT_SAAS=false`

- `LocalCompanyID`（默认 `00000000-0000-7000-8000-000000000002`）作为唯一租户
- 用户登录无需指定 company，middleware 自动注入
- 无 platform operator 管理后台
- 无 Trial/注册功能（`REGISTRATION_ENABLED` 默认 false）
- companies 表只有 1 条记录（type=selfhosted）+ 可选 1 条 testing

### 5.2 SaaS 模式

配置：`SUPPORT_SAAS=true`

- JWT 必须携带 `company_id`，无隐式租户
- 登录时前端传 `companyId` 参数
- Platform operator 可管理所有租户（创建、冻结、充值）
- Trial 注册可选开启（`REGISTRATION_ENABLED`）
- companies 表有多条记录

### 5.3 模式对比

| 能力 | 非 SaaS（`selfhosted`） | SaaS |
|------|-------------------|------|
| 租户数量 | 1 | 多个 |
| Company 解析 | `LocalCompanyID` 兜底 | JWT 必须携带 |
| Platform 管理后台 | 无 | 有 |
| Trial 注册 | 无 | 可选 |
| 多 company 切换 | 不支持 | 支持 |
| 注册/邀请 | 仅邀请（`/setup` wizard） | 注册 + 邀请 |
| 供应商 Key | 企业自管（完全 CRUD） | 平台统一管理（企业只读） |

---

## 6. 各 type 行为差异

| 行为 | `selfhosted` | `testing` | `trial` | `standard` |
|------|---------|-----------|---------|-----------|
| Gateway 代理目标 | 正式 NewAPI | dev-mock-llm | Mock LLM | 正式 NewAPI |
| 钱包 / 充值 | 全功能 | 全功能 | 模拟资金，充值禁用 | 全功能 |
| 通知投递 | 正常 | 不投递 | 正常 | 正常 |
| 生命周期管理 | 不清理 | 不清理 | 永久（升级后变 standard） | 不清理 |
| 第三方数据源 | 全功能 | 全功能 | 全功能 | 全功能 |
| 成员上限 | 无限制 | 无限制 | 50 人 | 无限制 |
| 可被删除 | 不可 | 可（手动） | 可（自动清理） | 不可 |
| 升级路径 | — | — | → `standard` | — |

---

## 7. JWT

```json
{
  "sub": "<memberID>",
  "company_id": "<companyID>",
  "user_id": "<userID>",
  "sid": "<sessionID>",
  "exp": 1710086400
}
```

- `sub` — member UUID（权限检查、业务操作用）
- `company_id` — 当前企业 UUID
- `user_id` — 全局用户 UUID
- 认证身份（改密码、改手机号）通过 member → user 关联

---

## 8. Session 响应扩展

```typescript
interface AppSession {
  companyType: 'standard' | 'trial' | 'demo' | 'selfhosted' | 'testing'
}
```

---

## 9. 代码影响（type 引入后）

| 位置 | 变更 |
|------|------|
| `schema.sql` companies 表 | 加 `type` 列定义 |
| `store.Company` struct | 加 `Type string` 字段 |
| `ctxcompany.Info` | 加 `Type string`（供 middleware / service 判断） |
| `CreateCompany` / `CreateCompanyRequest` | 加 `Type` 参数（默认 `standard`，Trial 传 `trial`） |
| `seed/bootstrap/companies.go` | INSERT 加上 type 列（selfhosted / testing） |
| `AuthzSvc.GetSessionContext` | session 响应增加 `companyType` 字段 |
| Gateway precheck | 检查 type，Trial 租户路由到 mock LLM |
| 前端 `AppSession` | 加 `companyType` 字段 |

---

## 10. 设计决策

| 决策 | 理由 |
|------|------|
| type 和 status 分开 | 职责不同。type 不变（本质），status 流转（状态）。混合导致组合爆炸 |
| 不拆表（trial_companies / prod_companies） | 所有 type 共享相同业务数据模型，FK 和 CASCADE DELETE 统一。行业标准同表 + type 列 |
| type 仅允许 trial → standard 流转 | 付费升级原地 UPDATE，数据全部保留 |
| User/Member 分表 | User 跨企业唯一（凭证），Member 企业内唯一（角色）。分离关注点 |
| JWT sub 为 memberID | 业务逻辑绑定 member_id，减少关联查询 |
