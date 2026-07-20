# 角色系统重构：全局预设 + 公司自定义

> **状态**：实现文档  
> **前提**：项目未上线，无迁移/兼容负担，直接改 schema。

---

## 1. 目标

将预设角色从"每公司一份"改为"全局唯一一份"；保留公司自定义角色能力；删除 `role_permission_grants` 表。

**当前**：
- 预设角色每公司复制一份，ID 由 `uuid.NewSHA1(companyID, "preset-role:"+name)` 哈希生成
- `role_permission_grants` 表存储角色↔权限映射（包括预设角色的展开后 grant）
- 每次启动 `ReconcilePresetRoles` 遍历所有公司补齐 grants

**目标**：
- 预设角色全局一份，固定 UUID 常量，`roles.company_id = NULL`
- 预设角色权限不落 DB，运行时从 manifest 展开（已有此逻辑）
- 自定义角色权限存 `roles.permissions TEXT[]`，删除 `role_permission_grants` 表
- 删除 `ReconcilePresetRoles`、`defaultCompanyRoles`；将 `PresetRoleID` 签名从 `(companyID, name)` 改为 `(name)`（仅做 name→ID 查找）

---

## 2. Schema 变更

### 2.1 `roles` 表（改）

```sql
CREATE TABLE IF NOT EXISTS roles (
    id          UUID PRIMARY KEY,                                          -- 单列 PK
    company_id  UUID REFERENCES companies (id) ON DELETE CASCADE,          -- NULL = 全局预设
    name        TEXT NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('preset', 'custom')),
    permissions TEXT[] NOT NULL DEFAULT '{}',                              -- 自定义角色存 p-* ID
    UNIQUE (company_id, name)
);

-- PostgreSQL 对 UNIQUE 约束中 NULL 视为不相等，(NULL, name) 不受 UNIQUE (company_id, name) 保护。
-- 需补充 partial unique index 保证全局预设角色名唯一：
CREATE UNIQUE INDEX idx_roles_preset_name ON roles (name) WHERE company_id IS NULL;
```

与现有的差异：
- PK 从 `(company_id, id)` → `(id)`
- `company_id` 允许 NULL
- 新增 `permissions TEXT[]`（替代 `role_permission_grants`）
- 新增 partial unique index `idx_roles_preset_name`（NULL 安全）

### 2.2 `role_permission_grants` 表（删）

整表删除。

### 2.3 `member_roles` 表（改 FK）

```sql
CREATE TABLE IF NOT EXISTS member_roles (
    company_id UUID NOT NULL,
    member_id  UUID NOT NULL,
    role_id    UUID NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
    PRIMARY KEY (company_id, member_id, role_id)
);
```

差异：
- `role_id` FK 从 `REFERENCES roles (company_id, id)` 改为 `REFERENCES roles (id)`
- **ON DELETE 改为 RESTRICT**（原为 CASCADE）：全局预设角色被所有公司共享，如果误删会级联清除所有公司的 member-role 关联。RESTRICT 强制先手动解绑再删角色，防止灾难性误操作。

### 2.4 `alert_rule_notify_roles` 表（改 FK）

```sql
CREATE TABLE IF NOT EXISTS alert_rule_notify_roles (
    company_id UUID NOT NULL,
    rule_id    UUID NOT NULL,
    role_id    UUID NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
    PRIMARY KEY (company_id, rule_id, role_id),
    FOREIGN KEY (company_id, rule_id) REFERENCES alert_rules (company_id, id) ON DELETE CASCADE
);
```

差异：
- `role_id` FK 从 `(company_id, role_id) REFERENCES roles` 改为直接 `REFERENCES roles (id)`
- ON DELETE 同样改为 RESTRICT（同 §2.3 原因）

> **应用层约束**：`role_id` FK 不再包含 `company_id`，理论上可以引用其他公司的自定义角色。应用层写入时须校验 `role_id` 对应的角色满足 `company_id IS NULL OR company_id = 当前公司 ID`。此校验在 `store/postgres/budget_repo_alerts.go` 的 upsert 逻辑中实现。

---

## 3. `internal/domain/grants/roles.go`（改）

