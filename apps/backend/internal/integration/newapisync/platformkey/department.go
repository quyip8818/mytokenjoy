package platformkey

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func DepartmentIDForPlatformKey(key types.PlatformKey, budgetCtx pkgbudget.BudgetContext) uuid.UUID {
	if key.MemberID != nil {
		if member, ok := org.FindMemberByID(budgetCtx.Members, *key.MemberID); ok && member.DepartmentID != uuid.Nil {
			return member.DepartmentID
		}
	}
	if key.ProjectID != nil {
		for _, project := range budgetCtx.Projects {
			if project.ID == *key.ProjectID && project.OwnerDepartmentID != uuid.Nil {
				return project.OwnerDepartmentID
			}
		}
	}
	return uuid.Nil
}

// TokenName is a unique NewAPI token name for create + lookup after POST /api/token/.
func TokenName(platformKeyID uuid.UUID) string {
	return fmt.Sprintf("tokenjoy:%s", platformKeyID)
}
