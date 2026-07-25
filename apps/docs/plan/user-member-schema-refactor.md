# User/Member 字段重构 — 实现计划

破坏性更新，直接改 schema.sql + 代码，不做 migration。

## 最终态

```
users: id, name, phone, email, password_hash, status, created_at, updated_at
members: id, company_id, user_id, alias, avatar, department_id, status, source, external_id, employee_id, job_title, override_fields, personal_budget, created_at, updated_at
```

- `users.name`：真实姓名，用户自管（注册时填写、/me 修改）
- `members.alias`：企业内别名（管理员在成员管理维护，可为空）
- `members.avatar`：头像（TEXT 列，两种格式）
  - DiceBear 生成：`dicebear:{style}:{seed}`（约 30 字节）
  - 用户上传：`data:image/webp;base64,...`（前端压缩后 ≤ 50KB）
- 前端显示名：`member.alias || session.userName`（userName 由 `/session` 响应返回，前端缓存在 session context 里）

移除：`members.name` 列、`types.Member.Phone/Email` 字段（Name → Alias）。

### 架构约束

- **禁止 COALESCE fallback**：`member.alias` 就是 alias，不做任何 SQL 层面的自动填充。为空就是空，前端自行决定 fallback 展示。
- **Phone/Email 不再 JOIN**：member 查询不再 JOIN users 取 phone/email。需要联系方式时，前端调 `/me/profile` 或管理员调独立的 user 信息接口。
- **Avatar 作用域是 member**：每个企业内头像独立。`PUT /me/profile` 改头像时只改当前 session 对应的 member 记录，不影响用户在其他企业的头像。

### name 来源对照表

| 场景 | users.name 写入 | members.alias 写入 |
|------|----------------|-------------------|
| 用户注册/接受邀请 | 输入的姓名 | 同一个姓名（初始值） |
| 管理员手动创建成员 | 输入的姓名 | 同一个姓名（初始值） |
| CSV 批量导入 | row.Name（法名） | row.Name（同值） |
| 飞书/企微远程导入 | remote.Name（法名） | remote.Name（同值） |
| 用户自改姓名（/me/profile） | 新姓名 | 不变 |
| 管理员改别名（UpdateMember） | 不变 | 新别名 |

---

## Step 1：数据库 schema.sql

```sql
-- users 加 name
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL DEFAULT '',
    phone         TEXT,
    email         TEXT,
    password_hash TEXT,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- members: name → alias + avatar
CREATE TABLE IF NOT EXISTS members (
    id              UUID NOT NULL,
    company_id      UUID NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    alias           TEXT NOT NULL DEFAULT '',
    avatar          TEXT NOT NULL DEFAULT '',
    department_id   UUID,
    status          TEXT NOT NULL,
    source          TEXT NOT NULL DEFAULT '',
    external_id     TEXT,
    employee_id     TEXT,
    job_title       TEXT,
    override_fields TEXT[] NOT NULL DEFAULT '{}',
    personal_budget BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id),
    ...
);
```

---

## Step 2：Go types

### store/user_repo.go
```go
type User struct {
    ID           uuid.UUID
    Name         string  // 新增
    Phone        string
    Email        string
    PasswordHash string
    Status       string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// UserRepository 新增:
UpdateName(ctx context.Context, id uuid.UUID, name string) error
```