```go
package grants

import "github.com/google/uuid"

const (
    RoleSuperAdmin     = "超级管理员"
    RoleOrgAdmin       = "组织管理员"
    RoleMember         = "普通成员"
    RoleAuditor        = "只读审计员"
    RoleAPICaller      = "API 调用者"
    RoleBudgetApprover = "预算审批员"
    RolePlatformAdmin  = "平台管理员"
)

// 全局预设角色固定 UUID。
// 注意：RoleBudgetApprover 不在此列表中——它在 demo seed 中作为公司自定义角色创建，不是全局预设。
var (
    IDSuperAdmin   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
    IDOrgAdmin     = uuid.MustParse("00000000-0000-0000-0000-000000000002")
    IDMember       = uuid.MustParse("00000000-0000-0000-0000-000000000003")
    IDAuditor      = uuid.MustParse("00000000-0000-0000-0000-000000000004")
    IDAPICaller    = uuid.MustParse("00000000-0000-0000-0000-000000000005")
    IDPlatformAdmin = uuid.MustParse("00000000-0000-0000-0000-000000000006")
)

// PresetRoles 名称→ID（仅全局预设角色）。
var PresetRoles = map[string]uuid.UUID{
    RoleSuperAdmin:  IDSuperAdmin,
    RoleOrgAdmin:    IDOrgAdmin,
    RoleMember:      IDMember,
    RoleAuditor:     IDAuditor,
    RoleAPICaller:   IDAPICaller,
    RolePlatformAdmin: IDPlatformAdmin,
}

// PresetRoleID 按名称查全局预设角色 ID。
// 签名从 (companyID, name) 简化为 (name)，因为预设角色不再按公司分配。
func PresetRoleID(name string) uuid.UUID {
    return PresetRoles[name]
}

func IsPresetRole(name string) bool {
    _, ok := PresetRoles[name]
    return ok
}
```

**关键变更**：
- 删除旧签名 `PresetRoleID(companyID uuid.UUID, roleName string) uuid.UUID`
- `RoleBudgetApprover` 保留为 const（角色名常量仍需引用），但 **不** 放入 `PresetRoles` map 和全局 ID 常量中——它是各公司按需创建的自定义角色，demo seed 中由 `snapshot/org.go` 用 `uuid.NewV7()` 生成 ID

---

## 4. Seed 变更

### 4.1 `seed/bootstrap/roles.go`（重写）

```go
package bootstrap

import (
    "context"
    "fmt"

    "github.com/tokenjoy/backend/internal/domain/grants"
)

// seedGlobalPresetRoles 写入全局预设角色行。幂等。
func seedGlobalPresetRoles(ctx context.Context, exec TableWriter) error {
    for name, id := range grants.PresetRoles {
        if _, err := exec.Exec(ctx, `
            INSERT INTO roles (id, company_id, name, type, permissions)
            VALUES ($1, NULL, $2, 'preset', '{}')
            ON CONFLICT (id) DO NOTHING
        `, id, name); err != nil {
            return fmt.Errorf("seed preset role %s: %w", name, err)
        }
    }
    return nil
}
```

**删除**：`insertPresetRoles(ctx, exec, companyID)`、`reconcileCompanyPresetRoles`、`resolveGrantIDs`、`allPermissionIDs`。

### 4.2 `seed/bootstrap/bootstrap.go`（改）

```go
// 原：insertPresetRoles(ctx, exec, companyID)
// 改：
if err := seedGlobalPresetRoles(ctx, exec); err != nil {
    return fmt.Errorf("bootstrap roles: %w", err)
}
```

**删除**：`ReconcilePresetRoles` 函数整体。

### 4.3 `seed/bootstrap/org.go`（改）

```go
// 原：superAdminRoleID := grants.PresetRoleID(companyID, grants.RoleSuperAdmin)
// 改：
superAdminRoleID := grants.IDSuperAdmin
```

### 4.4 `seed/init.go`（改）

删除整个 reconcile 步骤：

```go
// 删除：
// companyIDs, err := listCompanyIDs(ctx, pool)
// bootstrap.ReconcilePresetRoles(ctx, pool, companyIDs)
```

### 4.5 `seed/apply/seed_core.go`（改）

`insertSeedRoles`：不再写 `role_permission_grants`，改为写 `permissions` 列：

```go
func insertSeedRoles(ctx context.Context, exec TableWriter, roles []types.Role) error {
    for _, role := range roles {
        if _, err := exec.Exec(ctx, `
            INSERT INTO roles (id, company_id, name, type, permissions)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, type = EXCLUDED.type, permissions = EXCLUDED.permissions
        `, role.ID, nilUUID(role.CompanyID), role.Name, role.Type, role.Permissions); err != nil {
            return fmt.Errorf("seed role %s: %w", role.ID, err)
        }
    }
    return nil
}
```

注意：预设角色 `company_id = NULL`，自定义角色传实际公司 ID。

