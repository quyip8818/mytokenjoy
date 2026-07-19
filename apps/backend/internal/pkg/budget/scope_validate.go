package budget

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func ValidatePlatformKeyScope(scope string, memberID, projectID *uuid.UUID) error {
	if !types.ValidPlatformKeyScope(scope) {
		return fmt.Errorf("invalid scope %q", scope)
	}
	switch scope {
	case types.PlatformKeyScopeMember:
		if memberID == nil {
			return fmt.Errorf("memberId required for member scope")
		}
		if projectID != nil {
			return fmt.Errorf("projectId not allowed for member scope")
		}
	case types.PlatformKeyScopeProject:
		if projectID == nil {
			return fmt.Errorf("projectId required for project scope")
		}
	case types.PlatformKeyScopeProjectMember:
		if memberID == nil || projectID == nil {
			return fmt.Errorf("memberId and projectId required for project_member scope")
		}
	}
	return nil
}

func ValidateProjectMemberRoster(project types.Project, memberID uuid.UUID) error {
	for _, id := range project.MemberIDs {
		if id == memberID {
			budget := project.MemberBudgets[memberID]
			if budget <= 0 {
				return fmt.Errorf("member_budget must be > 0 for project_member keys")
			}
			return nil
		}
	}
	return fmt.Errorf("member not on project roster")
}

func ProjectMemberBudget(project types.Project, memberID uuid.UUID) int64 {
	if project.MemberBudgets == nil {
		return 0
	}
	return project.MemberBudgets[memberID]
}

func ValidateProjectMemberKeyBudget(
	project types.Project,
	platformKeys []types.PlatformKey,
	memberID uuid.UUID,
	budget int64,
	excludeKeyID uuid.UUID,
) *string {
	subCap := ProjectMemberBudget(project, memberID)
	var allocated int64
	for _, key := range platformKeys {
		if key.Scope != types.PlatformKeyScopeProjectMember {
			continue
		}
		if key.ProjectID == nil || *key.ProjectID != project.ID {
			continue
		}
		if key.MemberID == nil || *key.MemberID != memberID {
			continue
		}
		if key.Status != "active" {
			continue
		}
		if excludeKeyID != uuid.Nil && key.ID == excludeKeyID {
			continue
		}
		allocated += key.Budget
	}
	remaining := subCap - allocated
	if budget > remaining {
		display := remaining
		if display < 0 {
			display = 0
		}
		msg := fmt.Sprintf("成员项目子额度剩余约 %d quota", display)
		return &msg
	}
	return nil
}

func FindProject(projects []types.Project, projectID uuid.UUID) (types.Project, bool) {
	for _, project := range projects {
		if project.ID == projectID {
			return project, true
		}
	}
	return types.Project{}, false
}

func MemberDepartmentID(members []types.Member, memberID uuid.UUID) uuid.UUID {
	if member, ok := pkgorg.FindMemberByID(members, memberID); ok {
		return member.DepartmentID
	}
	return uuid.Nil
}

func ValidateMemberScopeKeyBudget(
	members []types.Member,
	platformKeys []types.PlatformKey,
	memberID uuid.UUID,
	budget int64,
	excludeKeyID uuid.UUID,
) *string {
	if budget > memberScopeBudgetRemaining(members, platformKeys, memberID, excludeKeyID) {
		msg := "额度不足，请先申请追加"
		return &msg
	}
	return nil
}

func ValidateProjectScopeKeyBudget(
	scope string,
	project types.Project,
	platformKeys []types.PlatformKey,
	memberID *uuid.UUID,
	budget int64,
	excludeKeyID uuid.UUID,
) *string {
	if scope == types.PlatformKeyScopeProjectMember {
		if memberID == nil {
			msg := "memberId required for project_member scope"
			return &msg
		}
		if msg := ValidateProjectMemberKeyBudget(project, platformKeys, *memberID, budget, excludeKeyID); msg != nil {
			return msg
		}
	}
	return ValidateProjectKeyBudget(project, platformKeys, budget, excludeKeyID)
}