### domain/types/session.go
```go
type Member struct {
    ID             uuid.UUID `json:"id"`
    CompanyID      uuid.UUID `json:"companyId"`
    UserID         uuid.UUID `json:"userId"`
    Alias          string    `json:"alias"`
    Avatar         string    `json:"avatar,omitempty"`
    Username       string    `json:"username,omitempty"`
    EmployeeID     string    `json:"employeeId,omitempty"`
    JobTitle       string    `json:"jobTitle,omitempty"`
    HireDate       string    `json:"hireDate,omitempty"`
    DepartmentID   uuid.UUID `json:"departmentId"`
    DepartmentName string    `json:"departmentName"`
    Status         string    `json:"status"`
    Roles          []string  `json:"roles"`
    Source         string    `json:"source"`
    ExternalID     *string   `json:"externalId,omitempty"`
    OverrideFields []string  `json:"overrideFields,omitempty"`  // 被用户/admin 手动修改的 user-owned 字段
    PersonalBudget int64     `json:"-"`
}

type SessionContext struct {
    CompanyID       uuid.UUID `json:"companyId"`
    CompanyType     string    `json:"companyType"`
    AuthzRevision   int64     `json:"authzRevision"`
    UserName        string    `json:"userName"`          // 新增：users.name，前端用作 alias 的 fallback
    Member          Member    `json:"member"`
    Permissions     []string  `json:"permissions"`
    ReadOnly        bool      `json:"readOnly"`
    BillingCurrency string    `json:"billingCurrency"`
    QuotaPerUnit    int64     `json:"quotaPerUnit"`
}
```

`GetSessionContext` 构建时从 users 表取 name 填入 `SessionContext.UserName`。前端 session cache 天然覆盖，无需额外请求。

---

## Step 3：SQL 查询 — store/postgres/org_repo_members.go

```diff
 const memberSelect = `
-  SELECT m.id, m.user_id, m.name, COALESCE(u.phone,''), COALESCE(u.email,''),
-    COALESCE(m.department_id, ...), COALESCE(o.name, ''), m.status, m.source, m.external_id, m.personal_budget
-  FROM members m JOIN users u ON u.id = m.user_id ...
+  SELECT m.id, m.user_id, m.alias, m.avatar,
+    COALESCE(m.department_id, ...), COALESCE(o.name, ''), m.status, m.source, m.external_id, m.personal_budget
+  FROM members m
+  LEFT JOIN org_nodes o ON o.company_id = m.company_id AND o.id = m.department_id
 `
```

- **移除 `JOIN users`**：member 列表查询不再依赖 users 表
- Scan 字段：`&item.Alias, &item.Avatar` 替代原来的 `&item.Name, &item.Phone, &item.Email`
- `SetMembers` upsert：写 alias, avatar 代替 name
- `MemberByEmail`：认证路径保留 `JOIN users`（登录需要原子校验"邮箱在该公司有活跃成员且密码匹配"），但不再填 Member.Phone/Email，只返回 passwordHash
- `GetMemberAuthz`：加一列 `u.name` 用于填 `SessionContext.UserName`（仍 JOIN users，这是单行查询，可接受）

---

## Step 4：注册流程

### 后端 handler/register/ + handler/auth/

`AcceptInvite` 和 `registerAccept` 当前传入 `Name` 给 `company.AcceptInviteRequest`：

```diff
 type AcceptInviteRequest struct {
     UserID     uuid.UUID
     InviteCode string
-    Name       string
+    Name       string  // 写入 users.name + members.alias
 }
```

`addMember` (service_create.go) 改为：
```diff
 member := types.Member{
     ...
-    Name:   name,
-    Email:  user.Email,
-    Phone:  user.Phone,
+    Alias:  name,  // 注册时用姓名作为初始 alias
 }
+// 同时更新 users.name（如果为空）
+if user.Name == "" && name != "" {
+    _ = tx.User().UpdateName(ctx, userID, name)
+}
```

### 前端 auth-popup.tsx

注册页 `memberName` 输入 → 传给 `registerAccept(inviteCode, name)` → 后端写入 `users.name` + `members.alias`。无需改前端逻辑，只是语义变了。

---

## Step 5：成员管理 — domain/org/structure + core

### resolveOrCreateUser 签名变更

```diff
-func ResolveOrCreateUser(ctx context.Context, st Store, phone, email string) (uuid.UUID, error) {
+func ResolveOrCreateUser(ctx context.Context, st Store, phone, email, name string) (uuid.UUID, error) {
     // 找到已有 user → 返回 ID（不覆盖已有 name）
     // 创建新 user → 带 name
 }
```

所有调用方统一传 name：

