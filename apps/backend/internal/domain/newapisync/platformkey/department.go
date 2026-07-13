package platformkey

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func DepartmentIDForPlatformKey(key types.PlatformKey, budgetCtx pkgbudget.BudgetContext) string {
	if key.MemberID != nil {
		if member, ok := org.FindMemberByID(budgetCtx.Members, *key.MemberID); ok && member.DepartmentID != "" {
			return member.DepartmentID
		}
	}
	if key.ProjectID != nil {
		for _, project := range budgetCtx.Projects {
			if project.ID == *key.ProjectID && project.OwnerDepartmentID != "" {
				return project.OwnerDepartmentID
			}
		}
	}
	return ""
}

// TokenName is a unique NewAPI token name for create + lookup after POST /api/token/.
func TokenName(platformKeyID string) string {
	return fmt.Sprintf("tokenjoy:%s", platformKeyID)
}
