# User-Member 模型

---

## 职责

| 表 | 职责 | 唯一性 |
|---|------|--------|
| `users` | 全局身份：凭证、联系方式 | phone 全局唯一，email 全局唯一 |
| `members` | 企业内角色：部门、权限、状态 | (user_id, company_id) 唯一 |

---

## Schema

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

---

## 关系

```
user (1) ──→ (N) member ──→ (1) company
```

一个人 = 一个 user。一个人在一家公司 = 一条 member。

---

## 字段归属

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

## JWT

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "company_id": "00000000-0000-7000-8000-000000000002",
  "iat": 1710000000,
  "exp": 1710086400
}
```

- `sub` — member UUID（权限检查、业务操作用）
- `company_id` — 当前企业 UUID
- 认证身份（改密码、改手机号）通过 member → user 关联

---

## 规则

1. 一个 phone/email 只能对应一个 user
2. 一个 user 在同一家 company 只能有一条 member
3. 一个 user 可以属于多家 company（通过多条 member）
4. 登录认证 user，业务授权 member
5. 所有业务逻辑（权限/预算/Key/审计/通知）绑定 member_id
