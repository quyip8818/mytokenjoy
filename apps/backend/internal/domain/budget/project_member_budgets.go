package budget

import "github.com/tokenjoy/backend/internal/domain"

func cloneMemberBudgets(src map[string]float64) map[string]float64 {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]float64, len(src))
	for memberID, budget := range src {
		out[memberID] = budget
	}
	return out
}

func pruneMemberBudgets(budgets map[string]float64, roster []string) map[string]float64 {
	if len(budgets) == 0 {
		return nil
	}
	out := make(map[string]float64)
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

func memberOnRoster(roster []string, memberID string) bool {
	for _, id := range roster {
		if id == memberID {
			return true
		}
	}
	return false
}

func validateProjectMemberBudgets(roster []string, budgets map[string]float64) error {
	if len(budgets) == 0 {
		return nil
	}
	for memberID, budget := range budgets {
		if budget < 0 {
			return domain.Validation("member budget must be non-negative")
		}
		if !memberOnRoster(roster, memberID) {
			return domain.Validation("member budget must belong to project roster")
		}
	}
	return nil
}

func mergeMemberBudgetPatch(existing map[string]float64, patch map[string]float64, roster []string) (map[string]float64, error) {
	if len(patch) == 0 {
		return existing, nil
	}
	if existing == nil {
		existing = make(map[string]float64)
	}
	for memberID, budget := range patch {
		if budget < 0 {
			return nil, domain.Validation("member budget must be non-negative")
		}
		if !memberOnRoster(roster, memberID) {
			return nil, domain.Validation("member not on project roster")
		}
		existing[memberID] = budget
	}
	return existing, nil
}
