package budget

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
)

func cloneMemberBudgets(src map[uuid.UUID]float64) map[uuid.UUID]float64 {
	if len(src) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]float64, len(src))
	for memberID, budget := range src {
		out[memberID] = budget
	}
	return out
}

func pruneMemberBudgets(budgets map[uuid.UUID]float64, roster []uuid.UUID) map[uuid.UUID]float64 {
	if len(budgets) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]float64)
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

func validateProjectMemberBudgets(projectBudget float64, roster []uuid.UUID, budgets map[uuid.UUID]float64) error {
	if len(budgets) == 0 {
		return nil
	}
	rosterSet := make(map[uuid.UUID]struct{}, len(roster))
	for _, id := range roster {
		rosterSet[id] = struct{}{}
	}
	var sum float64
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

func mergeMemberBudgetPatch(existing map[uuid.UUID]float64, patch map[uuid.UUID]float64, roster []uuid.UUID) (map[uuid.UUID]float64, error) {
	if len(patch) == 0 {
		return existing, nil
	}
	if existing == nil {
		existing = make(map[uuid.UUID]float64)
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
