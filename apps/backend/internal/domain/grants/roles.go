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
// RoleBudgetApprover 不在此列——它是各公司按需创建的自定义角色。
var (
	IDSuperAdmin    = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	IDOrgAdmin      = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	IDMember        = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	IDAuditor       = uuid.MustParse("00000000-0000-0000-0000-000000000004")
	IDAPICaller     = uuid.MustParse("00000000-0000-0000-0000-000000000005")
	IDPlatformAdmin = uuid.MustParse("00000000-0000-0000-0000-000000000006")
)

// PresetRoles 名称→ID（仅全局预设角色）。
var PresetRoles = map[string]uuid.UUID{
	RoleSuperAdmin:    IDSuperAdmin,
	RoleOrgAdmin:      IDOrgAdmin,
	RoleMember:        IDMember,
	RoleAuditor:       IDAuditor,
	RoleAPICaller:     IDAPICaller,
	RolePlatformAdmin: IDPlatformAdmin,
}

// PresetRoleID 按名称查全局预设角色 ID。
func PresetRoleID(name string) uuid.UUID {
	return PresetRoles[name]
}

// IsPresetRole 判断角色名是否为全局预设角色。
func IsPresetRole(name string) bool {
	_, ok := PresetRoles[name]
	return ok
}