### 4.6 `seed/snapshot/org.go`（改）

`buildRoles` 改为使用 `grants.ID*` 常量（预设角色）+ `uuid.NewV7()` 生成 ID（自定义角色）：

```go
func buildRoles(members []types.Member) []types.Role {
    return []types.Role{
        {ID: grants.IDSuperAdmin, Name: grants.RoleSuperAdmin, Type: "preset", MemberCount: org.CountMembersByRole(members, grants.RoleSuperAdmin)},
        {ID: grants.IDOrgAdmin, Name: grants.RoleOrgAdmin, Type: "preset", MemberCount: org.CountMembersByRole(members, grants.RoleOrgAdmin)},
        {ID: grants.IDMember, Name: grants.RoleMember, Type: "preset", MemberCount: org.CountMembersByRole(members, grants.RoleMember)},
        {ID: grants.IDAuditor, Name: grants.RoleAuditor, Type: "preset", MemberCount: org.CountMembersByRole(members, grants.RoleAuditor)},
        {ID: grants.IDAPICaller, Name: grants.RoleAPICaller, Type: "preset", MemberCount: org.CountMembersByRole(members, grants.RoleAPICaller)},
        // BudgetApprover 是公司自定义角色，用固定 seed ID（非全局预设）
        {ID: contract.IDRoleBudgetApprover, CompanyID: contract.DefaultCompanyID, Name: grants.RoleBudgetApprover, Type: "custom", Permissions: mustNormalize([]string{"p-6"}), MemberCount: org.CountMembersByRole(members, grants.RoleBudgetApprover)},
    }
}
```

### 4.7 `seed/contract/ids.go`（改）

```go
// 原：
// IDRole1 = grants.PresetRoleID(DefaultCompanyID, grants.RoleSuperAdmin)
// ...
// IDRole6 = grants.PresetRoleID(DefaultCompanyID, grants.RoleBudgetApprover)

// 改：预设角色直接引用全局常量，自定义角色用固定 seed UUID
var (
    IDRole1 = grants.IDSuperAdmin
    IDRole2 = grants.IDOrgAdmin
    IDRole3 = grants.IDMember
    IDRole4 = grants.IDAuditor
    IDRole5 = grants.IDAPICaller
    // BudgetApprover 是 demo seed 的自定义角色，用固定 UUID 保证 seed 幂等
    IDRoleBudgetApprover = uuid.MustParse("00000000-0000-7000-8000-000000000060")
)
```

---

## 5. Store 层变更

### 5.1 `store/postgres/org_repo_roles.go`（重写）

```go
func (r *pgOrgRepo) Roles(ctx context.Context) ([]types.Role, error) {
    companyID := store.CompanyID(ctx)
    // ponytail: OR 条件可能不走索引，当前角色数极少无影响；
    // 若将来角色数 >1000，改为 UNION ALL（全局预设 + 本公司自定义分开查）。
    rows, err := r.db.Query(ctx, `
        SELECT r.id, r.name, r.type, r.permissions, COUNT(mr.member_id)::int AS member_count
        FROM roles r
        LEFT JOIN member_roles mr ON mr.role_id = r.id AND mr.company_id = $1
        WHERE r.company_id IS NULL OR r.company_id = $1
        GROUP BY r.id, r.name, r.type, r.permissions
        ORDER BY r.type, r.name
    `, companyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var items []types.Role
    for rows.Next() {
        var role types.Role
        if err := rows.Scan(&role.ID, &role.Name, &role.Type, &role.Permissions, &role.MemberCount); err != nil {
            return nil, err
        }
        items = append(items, role)
    }
    return items, rows.Err()
}

func (r *pgOrgRepo) SetRoles(ctx context.Context, roles []types.Role) error {
    companyID := store.CompanyID(ctx)
    ids := make([]uuid.UUID, 0, len(roles))
    for _, role := range roles {
        if role.Type == "preset" {
            continue // 不操作全局预设行
        }
        ids = append(ids, role.ID)
        if _, err := r.db.Exec(ctx, `
            INSERT INTO roles (id, company_id, name, type, permissions)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, permissions = EXCLUDED.permissions
        `, role.ID, companyID, role.Name, role.Type, role.Permissions); err != nil {
            return fmt.Errorf("upsert role %s: %w", role.ID, err)
        }
    }
    // Prune deleted custom roles.
    // 安全说明：pruneByIDForCompany 使用 `DELETE FROM roles WHERE company_id = $1 AND id NOT IN (...)`，
    // 全局预设角色 company_id IS NULL 不会被 company_id = $1 匹配到，不会被误删。
    if len(ids) == 0 {
        _, err := r.db.Exec(ctx, `DELETE FROM roles WHERE company_id = $1`, companyID)
        return err
    }
    return pruneByIDForCompany(ctx, r.db, "roles", companyID, ids)
}
```

