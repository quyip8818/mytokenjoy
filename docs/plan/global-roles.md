# 全局角色 + 公司自定义角色 实现方案

> **背景**：当前预设角色按公司维度生成（每个公司一份独立 ID），导致重复数据和不必要的复杂度。项目未上线，无迁移负担。  
> **目标**：预设角色改为全局唯一，所有公司共享同一份角色定义；公司可创建自定义角色。

---

## 1. 设计决策

| 决策 | 方案 |
|------|------|
| 预设角色 ID | 固定 UUID，写死在代码中，所有公司共享 |
| 预设角色权限 | 仅由 `manifest.json` 定义，不存 DB 的 `role_permission_grants` |
| 自定义角色 ID | 公司创建时生成 `uuid.NewV7()`，`company_id` 隔离 |
| 角色归属 | `roles.company_id` 对预设角色为 `NULL`（全局），对自定义角色为具体公司 |
| 删除 `deterministicRoleID` | 不再需要哈希生成 ID |

---

## 2. 数据模型变更

### 2.1 `roles` 表

```sql
CREATE TABLE IF NOT EXISTS roles (
    id         UUID PRIMARY KEY,                    -- 改为单列 PK
    company_id UUID REFERENCES companies (id) ON DELETE CASCADE,  -- NULL = 全局预设
    name       TEXT NOT NULL,
    type       TEXT NOT NULL,                       -- 'preset' | 'custom'
    UNIQUE (company_id, name)                      -- 同公司内名称唯一；全局角色 company_id=NULL 也唯一
);
```

**变更点**：
- PK 从 `(company_id, id)` 改为 `(id)`
- `company_id` 允许 NULL
- UNIQUE 约束 `(company_id, name)` 保留（NULL company_id 下角色名仍唯一）

### 2.2 `role_permission_grants` 表

```sql
CREATE TABLE IF NOT EXISTS role_permission_grants (
    role_id       UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions (id) ON DELETE RESTRICT,
    PRIMARY KEY (role_id, permission_id)
);
```

**变更点**：
- 移除 `company_id` 列（角色 ID 已全局唯一，无需复合键）
- 预设角色不写入此表（权限由 manifest 运行时展开）
- 仅自定义角色在此表存储 `p-*` 权限 ID

### 2.3 `member_roles` 表

```sql
CREATE TABLE IF NOT EXISTS member_roles (
    company_id UUID NOT NULL,
    member_id  UUID NOT NULL,
    role_id    UUID NOT NULL REFERENCES roles (id),
    PRIMARY KEY (company_id, member_id, role_id)
);
```

**变更点**：
- `role_id` FK 改为直接引用 `roles (id)`，不再需要 `(company_id, role_id)` 复合 FK

---

## 3. 全局预设角色定义

### 3.1 固定 UUID

在 `internal/domain/grants/roles.go` 中硬编码：

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

// 全局预设角色 UUID —— 所有公司共享，永不变更。
var (
    IDSuperAdmin     = uuid.MustParse("00000000-0000-0000-0000-000000000001")
    IDOrgAdmin       = uuid.MustParse("00000000-0000-0000-0000-000000000002")
    IDMember         = uuid.MustParse("00000000-0000-0000-0000-000000000003")
    IDAuditor        = uuid.MustParse("00000000-0000-0000-0000-000000000004")
    IDAPICaller      = uuid.MustParse("00000000-0000-0000-0000-000000000005")
    IDBudgetApprover = uuid.MustParse("00000000-0000-0000-0000-000000000006")
    IDPlatformAdmin  = uuid.MustParse("00000000-0000-0000-0000-000000000007")
)

// PresetRolesByName 查询表：名称 → ID
var PresetRolesByName = map[string]uuid.UUID{
    RoleSuperAdmin:     IDSuperAdmin,
    RoleOrgAdmin:       IDOrgAdmin,
    RoleMember:         IDMember,
    RoleAuditor:        IDAuditor,
    RoleAPICaller:      IDAPICaller,
    RoleBudgetApprover: IDBudgetApprover,
    RolePlatformAdmin:  IDPlatformAdmin,
}

