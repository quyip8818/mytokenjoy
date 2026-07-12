package budget

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type overrunPayload struct {
	DepartmentID  string  `json:"departmentId"`
	MemberID      *string `json:"memberId,omitempty"`
	ProjectID     *string `json:"projectId,omitempty"`
	PlatformKeyID string  `json:"platformKeyId"`
}

type OverrunService struct {
	cfg        config.Config
	store      store.Store
	keyControl newapisync.OverrunKeyControl
	notifier   types.Notifier
	logger     *slog.Logger
}

func NewOverrunService(
	cfg config.Config,
	st store.Store,
	keyControl newapisync.OverrunKeyControl,
	notifier types.Notifier,
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

	var action *disableAction

	if err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}

		open, err := pkgbudget.OpenDepartmentPeriod(ctx, tx.Org().Nodes(), payload.DepartmentID, s.cfg.Clock())
		if err != nil {
			return err
		}
		periodKey := open.String()
		consumedRepo := tx.BudgetConsumed()

		if payload.PlatformKeyID != "" {
			key, err := tx.Keys().PlatformKeyByID(ctx, payload.PlatformKeyID)
			if err != nil {
				return err
			}
			if key != nil && key.Budget > 0 {
				keyConsumed, _, err := consumedRepo.GetConsumed(ctx, store.AxisKindPlatformKey, payload.PlatformKeyID, periodKey)
				if err != nil {
					return err
				}
				if pkgbudget.BudgetExhausted(keyConsumed, key.Budget) {
					keyID := payload.PlatformKeyID
					action = &disableAction{
						scope:  "platformKey",
						target: keyID,
						keys:   func(ctx context.Context) error { return s.keyControl.DisablePlatformKey(ctx, keyID) },
						payload: map[string]any{
							"scope": "platformKey", "platformKeyId": keyID,
							"consumed": keyConsumed, "budget": key.Budget,
						},
					}
					return nil
				}
			}
		}

		if payload.MemberID != nil && payload.ProjectID == nil {
			memberConsumed, _, err := consumedRepo.GetConsumed(ctx, store.AxisKindMember, *payload.MemberID, periodKey)
			if err != nil {
				return err
			}
			capacity, found, err := tx.Org().MemberPersonalBudget(ctx, *payload.MemberID)
			if err != nil {
				return err
			}
			if found && pkgbudget.BudgetExhausted(memberConsumed, capacity) {
				memberID := *payload.MemberID
				action = &disableAction{
					scope:  "member",
					target: memberID,
					keys:   func(ctx context.Context) error { return s.disableMemberKeys(ctx, memberID) },
					payload: map[string]any{
						"scope": "member", "memberId": memberID, "consumed": memberConsumed, "capacity": capacity,
					},
				}
				return nil
			}
		}

		deptBudget, deptFound, err := tx.Org().Nodes().GetNodeBudget(ctx, payload.DepartmentID)
		if err != nil {
			return err
		}
		deptConsumed, _, err := consumedRepo.GetConsumed(ctx, store.AxisKindOrgNode, payload.DepartmentID, periodKey)
		if err != nil {
			return err
		}
		if deptFound && pkgbudget.BudgetExhausted(deptConsumed, deptBudget) {
			deptID := payload.DepartmentID
			action = &disableAction{
				scope:  "org_node",
				target: deptID,
				keys:   func(ctx context.Context) error { return s.disableDepartmentKeys(ctx, deptID) },
				payload: map[string]any{
					"axisKind": "org_node", "departmentId": deptID,
					"consumed": deptConsumed, "budget": deptBudget,
				},
			}
			return nil
		}

		if payload.ProjectID != nil {
			projectBudget, _, projectFound, err := tx.Budget().GetProjectBudget(ctx, *payload.ProjectID)
			if err != nil {
				return err
			}
			projectConsumed, _, err := consumedRepo.GetConsumed(ctx, store.AxisKindProject, *payload.ProjectID, periodKey)
			if err != nil {
				return err
			}
			if projectFound && pkgbudget.BudgetExhausted(projectConsumed, projectBudget) {
				projectID := *payload.ProjectID
				action = &disableAction{
					scope:  "project",
					target: projectID,
					keys:   func(ctx context.Context) error { return s.disableProjectKeys(ctx, projectID) },
					payload: map[string]any{
						"axisKind": "project", "projectId": projectID,
						"consumed": projectConsumed, "budget": projectBudget,
					},
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if action == nil {
		return nil
	}
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

func (s *OverrunService) disableProjectKeys(ctx context.Context, projectID string) error {
	mappings, err := s.store.PlatformKeyMappings().ListMappingsByProjectID(ctx, projectID)
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
