package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) ListAlerts(ctx context.Context) ([]types.AlertRule, error) {
	return s.store.Budget().AlertRules(ctx)
}

func (s *service) CreateAlert(ctx context.Context, rule types.AlertRule) (types.AlertRule, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.AlertRule{}, err
	}
	var result types.AlertRule
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		created := rule
		created.ID = generateBudgetID("alert")
		rules, err := tx.Budget().AlertRules(ctx)
		if err != nil {
			return err
		}
		rules = append(rules, created)
		if err := tx.Budget().SetAlertRules(ctx, rules); err != nil {
			return fmt.Errorf("persist alert rules: %w", err)
		}
		result = created
		return nil
	})
	if err == nil {
		s.logger.Info("budget.alert.created", "alert_id", result.ID, "node_id", result.NodeID)
	}
	return result, err
}

func (s *service) UpdateAlert(ctx context.Context, id string, patch types.AlertRule) (types.AlertRule, error) {
	var result types.AlertRule
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		rules, err := tx.Budget().AlertRules(ctx)
		if err != nil {
			return err
		}
		for i := range rules {
			if rules[i].ID == id {
				if patch.NodeID != "" {
					rules[i].NodeID = patch.NodeID
				}
				if patch.NodeName != "" {
					rules[i].NodeName = patch.NodeName
				}
				if patch.Thresholds != nil {
					rules[i].Thresholds = append([]int{}, patch.Thresholds...)
				}
				if patch.NotifyRoleIDs != nil {
					rules[i].NotifyRoleIDs = append([]string{}, patch.NotifyRoleIDs...)
				}
				rules[i].Enabled = patch.Enabled
				if err := tx.Budget().SetAlertRules(ctx, rules); err != nil {
					return fmt.Errorf("persist alert rules: %w", err)
				}
				result = rules[i]
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
	if err == nil {
		s.logger.Info("budget.alert.updated", "alert_id", id, "enabled", result.Enabled)
	}
	return result, err
}

func (s *service) DeleteAlert(ctx context.Context, id string) error {
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		rules, err := tx.Budget().AlertRules(ctx)
		if err != nil {
			return err
		}
		filtered := make([]types.AlertRule, 0, len(rules))
		found := false
		for _, rule := range rules {
			if rule.ID == id {
				found = true
				continue
			}
			filtered = append(filtered, rule)
		}
		if !found {
			return domain.NotFound("Not found")
		}
		if err := tx.Budget().SetAlertRules(ctx, filtered); err != nil {
			return fmt.Errorf("persist alert rules: %w", err)
		}
		return nil
	})
	if err == nil {
		s.logger.Info("budget.alert.deleted", "alert_id", id)
	}
	return err
}