| 调用方 | name 来源 |
|--------|-----------|
| `member_mutate.go` CreateMember | `input.Name`（管理员输入的姓名） |
| `member_batch.go` BatchImport | `row.Name`（CSV 行的姓名） |
| `remote/import.go` 飞书导入 | `remote.Name`（飞书返回的姓名） |

### CreateMember
```diff
 member := types.Member{
-    Name: input.Name, Phone: input.Phone, Email: input.Email,
+    Alias: input.Name,  // 管理员填的姓名 → alias
 }
-userID, err := s.resolveOrCreateUser(ctx, input.Phone, input.Email)
+userID, err := s.resolveOrCreateUser(ctx, input.Phone, input.Email, input.Name)
```

### BatchImport
```diff
-userID, uerr := s.resolveOrCreateUser(ctx, row.Phone, row.Email)
+userID, uerr := s.resolveOrCreateUser(ctx, row.Phone, row.Email, row.Name)
 members = append(members, types.Member{
     ID: generateID(), UserID: userID,
-    Name: row.Name, Phone: row.Phone, Email: row.Email,
+    Alias: row.Name,
     ...
 })
```

### 飞书远程导入 (remote/import.go)
```diff
-userID, uerr := core.ResolveOrCreateUser(ctx, st, remote.Mobile, remote.Email)
+userID, uerr := core.ResolveOrCreateUser(ctx, st, remote.Mobile, remote.Email, remote.Name)
 members = append(members, types.Member{
     ...
-    Name:  remote.Name,
-    Phone: remote.Mobile,
-    Email: remote.Email,
+    Alias: remote.Name,
+    EmployeeID: remote.EmployeeNo,
     ...
 })
```

更新已有成员时使用 `syncMember()` 按字段策略逐字段覆盖（参见 `docs/plan/org-sync-override-strategy.md`）：
- immutable（employeeId）：本地为空才写入
- user-owned（alias, avatar）：字段不在 `OverrideFields` 中才覆盖
- sync-always（departmentId, departmentName）：无条件覆盖

### UpdateMember
```diff
-if input.Name != "" { existing.Name = input.Name }
-if input.Phone != "" { existing.Phone = input.Phone }
-if input.Email != "" { existing.Email = input.Email }
+if input.Name != "" && input.Name != existing.Alias {
+    existing.OverrideFields = core.TrackOverride(existing.OverrideFields, "alias")
+    existing.Alias = input.Name
+}
```

Phone/Email 更新：`UpdateMember` 接口仍接受 phone/email 参数（通过额外请求字段或 handler 层直接操作），但写入目标是 users 表。实现：handler 层拿到 member.UserID，调用 `UserRepository.UpdatePhone/UpdateEmail`。

---

## Step 6：handler/me — 合并为 `/me/profile`

复用现有 `/me/profile` 端点，扩展为可写：

```
GET  /me/profile                                          → 现有 profileResponse（加 avatar 字段）
PUT  /me/profile  { "name": "...", "avatar": "..." }      → 204
```

PUT 请求中字段为可选，只传了哪个改哪个：
- `name` → 写 `users.name`，同时 invalidate session cache（触发前端 refresh）
- `avatar` → 写当前 company 的 `members.avatar`（通过 session claims 定位 companyID + memberID）

无需新增独立的 `/me/name` 和 `/me/avatar` 端点。一个 PUT 覆盖。

### handler 依赖变更

`me.Handler` 新增 `store.OrgRepository` 依赖（用于写 member.avatar）：
```go
type Handler struct {
    shared.ProtectedHandlerBase
    memberAnalytics domainmemberanalytics.Service
    users           store.UserRepository
    org             store.OrgRepository   // 新增
    sessions        store.SessionRepository
    verifyCode      *verifycode.Service
}
```

或者更简洁：在 `store.OrgRepository` 新增 `UpdateMemberAvatar(ctx, companyID, memberID, avatar)` 方法，避免 handler 直接操作全量 member 列表。

### Avatar 校验规则（后端）