**删除**：所有 `role_permission_grants` 相关查询。

`loadRoleNameIndex` 改为查全局预设 + 公司自定义：

```go
func loadRoleNameIndex(ctx context.Context, db dbQuerier, companyID uuid.UUID) (map[string]string, error) {
    rows, err := db.Query(ctx, `
        SELECT id::text, name FROM roles WHERE company_id IS NULL OR company_id = $1
    `, companyID)
    if err != nil {
        return nil, fmt.Errorf("load roles index: %w", err)
    }
    defer rows.Close()
    index := make(map[string]string)
    for rows.Next() {
        var id, name string
        if err := rows.Scan(&id, &name); err != nil {
            return nil, err
        }
        index[name] = id
    }
    return index, rows.Err()
}
```

### 5.2 `store/postgres/org_repo_members.go`（改）

`memberListSelect` 中 JOIN 条件改为：

```sql
LEFT JOIN member_roles mr ON mr.company_id = m.company_id AND mr.member_id = m.id
LEFT JOIN roles r ON r.id = mr.role_id
```

（去掉 `r.company_id = mr.company_id`，因为全局预设角色 `company_id` 为 NULL。）

`MemberByID`、`MemberByEmail`、`GetMemberAuthz` 中角色子查询统一改为：

```sql
SELECT r.name FROM member_roles mr
JOIN roles r ON r.id = mr.role_id
WHERE mr.company_id = $1 AND mr.member_id = $2
ORDER BY r.name
```

（去掉 `ro.company_id = mr.company_id`。）

### 5.3 `store/postgres/budget_repo_alerts.go`（改）

告警规则关联角色的写入处须增加校验：

```go
// 写入 alert_rule_notify_roles 前校验 role_id 归属
// 确保 role.company_id IS NULL（全局预设）或 role.company_id = 当前公司
```

### 5.4 `GetMemberAuthz` 简化

`rolesForCompany` 已改为返回全局预设 + 公司自定义（见 5.1 `Roles` 方法），无需变更 authz 调用方。

---

## 6. Domain 层变更

### 6.1 `domain/company/service_create.go`（改）

**删除** `defaultCompanyRoles` 函数。

`provisionCompany` 中删除：

```go
// 删除：
if err := tx.Org().SetRoles(companyCtx, defaultCompanyRoles(companyID, s.grants)); err != nil {
    return store.Company{}, err
}
```

创建公司时不再插入角色。全局预设角色已在 `roles` 表中（bootstrap seed）。

### 6.2 `domain/org/structure/role_crud.go`（改）

`CreateRole`：名称冲突校验须包含全局预设角色名：

```go
func (s *LocalService) CreateRole(ctx context.Context, name string, permissions []string) (types.Role, error) {
    trimmedName := strings.TrimSpace(name)
    if trimmedName == "" {
        return types.Role{}, domain.Validation("role name must not be empty")
    }
    if grants.IsPresetRole(trimmedName) {
        return types.Role{}, domain.NewDomainError(400, "role name already exists")
    }
    roles, err := s.d.Store.Org().Roles(ctx)
    if err != nil {
        return types.Role{}, err
    }
    for _, existing := range roles {
        if existing.Name == trimmedName {
            return types.Role{}, domain.NewDomainError(400, "role name already exists")
        }
    }
    grantIDs, err := s.d.Grants.NormalizeGrantIDs(permissions)
    if err != nil {
        return types.Role{}, domain.NewDomainError(400, err.Error())
    }
    role := types.Role{
        ID:          uuid.Must(uuid.NewV7()),
        Name:        trimmedName,
        Type:        "custom",
        Permissions: grantIDs,
    }
    roles = append(roles, role)
    if err := s.d.Store.Org().SetRoles(ctx, roles); err != nil {
        return types.Role{}, err
    }
    if err := core.BumpAuthzRevision(ctx, s.d); err != nil {
        return types.Role{}, err
    }
    return role, nil
}
```

### 6.3 `domain/org/structure/role_members.go`（改）

`AddRoleMember`：查找角色时需要在包含全局预设的列表中查找。当前 `ListRoles` 已返回全局 + 公司，无需额外改动逻辑。

### 6.4 `identity/credentials/service.go`（改）

