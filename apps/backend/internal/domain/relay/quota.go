package relay

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
	"github.com/tokenjoy/backend/internal/pkg/memberquota"
)

func ComputeRemainQuotaCNY(
	key types.PlatformKey,
	tree []types.BudgetNode,
	pools map[string]types.MemberQuotaPool,
	platformKeys []types.PlatformKey,
	groups []types.BudgetGroup,
	departmentID string,
) float64 {
	keyRemaining := key.Quota - key.Used
	if keyRemaining < 0 {
		keyRemaining = 0
	}
	candidates := []float64{keyRemaining}

	if key.BudgetGroupID != nil {
		for _, group := range groups {
			if group.ID == *key.BudgetGroupID {
				bgRemaining := group.Budget - group.Consumed
				if bgRemaining < 0 {
					bgRemaining = 0
				}
				candidates = append(candidates, bgRemaining)
				break
			}
		}
	} else if key.MemberID != nil {
		memberUsed := memberquota.GetUsedKeyQuota(platformKeys, *key.MemberID)
		memberCap := memberquota.GetPersonalQuota(pools, *key.MemberID)
		memberRemaining := memberCap - memberUsed
		if memberRemaining < 0 {
			memberRemaining = 0
		}
		candidates = append(candidates, memberRemaining)
	}

	if node := budgetutil.FindBudgetNode(tree, departmentID); node != nil {
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

	min := candidates[0]
	for _, value := range candidates[1:] {
		if value < min {
			min = value
		}
	}
	return min
}
