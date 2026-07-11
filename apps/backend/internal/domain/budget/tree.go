package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) GetTree(ctx context.Context) ([]types.BudgetNode, error) {
	return pkgbudget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes(), s.cfg.Clock())
}

func (s *service) UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetNode{}, err
	}
	var result types.BudgetNode
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		nodes, err := tx.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		tree := types.OrgNodesToBudgetTree(nodes)
		existing := pkgbudget.FindBudgetNode(tree, id)
		if existing == nil {
			return domain.NotFound("Node not found")
		}
		reserved := existing.ReservedPool
		if reservedPool != nil {
			reserved = reservedPool
		}
		reservedValue := 0.0
		if reserved != nil {
			reservedValue = *reserved
		}
		if msg := pkgbudget.ValidateBudgetNodeUpdate(tree, id, budget, reservedValue); msg != nil {
			return domain.Validation(*msg)
		}
		update := types.BudgetNode{Budget: budget, ReservedPool: reserved}
		if !pkgbudget.UpdateBudgetNodeInTree(tree, id, update) {
			return domain.NotFound("Node not found")
		}
		types.ApplyBudgetTreeToOrgNodes(nodes, tree)
		if err := tx.Org().Nodes().SetTree(ctx, nodes); err != nil {
			return fmt.Errorf("persist budget tree: %w", err)
		}
		updated := pkgbudget.FindBudgetNode(tree, id)
		result = *updated
		return nil
	})
	return result, err
}

func (s *service) ListMemberBudgets(ctx context.Context, deptID string) ([]types.MemberBudgetQuota, error) {
	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes(), s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	if pkgbudget.FindBudgetNode(tree, deptID) == nil {
		return nil, domain.NotFound("Department not found")
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	quotas := make([]types.MemberBudgetQuota, 0)
	for _, member := range members {
		if member.DepartmentID == deptID {
			quotas = append(quotas, pkgbudget.BuildMemberBudgetQuota(member, platformKeys))
		}
	}
	return quotas, nil
}

func (s *service) UpdateMemberBudget(ctx context.Context, memberID string, personalBudget float64) (types.MemberBudgetQuota, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.MemberBudgetQuota{}, err
	}
	var result types.MemberBudgetQuota
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, tx.BudgetSnapshots(), tx.Org().Nodes(), s.cfg.Clock())
		if err != nil {
			return err
		}
		members, err := tx.Org().Members(ctx)
		if err != nil {
			return err
		}
		platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, tx.BudgetSnapshots(), tx.Org(), tx.Budget(), tx.Keys(), s.cfg.Clock())
		if err != nil {
			return err
		}
		if msg := pkgbudget.ValidateMemberBudgetUpdate(tree, members, platformKeys, memberID, personalBudget); msg != nil {
			return domain.Validation(*msg)
		}
		r, updatedMembers := pkgbudget.ApplyMemberBudgetUpdate(members, platformKeys, memberID, personalBudget)
		if err := tx.Org().SetMembers(ctx, updatedMembers); err != nil {
			return fmt.Errorf("persist member personal budget: %w", err)
		}
		result = r
		return nil
	})
	return result, err
}

func (s *service) GetGroupMemberConsumed(ctx context.Context, groupID string) (map[string]float64, error) {
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return nil, err
	}
	var target *types.BudgetGroup
	for i := range groups {
		if groups[i].ID == groupID {
			target = &groups[i]
			break
		}
	}
	if target == nil {
		return nil, domain.NotFound("Group not found")
	}

	deptID := ""
	if len(target.DepartmentIDs) > 0 {
		deptID = target.DepartmentIDs[0]
	}
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, s.store.Org().Nodes(), deptID, s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	periodKey := open.String()

	result := make(map[string]float64)
	for _, memberID := range target.MemberIDs {
		consumed, _, err := s.store.BudgetSnapshots().GetConsumed(ctx, store.SnapshotAxisMember, memberID, periodKey)
		if err != nil {
			return nil, err
		}
		result[memberID] = consumed
	}
	return result, nil
}
