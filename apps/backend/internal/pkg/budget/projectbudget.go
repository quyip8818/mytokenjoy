package budget

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/exchange"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func ResolveProjectPeriodKeys(
	project types.Project,
	members []types.Member,
	deptPeriod map[uuid.UUID]string,
	rootPeriodKey string,
	at time.Time,
) []string {
	deptIDs := make([]uuid.UUID, 0, 1+len(project.MemberIDs))
	if project.OwnerDepartmentID != uuid.Nil {
		deptIDs = append(deptIDs, project.OwnerDepartmentID)
	}
	for _, memberID := range project.MemberIDs {
		if member, ok := pkgorg.FindMemberByID(members, memberID); ok && member.DepartmentID != uuid.Nil {
			deptIDs = append(deptIDs, member.DepartmentID)
		}
	}
	deptIDs = uniqueUUIDs(deptIDs)
	keys := make([]string, 0, len(deptIDs))
	for _, deptID := range deptIDs {
		if orgPeriod, ok := deptPeriod[deptID]; ok {
			keys = append(keys, SnapshotKey(orgPeriod, at))
		}
	}
	keys = uniqueStrings(keys)
	if len(keys) == 0 {
		return []string{rootPeriodKey}
	}
	return keys
}

func GetAllocatedProjectKeyBudget(platformKeys []types.PlatformKey, projectID uuid.UUID) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.ProjectID != nil && *key.ProjectID == projectID && key.Status == "active" {
			sum += key.Budget
		}
	}
	return sum
}

func GetProjectBudgetRemaining(project types.Project, platformKeys []types.PlatformKey) float64 {
	allocated := GetAllocatedProjectKeyBudget(platformKeys, project.ID)
	remaining := project.Budget - project.Consumed - allocated
	if remaining < 0 {
		return 0
	}
	return remaining
}

func ValidateProjectKeyBudget(project types.Project, platformKeys []types.PlatformKey, budget float64, excludeKeyID uuid.UUID) *string {
	allocated := 0.0
	for _, key := range platformKeys {
		if key.ProjectID != nil && *key.ProjectID == project.ID && key.Status == "active" {
			if excludeKeyID != uuid.Nil && key.ID == excludeKeyID {
				continue
			}
			allocated += key.Budget
		}
	}
	remaining := project.Budget - project.Consumed - allocated
	if budget > remaining {
		display := remaining
		if display < 0 {
			display = 0
		}
		msg := fmt.Sprintf("项目剩余可分配额度约 %s", exchange.Format(display))
		return &msg
	}
	return nil
}
