package budget

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) ListGroups(ctx context.Context) ([]types.BudgetGroup, error) {
	return pkgbudget.LoadBudgetGroupsWithConsumed(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.cfg.Clock())
}

func (s *service) CreateGroup(ctx context.Context, group types.BudgetGroup) (types.BudgetGroup, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetGroup{}, err
	}
	if strings.TrimSpace(group.Name) == "" {
		return types.BudgetGroup{}, domain.Validation("group name is required")
	}
	if len(group.Name) > 100 {
		return types.BudgetGroup{}, domain.Validation("group name must be 100 characters or less")
	}
	if group.Budget < 0 {
		return types.BudgetGroup{}, domain.Validation("budget must be non-negative")
	}
	var result types.BudgetGroup
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		groups, err := tx.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		trimmedName := strings.TrimSpace(group.Name)
		for _, existing := range groups {
			if existing.Name == trimmedName {
				return domain.Conflict("group name already exists")
			}
		}
		created := types.BudgetGroup{
			ID:   generateBudgetID("bg"),
			Name: trimmedName, Budget: group.Budget, Consumed: 0,
			MemberIDs:     append([]string{}, group.MemberIDs...),
			DepartmentIDs: append([]string{}, group.DepartmentIDs...),
		}
		groups = append(groups, created)
		if err := tx.Budget().SetGroups(ctx, groups); err != nil {
			return fmt.Errorf("persist budget groups: %w", err)
		}
		result = created
		return nil
	})
	if err == nil {
		s.logger.Info("budget.group.created", "group_id", result.ID, "name", result.Name, "budget", result.Budget)
	}
	return result, err
}

func (s *service) UpdateGroup(ctx context.Context, id string, patch types.BudgetGroup) (types.BudgetGroup, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetGroup{}, err
	}
	if patch.Name != "" && len(patch.Name) > 100 {
		return types.BudgetGroup{}, domain.Validation("group name must be 100 characters or less")
	}
	if patch.Budget < 0 {
		return types.BudgetGroup{}, domain.Validation("budget must be non-negative")
	}
	var result types.BudgetGroup
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		groups, err := tx.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		for i := range groups {
			if groups[i].ID == id {
				if patch.Name != "" {
					groups[i].Name = patch.Name
				}
				groups[i].Budget = patch.Budget
				if patch.MemberIDs != nil {
					groups[i].MemberIDs = append([]string{}, patch.MemberIDs...)
				}
				if patch.DepartmentIDs != nil {
					groups[i].DepartmentIDs = append([]string{}, patch.DepartmentIDs...)
				}
				if err := tx.Budget().SetGroups(ctx, groups); err != nil {
					return fmt.Errorf("persist budget groups: %w", err)
				}
				budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, tx.BudgetConsumed(), tx.Org(), tx.Budget(), tx.Keys(), s.cfg.Clock())
				if err != nil {
					return fmt.Errorf("load budget context: %w", err)
				}
				for _, group := range budgetCtx.Groups {
					if group.ID == id {
						result = group
						return nil
					}
				}
				result = groups[i]
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
	if err == nil {
		s.logger.Info("budget.group.updated", "group_id", id, "name", result.Name, "budget", result.Budget)
	}
	return result, err
}

func (s *service) DeleteGroup(ctx context.Context, id string) error {
	var deletedMemberIDs []string
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		groups, err := tx.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		for i := range groups {
			if groups[i].ID == id {
				deletedMemberIDs = append([]string{}, groups[i].MemberIDs...)
				groups = append(groups[:i], groups[i+1:]...)
				if err := tx.Budget().SetGroups(ctx, groups); err != nil {
					return fmt.Errorf("persist budget groups: %w", err)
				}
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
	if err == nil {
		s.logger.Info("budget.group.deleted", "group_id", id)
		// Enqueue rebalance for affected members so their keys get updated quotas
		if s.enqueueRebalanceAxis != nil {
			for _, memberID := range deletedMemberIDs {
				if rebalErr := s.enqueueRebalanceAxis(ctx, store.RebalanceAxisMember, memberID); rebalErr != nil {
					s.logger.Error("enqueue rebalance failed after group delete",
						"group_id", id, "member_id", memberID, "error", rebalErr)
				}
			}
		}
	}
	return err
}
