package budget

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
)

func cloneMemberBudgets(src map[uuid.UUID]int64) map[uuid.UUID]int64 {
	if len(src) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]int64, len(src))
	for memberID, budget := range src {
		out[memberID] = budget
	}
	return out
}

func pruneMemberBudgets(budgets map[uuid.UUID]int64, roster []uuid.UUID) map[uuid.UUID]int64 {
	if len(budgets) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]int64)
	for _, memberID := range roster {
		if budget, ok := budgets[memberID]; ok {
			out[memberID] = budget
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func validateProjectMemberBudgets(projectBudget int64, roster []uuid.UUID, budgets map[uuid.UUID]int64) error {
	if len(budgets) == 0 {
		return nil
	}
	rosterSet := make(map[uuid.UUID]struct{}, len(roster))
	for _, id := range roster {
		rosterSet[id] = struct{}{}
	}
	var sum int64
	for memberID, budget := range budgets {
		if budget < 0 {
			return domain.Validation("member budget must be non-negative")
		}
		if _, ok := rosterSet[memberID]; !ok {
			return domain.Validation("member budget must belong to project roster")
		}
		sum += budget
	}
	if sum > projectBudget {
		return domain.Validation("member budgets exceed project budget")
	}
	return nil
}

func mergeMemberBudgetPatch(existing map[uuid.UUID]int64, patch map[uuid.UUID]int64, roster []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(patch) == 0 {
		return existing, nil
	}
	if existing == nil {
		existing = make(map[uuid.UUID]int64)
	}
	rosterSet := make(map[uuid.UUID]struct{}, len(roster))
	for _, id := range roster {
		rosterSet[id] = struct{}{}
	}
	for memberID, budget := range patch {
		if budget < 0 {
			return nil, domain.Validation("member budget must be non-negative")
		}
		if _, ok := rosterSet[memberID]; !ok {
			return nil, domain.Validation("member not on project roster")
		}
		existing[memberID] = budget
	}
	return existing, nil
}

func validateOwnerInMembers(ownerID *uuid.UUID, memberIDs []uuid.UUID) error {
	if ownerID == nil {
		return nil
	}
	for _, mid := range memberIDs {
		if mid == *ownerID {
			return nil
		}
	}
	return domain.Validation("project owner must be a member of the project")
}
