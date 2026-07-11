package budget

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/notification"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type overrunPayload struct {
	DepartmentID  string  `json:"departmentId"`
	MemberID      *string `json:"memberId,omitempty"`
	BudgetGroupID *string `json:"budgetGroupId,omitempty"`
	PlatformKeyID string  `json:"platformKeyId"`
}

type OverrunService struct {
	cfg        config.Config
	store      store.Store
	keyControl newapisync.OverrunKeyControl
	notifier   notification.Notifier
	logger     *slog.Logger
}

func NewOverrunService(
	cfg config.Config,
	st store.Store,
	keyControl newapisync.OverrunKeyControl,
	notifier notification.Notifier,
	logger *slog.Logger,
) *OverrunService {
	return &OverrunService{
		cfg: cfg, store: st, keyControl: keyControl, notifier: notifier, logger: logger,
	}
}

func (s *OverrunService) ProcessOverrunPayload(ctx context.Context, raw json.RawMessage) error {
	var payload overrunPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	return s.evaluateOverrun(ctx, payload)
}

func (s *OverrunService) evaluateOverrun(ctx context.Context, payload overrunPayload) error {
	if s.keyControl == nil || !s.keyControl.Enabled() {
		return nil
	}

	type disableAction struct {
		scope   string
		target  string
		keys    func(ctx context.Context) error
		payload map[string]any
	}

	var actions []disableAction

	// Read all budget/consumed values inside a transaction with advisory lock
	// to get a consistent snapshot. Key disabling happens after the transaction.
	if err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}

		if payload.MemberID != nil && payload.BudgetGroupID == nil {
			used, err := tx.Keys().SumMemberKeyUsed(ctx, *payload.MemberID, s.cfg.Clock())
			if err != nil {
				return err
			}
			capacity, found, err := tx.Org().MemberPersonalBudget(ctx, *payload.MemberID)
			if err != nil {
				return err
			}
			if found && used >= capacity {
				memberID := *payload.MemberID
				actions = append(actions, disableAction{
					scope:  "member",
					target: memberID,
					keys:   func(ctx context.Context) error { return s.disableMemberKeys(ctx, memberID) },
					payload: map[string]any{
						"scope": "member", "memberId": memberID, "used": used, "capacity": capacity,
					},
				})
				return nil
			}
		}

		open, err := pkgbudget.OpenDepartmentPeriod(ctx, tx.Org().Nodes(), payload.DepartmentID, s.cfg.Clock())
		if err != nil {
			return err
		}
		snapshotPeriod := open.String()

		budgetAmount, found, err := tx.Org().Nodes().GetNodeBudget(ctx, payload.DepartmentID)
		if err != nil {
			return err
		}
		consumed, _, err := tx.BudgetSnapshots().GetConsumed(ctx, store.SnapshotAxisOrgNode, payload.DepartmentID, snapshotPeriod)
		if err != nil {
			return err
		}
		if found && consumed >= budgetAmount {
			deptID := payload.DepartmentID
			actions = append(actions, disableAction{
				scope:  "department",
				target: deptID,
				keys:   func(ctx context.Context) error { return s.disableDepartmentKeys(ctx, deptID) },
				payload: map[string]any{
					"scope": "department", "departmentId": deptID,
					"consumed": consumed, "budget": budgetAmount,
				},
			})
			return nil
		}

		if payload.BudgetGroupID != nil {
			groupBudget, _, groupFound, err := tx.Budget().GetGroupBudget(ctx, *payload.BudgetGroupID)
			if err != nil {
				return err
			}
			groupConsumed, _, err := tx.BudgetSnapshots().GetConsumed(ctx, store.SnapshotAxisBudgetGroup, *payload.BudgetGroupID, snapshotPeriod)
			if err != nil {
				return err
			}
			if groupFound && groupConsumed >= groupBudget {
				groupID := *payload.BudgetGroupID
				actions = append(actions, disableAction{
					scope:  "budgetGroup",
					target: groupID,
					keys:   func(ctx context.Context) error { return s.disableBudgetGroupKeys(ctx, groupID) },
					payload: map[string]any{
						"scope": "budgetGroup", "budgetGroupId": groupID,
						"consumed": groupConsumed, "budget": groupBudget,
					},
				})
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Execute disable actions and send notifications outside the transaction
	for _, action := range actions {
		s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, action.target, action.payload)
		if err := action.keys(ctx); err != nil {
			if s.logger != nil {
				s.logger.Error("disable keys failed",
					"scope", action.scope,
					"target", action.target,
					"error", err,
				)
			}
			return err
		}
	}
	return nil
}

func (s *OverrunService) notifyOverrun(ctx context.Context, eventType, recipient string, payload map[string]any) {
	if s.notifier == nil {
		return
	}
	if err := s.notifier.Send(ctx, types.Notification{EventType: eventType, Recipient: recipient, Payload: payload}); err != nil {
		if s.logger != nil {
			s.logger.Error("overrun notification failed",
				"event_type", eventType,
				"recipient", recipient,
				"error", err,
			)
		}
	}
}

func (s *OverrunService) disableMemberKeys(ctx context.Context, memberID string) error {
	keys, err := s.store.Keys().ListActiveMemberKeys(ctx, memberID)
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := s.keyControl.DisablePlatformKey(ctx, key.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *OverrunService) disableDepartmentKeys(ctx context.Context, departmentID string) error {
	mappings, err := s.store.PlatformKeyMappings().ListMappingsByDepartmentID(ctx, departmentID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if err := s.keyControl.DisablePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}

func (s *OverrunService) disableBudgetGroupKeys(ctx context.Context, budgetGroupID string) error {
	mappings, err := s.store.PlatformKeyMappings().ListMappingsByBudgetGroupID(ctx, budgetGroupID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if err := s.keyControl.DisablePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}

var _ OverrunProcessor = (*OverrunService)(nil)