- `dicebear:{style}:{seed}` 格式：style 必须在白名单内，seed ≤ 64 字符。通过即存。
- `data:image/...;base64,...` 格式：base64 decode 后 ≤ 50KB。超出返回 400。
- 总长度硬限制：API 层拒绝 `avatar` 字段超过 70KB 的请求体（base64 膨胀 ~33%）。

### profileResponse 变更

```go
type profileResponse struct {
    Phone       string           `json:"phone"`
    Email       string           `json:"email"`
    Name        string           `json:"name"`      // users.name
    Avatar      string           `json:"avatar"`    // members.avatar（当前企业）
    HasPassword bool             `json:"hasPassword"`
    Companies   []profileCompany `json:"companies"`
}
```

### 前端缓存策略

前端已通过 `useInjectedQuery(queryKeys.session.current())` 缓存 session 数据（含 member + userName）。改 profile 后调用 `refreshSession()` 即可刷新。不需要额外缓存机制。

---

## Step 7：前端 TypeScript types

### api/types/org.ts
```diff
 export interface Member {
   id: string
   companyId: string
-  name: string
-  phone: string
-  email: string
+  alias: string
+  avatar: string
   departmentId: string
   departmentName: string
   status: MemberStatus
   roles: string[]
   source: 'imported' | 'manual' | 'invited'
   ...
 }
```

### api/types/common.ts (SessionContext)
```diff
 export interface SessionContext {
   companyId: string
   companyType: CompanyType
   authzRevision: number
+  userName: string            // users.name，显示名 fallback
   member: Member
   permissions: string[]
   readOnly: boolean
   billingCurrency: string
   quotaPerUnit: number
 }
```

### features/session/types.ts (AppSession)
```diff
 export interface AppSession {
   companyId: string
   companyType: CompanyType
   authzRevision: number
   memberId: string
   member: Member | null
+  userName: string            // 缓存自 sessionContext.userName
   permissions: string[]
   ...
 }
```

---

## Step 8：前端组件改动

显示名逻辑：`member.alias || session.userName`。

- 看自己：`useSession().member?.alias || useSession().userName`
- 看他人：`otherMember.alias`（别人的 userName 前端拿不到，也不需要）

| 文件 | 改动 |
|------|------|
| `components/layout/header.tsx` | `member?.name` → `member?.alias \|\| userName` |
| `components/layout/member-layout.tsx` | `member?.name` → `member?.alias`；头像用 `member?.avatar` |
| `features/org/components/structure/member-form-dialog.tsx` | name 字段语义改为写 alias；phone/email 保留（写入 users 表） |
| `features/org/components/structure/member-table.tsx` | `.name` → `.alias` |
| `features/org/components/role-member-table.tsx` | `.name` → `.alias` |
| `features/org/components/roles-page-shell.tsx` | `.name` → `.alias` |
| `features/org/hooks/use-structure-page.ts` | `data.name` → 传给 API 的 name 字段（后端映射到 alias） |
| `features/budget/components/budget-member-picker.tsx` | `.name` → `.alias` |
| `features/budget/components/budget-org-member-picker.tsx` | `.name` → `.alias` |
| `features/audit/hooks/use-audit-calls-page.ts` | `.name` → `.alias` |
| `features/audit/hooks/use-audit-operations-page.ts` | `.name` → `.alias` |
| `features/account/components/account-page-shell.tsx` | profile.name 来自 users.name，不变；新增 avatar 编辑 |
| 所有展示头像的地方 | 新增 avatar 渲染（renderAvatar 工具函数） |
| `e2e/org-structure.spec.ts` | `member.name` → `member.alias` |
| `e2e/feishu-import.spec.ts` | `m.name` → `m.alias` |

---

## Step 9：Seed 数据 — seed/apply/seed_org.go

