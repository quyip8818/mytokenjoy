package snapshot

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

// BuildEmpty returns a minimal snapshot: company + admin + root department only.
// Used by SEED=empty to create a blank workspace for manual testing.
func BuildEmpty(cfg config.Config) store.Snapshot {
	admin := types.Member{
		ID: contract.IDMemberAdmin, CompanyID: contract.DefaultCompanyID,
		Alias: "管理员", Email: "admin@example.com",
		DepartmentID: contract.IDDept1, DepartmentName: "总公司",
		Status: "active", Roles: []string{permission.RoleSuperAdmin}, Source: "manual",
	}
	members := []types.Member{admin}
	roles := buildRoles(members)
	rootNode := types.OrgNode{
		ID:   contract.IDDept1,
		Name: "总公司",
	}
	return store.Snapshot{
		Company:       defaultCompany(cfg),
		OrgNodes:      []types.OrgNode{rootNode},
		Members:       members,
		Roles:         roles,
		Permissions:   buildPermissions(),
		Models:        buildModels(),
		OverrunPolicy: buildOverrunPolicy(),
		SeedAt:        clock.NowUTC(cfg.Clock()),
	}
}