```go
// 原：roleID := grants.PresetRoleID(companyID, grants.RolePlatformAdmin) 之类的
// 改：
role := types.Role{
    ID:   grants.IDPlatformAdmin,
    Name: grants.RolePlatformAdmin,
    Type: "preset",
}
```

注意：此处不再调用 `SetRoles` 写入预设角色行（已由 bootstrap seed 完成），仅需确保 `member_roles` 关联。

---

## 7. PDP（identity/authz）—— 无逻辑变更

`ResolveMemberPermissions(member, roles)` 逻辑不变：
- 预设角色按名称查 `manifest.PresetRoles` 展开（不看 `role.Permissions`）
- 自定义角色按 `role.Permissions`（`p-*`）展开

`GetMemberAuthz` 返回的 `store.MemberAuthz.Roles` 现在包含全局预设 + 公司自定义（来自改后的 `rolesForCompany`），PDP 消费方式不变。

---

## 8. API / 前端 —— 无变更

- 角色 CRUD API 接口签名不变
- 前端看到的响应结构不变（`Role { id, name, type, permissions, memberCount }`）
- 预设角色 ID 从哈希值变为固定常量，对前端透明

---

## 9. `Normalizer` 接口（简化）

`grants.Normalizer` 接口保留 `NormalizeGrantIDs`，但可删除 `RoleGrantIDs`：

- 预设角色权限不落 DB，不需要 `RoleGrantIDs` 来展开后写入
- `NormalizeGrantIDs` 仅服务自定义角色 CRUD

如果 `snapshot/org.go` 仍需要给 demo seed 构建权限，保留 `RoleGrantIDs` 也可以，不是必须删。

---

## 10. 删除清单

| 删除对象 | 文件 |
|---------|------|
| `role_permission_grants` 表 | `schema.sql` |
| `PresetRoleID(companyID, roleName)` 函数（旧双参数签名） | `domain/grants/roles.go` |
| `insertPresetRoles(ctx, exec, companyID)` | `seed/bootstrap/roles.go` |
| `reconcileCompanyPresetRoles` | `seed/bootstrap/roles.go` |
| `resolveGrantIDs` | `seed/bootstrap/roles.go` |
| `allPermissionIDs` | `seed/bootstrap/roles.go` |
| `ReconcilePresetRoles` | `seed/bootstrap/bootstrap.go` |
| `defaultCompanyRoles` | `domain/company/service_create.go` |
| `seed/init.go` 中 reconcile 步骤 | `seed/init.go` |
| 所有 `role_permission_grants` INSERT/DELETE/SELECT | `org_repo_roles.go`、`seed_core.go`、`budget_repo_alerts.go` |
| `IDRole1`–`IDRole6`（基于 `PresetRoleID` 哈希的旧 ID） | `seed/contract/ids.go` |

---

## 11. 实现步骤

1. **schema.sql**：改 `roles` 表 PK + 加 `permissions TEXT[]` + 加 partial unique index + 删 `role_permission_grants` + 改 FK（member_roles、alert_rule_notify_roles 均改为 ON DELETE RESTRICT）
2. **grants/roles.go**：加固定 UUID 常量（不含 BudgetApprover），改 `PresetRoleID` 签名为单参数，加 `IsPresetRole`
3. **seed/contract/ids.go**：预设角色 ID 改为引用 `grants.ID*`，BudgetApprover 用固定 seed UUID
4. **seed/bootstrap/**：重写 `roles.go`（`seedGlobalPresetRoles`），改 `bootstrap.go`、`org.go`
5. **seed/init.go**：删 reconcile
6. **seed/apply/seed_core.go**：删 `role_permission_grants` 写入，改 `insertSeedRoles`
7. **seed/snapshot/org.go**：改 `buildRoles` 用固定 ID + 自定义角色带 CompanyID
8. **store/postgres/org_repo_roles.go**：重写查询和 `SetRoles`
9. **store/postgres/org_repo_members.go**：改 JOIN 条件去掉 `r.company_id = mr.company_id`
10. **store/postgres/budget_repo_alerts.go**：增加 role_id 归属校验
11. **domain/company/service_create.go**：删 `defaultCompanyRoles`，`provisionCompany` 去掉 `SetRoles` 调用
12. **domain/org/structure/role_crud.go**：加 `IsPresetRole` 校验
13. **identity/credentials/service.go**：改 `IDPlatformAdmin`
14. **tests**：全部 `grants.PresetRoleID(companyID, ...)` 替换为 `grants.ID*`
15. **跑 `go test ./...`** 确认通过