```diff
-INSERT INTO users (id, phone, email, password_hash, status)
-VALUES ($1, $2, $3, $4, 'active')
+INSERT INTO users (id, name, phone, email, password_hash, status)
+VALUES ($1, $2, $3, $4, $5, 'active')

-INSERT INTO members (id, company_id, user_id, name, department_id, ...)
-VALUES ($1, $2, $3, $4, $5, ...)
+INSERT INTO members (id, company_id, user_id, alias, avatar, department_id, ...)
+VALUES ($1, $2, $3, $4, '', $5, ...)
```

seed 数据中原来的 `member.Name` → 同时写入 `users.name` 和 `members.alias`（初始保持一致）。

---

## Step 10：Avatar Picker 组件

### 依赖

```
pnpm add @dicebear/core @dicebear/adventurer @dicebear/notionists @dicebear/bottts @dicebear/shapes @dicebear/lorelei @dicebear/fun-emoji
```

注意：不用 `@dicebear/collection`（全量包 300KB+），只装用到的风格包（每个 20-40KB）。

### 存储格式

```
members.avatar TEXT:
  ""                              → 无头像，显示首字母
  "dicebear:adventurer:abc123"    → 前端用 @dicebear/core 本地渲染 SVG
  "data:image/webp;base64,..."    → 前端直接作为 img src（≤ 50KB）
```

### DiceBear 可选风格白名单

`adventurer` | `notionists` | `bottts` | `shapes` | `lorelei` | `fun-emoji`

### 前端渲染工具函数

```ts
// lib/avatar.ts
import { createAvatar } from '@dicebear/core'
import { adventurer } from '@dicebear/adventurer'
import { notionists } from '@dicebear/notionists'
import { bottts } from '@dicebear/bottts'
import { shapes } from '@dicebear/shapes'
import { lorelei } from '@dicebear/lorelei'
import { funEmoji } from '@dicebear/fun-emoji'

const styles = { adventurer, notionists, bottts, shapes, lorelei, 'fun-emoji': funEmoji } as const

export function renderAvatar(avatar: string): string {
  if (!avatar) return ''
  if (avatar.startsWith('dicebear:')) {
    const [, style, seed] = avatar.split(':')
    const styleFn = styles[style as keyof typeof styles]
    if (!styleFn) return ''
    return createAvatar(styleFn, { seed }).toDataUri()
  }
  return avatar  // data:image/... 直接用
}
```

### Picker 交互

```
弹出 Dialog：
├── Tab 1：随机头像
│   ├── 风格选择器（6 个 Tab：adventurer / notionists / bottts / shapes / lorelei / fun-emoji）
│   ├── 网格展示 9 个随机 seed 生成的头像
│   ├── 🎲 按钮：重新随机 9 个
│   └── 点击任一头像 → 选中
├── Tab 2：上传图片
│   ├── 拖拽/点击上传
│   ├── 前端 canvas 裁剪 → encode webp → base64
│   └── 校验 base64 decode ≤ 50KB，超出提示
└── 底部：[取消] [确认]
```

确认后调用 `PUT /me/profile { avatar: "dicebear:..." }` 或 `PUT /me/profile { avatar: "data:image/webp;base64,..." }`，成功后 `refreshSession()` 刷新缓存。

---

## 执行顺序

1. schema.sql（users 加 name、members 去 name 加 alias+avatar）+ seed（紧邻，避免列数不匹配）
2. Go store 层（User struct 加 Name + UserRepository.UpdateName + OrgRepository.UpdateMemberAvatar）
3. Go types（Member struct 改字段、SessionContext 加 UserName）
4. Go SQL 查询（org_repo_members.go scan 适配、GetMemberAuthz 取 userName）
5. Go domain（ResolveOrCreateUser 加 name 参数、member_mutate/batch/remote 全部适配）
6. Go handler（PUT /me/profile 扩展，含 avatar 校验 + name 写入）
7. TS types（Member interface、SessionContext、AppSession）
8. TS session provider（传递 userName）
9. TS 组件（全局 member.name → member.alias）
10. TS avatar 渲染工具 + Avatar Picker 组件
11. E2E 测试适配
12. 全量 build + lint + test
