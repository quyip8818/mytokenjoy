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
	st := s.store

	if payload.MemberID != nil && payload.BudgetGroupID == nil {
		used, err := st.Keys().SumMemberKeyUsed(ctx, *payload.MemberID, s.cfg.Clock())
		if err != nil {
			return err
		}
		capacity, found, err := st.Org().MemberPersonalBudget(ctx, *payload.MemberID)
		if err != nil {
			return err
		}
		if found && used >= capacity {
			s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, *payload.MemberID, map[string]any{
				"scope": "member", "memberId": *payload.MemberID, "used": used, "capacity": capacity,
			})
			return s.disableMemberKeys(ctx, *payload.MemberID)
		}
	}

	open, err := pkgbudget.OpenDepartmentPeriod(ctx, st.Org().Nodes(), payload.DepartmentID, s.cfg.Clock())
	if err != nil {
		return err
	}
	snapshotPeriod := open.String()

	budgetAmount, found, err := st.Org().Nodes().GetNodeBudget(ctx, payload.DepartmentID)
	if err != nil {
		return err
	}
	consumed, _, err := st.BudgetSnapshots().GetConsumed(ctx, store.SnapshotAxisOrgNode, payload.DepartmentID, snapshotPeriod)
	if err != nil {
		return err
	}
	if found && consumed >= budgetAmount {
		s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, payload.DepartmentID, map[string]any{
			"scope": "department", "departmentId": payload.DepartmentID,
			"consumed": consumed, "budget": budgetAmount,
		})
		return s.disableDepartmentKeys(ctx, payload.DepartmentID)
	}

	if payload.BudgetGroupID != nil {
		groupBudget, _, groupFound, err := st.Budget().GetGroupBudget(ctx, *payload.BudgetGroupID)
		if err != nil {
			return err
		}
		groupConsumed, _, err := st.BudgetSnapshots().GetConsumed(ctx, store.SnapshotAxisBudgetGroup, *payload.BudgetGroupID, snapshotPeriod)
		if err != nil {
			return err
		}
		if groupFound && groupConsumed >= groupBudget {
			s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, *payload.BudgetGroupID, map[string]any{
				"scope": "budgetGroup", "budgetGroupId": *payload.BudgetGroupID,
				"consumed": groupConsumed, "budget": groupBudget,
			})
			return s.disableBudgetGroupKeys(ctx, *payload.BudgetGroupID)
		}
	}
	return nil
}

func (s *OverrunService) notifyOverrun(ctx context.Context, eventType, recipient string, payload map[string]any) {
	if s.notifier == nil {
		return
	}
	_ = s.notifier.Send(ctx, types.Notification{EventType: eventType, Recipient: recipient, Payload: payload})
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