// IsPresetRole 判断一个角色名是否为全局预设
func IsPresetRole(name string) bool {
    _, ok := PresetRolesByName[name]
    return ok
}
```

### 3.2 删除 `PresetRoleID` 函数

`deterministicRoleID` / `PresetRoleID` 整体删除，全部替换为 `PresetRolesByName[roleName]` 查表。

---

## 4. Bootstrap / Seed 变更

### 4.1 全局预设角色 seed

只需执行一次（启动时幂等），不绑定任何公司：

```go
func seedGlobalPresetRoles(ctx context.Context, exec TableWriter) error {
    for name, id := range grants.PresetRolesByName {
        if _, err := exec.Exec(ctx, `
            INSERT INTO roles (id, company_id, name, type) VALUES ($1, NULL, $2, 'preset')
            ON CONFLICT (id) DO NOTHING
        `, id, name); err != nil {
            return fmt.Errorf("seed preset role %s: %w", name, err)
        }
    }
    return nil
}
```

### 4.2 创建公司时不再插入预设角色

`provisionCompany` 中删除 `defaultCompanyRoles` 调用和 `SetRoles` 预设部分。公司创建时仅处理自定义角色（如果有）。

### 4.3 `ReconcilePresetRoles` 简化

不再遍历每个公司做 reconcile，预设角色权限由 manifest 运行时展开，DB 无 `role_permission_grants` 行。删除整个 reconcile 逻辑。

---

## 5. Domain / PDP 变更

### 5.1 `ListRoles`

查询改为：返回全局预设角色 + 当前公司自定义角色。

```go
func (s *LocalService) ListRoles(ctx context.Context) ([]types.Role, error) {
    // 全局预设 + 本公司自定义（SQL: WHERE company_id IS NULL OR company_id = $1）
    return s.d.Store.Org().Roles(ctx)
}
```

Store 层 SQL：

```sql
SELECT id, company_id, name, type FROM roles
WHERE company_id IS NULL OR company_id = $1
ORDER BY type ASC, name ASC;
```

### 5.2 `CreateRole`（不变，仅自定义）

创建自定义角色时 `company_id` 取当前租户，`type = 'custom'`。命名冲突校验包含全局角色名（禁止自定义角色与预设角色重名）。

### 5.3 `UpdateRole` / `DeleteRole`

保持 `type == "preset"` 拒绝修改/删除的逻辑不变。

### 5.4 权限解析（`ResolveMemberPermissions`）

逻辑不变：预设角色按名称走 `manifest.PresetRoles` 展开，自定义角色按 `role.Permissions`（`p-*`）展开。

### 5.5 `member.Roles` 字段

保持 `[]string`（角色名列表）。预设角色名全局唯一、自定义角色名公司内唯一，不会冲突。

---

## 6. Store 层变更

### 6.1 `OrgRepo.Roles(ctx)` 查询

```go
func (r *pgOrgRepo) Roles(ctx context.Context) ([]types.Role, error) {
    companyID := ctxcompany.MustID(ctx)
    rows, err := r.pool.Query(ctx, `
        SELECT id, company_id, name, type FROM roles
        WHERE company_id IS NULL OR company_id = $1
    `, companyID)
    // ...
}
```

### 6.2 `SetRoles`

仅操作当前公司的自定义角色（`company_id = $1`），不触碰全局预设行。

### 6.3 `member_roles` 插入

插入时 `role_id` 直接用全局预设 UUID 或自定义角色 UUID。

---

## 7. 受影响的代码清单

| 文件 | 变更 |
|------|------|
| `internal/domain/grants/roles.go` | 加固定 UUID 常量，删 `PresetRoleID` |
| `internal/store/postgres/schema.sql` | 表结构调整 |
| `internal/store/postgres/org_repo_roles.go` | 查询改 `company_id IS NULL OR` |
| `internal/domain/company/service_create.go` | 删 `defaultCompanyRoles`，不再 `SetRoles` 预设 |
| `seed/bootstrap/roles.go` | 改为 `seedGlobalPresetRoles`，删 reconcile |
| `seed/bootstrap/bootstrap.go` | 调整调用 |
| `seed/bootstrap/org.go` | `superAdminRoleID` 改为 `grants.IDSuperAdmin` |
| `seed/init.go` | 删 `ReconcilePresetRoles` |
| `seed/snapshot/org.go` | 用固定 ID |
| `internal/domain/org/structure/role_crud.go` | 名称冲突校验含全局角色 |
| `internal/identity/credentials/service.go` | 平台管理员用 `grants.IDPlatformAdmin` |
| `tests/` | 全部预设角色 ID 改为 `grants.ID*` 常量 |

---

## 8. 查询示例

```sql
-- 获取某公司成员可见的全部角色
SELECT r.id, r.name, r.type,
       COUNT(mr.member_id) AS member_count
FROM roles r
LEFT JOIN member_roles mr ON mr.role_id = r.id AND mr.company_id = $1
WHERE r.company_id IS NULL OR r.company_id = $1
GROUP BY r.id, r.name, r.type;

-- 获取成员的角色列表（含全局预设）
SELECT r.name FROM member_roles mr
JOIN roles r ON r.id = mr.role_id
WHERE mr.company_id = $1 AND mr.member_id = $2;
```

---

## 9. 不变的部分

- `manifest.json` 结构不变（`presetRoles` 仍按名称定义权限）
- `ResolveMemberPermissions` 逻辑不变
- `authz_revision` 机制不变
- 前端 Session / Permission 体系不变
- JWT 结构不变
- PEP middleware 不变
- `member.Roles` 仍为 `[]string`（角色名）

---

## 10. 优势

1. **减少重复**：预设角色全局一行，不再每个公司一份
2. **简化 seed**：删除 reconcile 逻辑
3. **ID 可预测**：测试中直接用常量，无需计算哈希
4. **概念清晰**：全局 vs 公司自定义，一目了然
5. **无迁移成本**：项目未上线，直接改 schema
