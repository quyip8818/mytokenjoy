package budget

import (
	"math"
	"slices"

	"github.com/tokenjoy/backend/internal/domain/types"
)

// budgetEpsilon is the tolerance for floating-point budget comparisons.
// Values within this range of zero are treated as zero.
const budgetEpsilon = 0.001

// ClampNonNegative returns 0 if v is negative or within epsilon of zero.
func ClampNonNegative(v float64) float64 {
	if v < budgetEpsilon {
		return 0
	}
	return v
}

// BudgetExhausted returns true if consumed >= budget within floating-point tolerance.
func BudgetExhausted(consumed, budget float64) bool {
	return consumed >= budget-budgetEpsilon
}

// NearZero returns true if the absolute value is within epsilon.
func NearZero(v float64) bool {
	return math.Abs(v) < budgetEpsilon
}

type MemberAxisInput struct {
	Skip     bool
	Cap      float64
	Consumed float64
}

type DeptAxisInput struct {
	Budget   float64
	Consumed float64
	Reserved float64
}

// ComputeRemainBudget returns the effective remaining budget for a platform key as the
// minimum of key, optional project or member, and department caps.
// memberAxis nil uses summed platform-key usage (NewAPI sync).
// deptAxis nil uses budget tree nodes (NewAPI sync); non-nil uses explicit dept snapshot values.
func ComputeRemainBudget(
	key types.PlatformKey,
	tree []types.BudgetNode,
	members []types.Member,
	platformKeys []types.PlatformKey,
	projects []types.Project,
	departmentID string,
	memberAxis *MemberAxisInput,
	deptAxis *DeptAxisInput,
) float64 {
	candidates := make([]float64, 0, 4)

	if key.Budget > 0 {
		keyRemaining := key.Budget - key.Consumed
		if keyRemaining < 0 {
			keyRemaining = 0
		}
		candidates = append(candidates, keyRemaining)
	}

	if key.ProjectID != nil {
		for _, project := range projects {
			if project.ID == *key.ProjectID {
				projectRemaining := project.Budget - project.Consumed
				if projectRemaining < 0 {
					projectRemaining = 0
				}
				candidates = append(candidates, projectRemaining)
				break
			}
		}
	} else if key.MemberID != nil {
		switch {
		case memberAxis != nil && memberAxis.Skip:
		case memberAxis != nil:
			memberRemaining := memberAxis.Cap - memberAxis.Consumed
			if memberRemaining < 0 {
				memberRemaining = 0
			}
			candidates = append(candidates, memberRemaining)
		default:
			memberUsed := GetConsumedKeyBudget(platformKeys, *key.MemberID)
			memberCap := GetPersonalBudget(members, *key.MemberID)
			memberRemaining := memberCap - memberUsed
			if memberRemaining < 0 {
				memberRemaining = 0
			}
			candidates = append(candidates, memberRemaining)
		}
	}

	if deptAxis != nil {
		deptRemaining := deptAxis.Budget - deptAxis.Consumed - deptAxis.Reserved
		if deptRemaining < 0 {
			deptRemaining = 0
		}
		candidates = append(candidates, deptRemaining)
	} else if node := FindBudgetNode(tree, departmentID); node != nil {
		deptRemaining := node.Budget - node.Consumed
		reserved := 0.0
		if node.ReservedPool != nil {
			reserved = *node.ReservedPool
		}
		deptRemaining -= reserved
		if deptRemaining < 0 {
			deptRemaining = 0
		}
		candidates = append(candidates, deptRemaining)
	}

	if len(candidates) == 0 {
		return 0
	}
	return slices.Min(candidates)
}
