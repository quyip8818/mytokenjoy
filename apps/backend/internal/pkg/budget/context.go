package budget

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

type BudgetContext struct {
	Tree         []types.BudgetNode
	Members      []types.Member
	PlatformKeys []types.PlatformKey
	Projects     []types.Project
}

func LoadBudgetContext(
	ctx context.Context,
	snapshots store.BudgetConsumedRepository,
	org store.OrgRepository,
	budgetRepo store.BudgetRepository,
	keys store.KeysRepository,
	clk clock.Clock,
) (BudgetContext, error) {
	tree, err := LoadBudgetTreeWithConsumed(ctx, snapshots, org.Nodes(), clk)
	if err != nil {
		return BudgetContext{}, err
	}
	members, err := org.Members(ctx)
	if err != nil {
		return BudgetContext{}, err
	}
	platformKeys, err := LoadPlatformKeysWithUsed(ctx, snapshots, org, budgetRepo, keys, clk)
	if err != nil {
		return BudgetContext{}, err
	}
	groups, err := LoadProjectsWithConsumed(ctx, snapshots, org, budgetRepo, clk)
	if err != nil {
		return BudgetContext{}, err
	}
	return BudgetContext{
		Tree:         tree,
		Members:      members,
		PlatformKeys: platformKeys,
		Projects:     groups,
	}, nil
}

func (c BudgetContext) FindPlatformKey(id string) (types.PlatformKey, bool) {
	for _, key := range c.PlatformKeys {
		if key.ID == id {
			return key, true
		}
	}
	return types.PlatformKey{}, false
}

func (c BudgetContext) ComputeRemain(
	key types.PlatformKey,
	departmentID string,
	memberAxis *MemberAxisInput,
	deptAxis *DeptAxisInput,
) float64 {
	return ComputeRemainBudget(key, c.Tree, c.Members, c.PlatformKeys, c.Projects, departmentID, memberAxis, deptAxis)
}
