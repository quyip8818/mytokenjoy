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

// PresetRoleID produces a stable deterministic UUID for a preset role within a company.
// All code paths (provisioning, seed, bootstrap) MUST use this to avoid ID conflicts.
func PresetRoleID(companyID uuid.UUID, roleName string) uuid.UUID {
	return uuid.NewSHA1(companyID, []byte("preset-role:"+roleName))
}
