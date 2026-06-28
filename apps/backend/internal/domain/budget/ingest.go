package budget

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/notification"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
	"github.com/tokenjoy/backend/internal/pkg/memberquota"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestService struct {
	cfg       config.Config
	store     store.Store
	lifecycle relay.Lifecycle
	notifier  notification.Notifier
	logger    *slog.Logger
}

func NewIngestService(
	cfg config.Config,
	st store.Store,
	lifecycle relay.Lifecycle,
	notifier notification.Notifier,
	logger *slog.Logger,
) *IngestService {
	return &IngestService{cfg: cfg, store: st, lifecycle: lifecycle, notifier: notifier, logger: logger}
}

func (s *IngestService) Ingest(ctx context.Context, payload newapi.WebhookLogPayload) error {
	exists, err := s.store.Relay().HasIngestedLogID(payload.ID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	mapping, err := s.store.Relay().GetMappingByNewAPITokenID(payload.TokenID)
	if err != nil {
		return err
	}
	if mapping == nil {
		s.logger.Warn("ingest rejected: mapping missing", "token_id", payload.TokenID, "log_id", payload.ID)
		return fmt.Errorf("mapping not found for token %d", payload.TokenID)
	}

	models := s.store.Models().Models()
	modelName := domainusage.ResolveWebhookModel(payload)
	costCNY := domainusage.CostCNYFromLog(payload.Quota, modelName, models)

	var memberID *string
	if mapping.MemberID != nil {
		memberID = mapping.MemberID
	}

	if err := s.store.WithTx(ctx, func(st store.Store) error {
		platformKeys := st.Keys().PlatformKeys()
		keyIdx := -1
		for i := range platformKeys {
			if platformKeys[i].ID == mapping.PlatformKeyID {
				keyIdx = i
				break
			}
		}
		if keyIdx < 0 {
			return fmt.Errorf("platform key not found: %s", mapping.PlatformKeyID)
		}
		platformKeys[keyIdx].Used += costCNY
		if err := st.Keys().SetPlatformKeys(platformKeys); err != nil {
			return err
		}

		tree := st.Budget().Tree()
		members := st.Org().Members()
		if memberID != nil {
			if member, ok := queryutil.FindMemberByID(members, *memberID); ok {
				if err := rollupDepartmentConsumed(tree, member.DepartmentID, costCNY); err != nil {
					return err
				}
				if err := st.Budget().SetTree(tree); err != nil {
					return err
				}
			}
		}

		if mapping.BudgetGroupID != nil {
			groups := st.Budget().Groups()
			for i := range groups {
				if groups[i].ID == *mapping.BudgetGroupID {
					groups[i].Consumed += costCNY
					if err := st.Budget().SetGroups(groups); err != nil {
						return err
					}
					if err := st.Relay().EnqueueRebalance(store.RebalanceAxisBudgetGroup, groups[i].ID); err != nil {
						return err
					}
					break
				}
			}
		}

		if memberID != nil {
			if err := st.Relay().EnqueueRebalance(store.RebalanceAxisMember, *memberID); err != nil {
				return err
			}
		}
		if err := st.Relay().EnqueueRebalance(store.RebalanceAxisDepartment, mapping.DepartmentID); err != nil {
			return err
		}

		tree = st.Budget().Tree()
		members = st.Org().Members()
		if err := s.evaluateOverrun(ctx, st, tree, members, memberID, mapping); err != nil {
			return err
		}

		memberIDValue := ""
		if memberID != nil {
			memberIDValue = *memberID
		}
		if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
			BucketStart:  usageHourFromPayload(payload),
			DepartmentID: mapping.DepartmentID,
			MemberID:     memberIDValue,
			Model:        modelName,
			CostCNY:      costCNY,
			CallCount:    1,
		}); err != nil {
			return err
		}

		if err := st.Relay().InsertIngestedLogID(payload.ID); err != nil {
			return err
		}
		if payload.ID > 0 {
			last, err := st.Relay().GetLastLogID()
			if err != nil {
				return err
			}
			if payload.ID > last {
				if err := st.Relay().SetLastLogID(payload.ID); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *IngestService) IngestFromOutbox(ctx context.Context, raw json.RawMessage) error {
	var payload newapi.WebhookLogPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	return s.Ingest(ctx, payload)
}

func rollupDepartmentConsumed(tree []types.BudgetNode, leafDepartmentID string, costCNY float64) error {
	node := budgetutil.FindBudgetNode(tree, leafDepartmentID)
	if node == nil {
		return fmt.Errorf("department node not found: %s", leafDepartmentID)
	}
	node.Consumed += costCNY
	ancestors := collectAncestorIDs(tree, leafDepartmentID)
	for _, ancestorID := range ancestors {
		ancestor := budgetutil.FindBudgetNode(tree, ancestorID)
		if ancestor != nil {
			ancestor.Consumed += costCNY
		}
	}
	return nil
}

func collectAncestorIDs(tree []types.BudgetNode, leafID string) []string {
	var ancestors []string
	var walk func(nodes []types.BudgetNode, path []string) bool
	walk = func(nodes []types.BudgetNode, path []string) bool {
		for _, node := range nodes {
			nextPath := append(path, node.ID)
			if node.ID == leafID {
				ancestors = append([]string{}, path...)
				return true
			}
			if len(node.Children) > 0 && walk(node.Children, nextPath) {
				return true
			}
		}
		return false
	}
	walk(tree, nil)
	return ancestors
}

func (s *IngestService) evaluateOverrun(
	ctx context.Context,
	st store.Store,
	tree []types.BudgetNode,
	members []types.Member,
	memberID *string,
	mapping *store.RelayMapping,
) error {
	if s.lifecycle == nil || !s.lifecycle.Enabled() {
		return nil
	}
	platformKeys := st.Keys().PlatformKeys()
	pools := st.Budget().MemberQuotaPools()

	if memberID != nil && mapping.BudgetGroupID == nil {
		used := memberquota.GetUsedKeyQuota(platformKeys, *memberID)
		capacity := memberquota.GetPersonalQuota(pools, *memberID)
		if used >= capacity {
			s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, *memberID, map[string]any{
				"scope": "member", "memberId": *memberID, "used": used, "capacity": capacity,
			})
			return s.disableMemberKeys(ctx, platformKeys, *memberID)
		}
	}

	if node := budgetutil.FindBudgetNode(tree, mapping.DepartmentID); node != nil {
		if node.Consumed >= node.Budget {
			s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, mapping.DepartmentID, map[string]any{
				"scope": "department", "departmentId": mapping.DepartmentID,
				"consumed": node.Consumed, "budget": node.Budget,
			})
			return s.disableDepartmentKeys(ctx, mapping.DepartmentID)
		}
	}

	if mapping.BudgetGroupID != nil {
		groups := st.Budget().Groups()
		for _, group := range groups {
			if group.ID == *mapping.BudgetGroupID && group.Consumed >= group.Budget {
				s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, group.ID, map[string]any{
					"scope": "budgetGroup", "budgetGroupId": group.ID,
					"consumed": group.Consumed, "budget": group.Budget,
				})
				return s.disableBudgetGroupKeys(ctx, *mapping.BudgetGroupID)
			}
		}
	}
	return nil
}

func (s *IngestService) notifyOverrun(ctx context.Context, eventType, recipient string, payload map[string]any) {
	if s.notifier == nil {
		return
	}
	_ = s.notifier.Send(ctx, types.Notification{EventType: eventType, Recipient: recipient, Payload: payload})
}

func usageHourFromPayload(payload newapi.WebhookLogPayload) time.Time {
	ts := payload.CreatedAt
	if ts <= 0 {
		ts = time.Now().Unix()
	}
	return time.Unix(ts, 0).UTC().Truncate(time.Hour)
}

func (s *IngestService) disableMemberKeys(ctx context.Context, platformKeys []types.PlatformKey, memberID string) error {
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil && key.Status == "active" {
			if err := s.lifecycle.DisablePlatformKey(ctx, key.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *IngestService) disableDepartmentKeys(ctx context.Context, departmentID string) error {
	mappings, err := s.store.Relay().ListMappingsByDepartmentID(departmentID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if err := s.lifecycle.DisablePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}

func (s *IngestService) disableBudgetGroupKeys(ctx context.Context, budgetGroupID string) error {
	mappings, err := s.store.Relay().ListMappingsByBudgetGroupID(budgetGroupID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if err := s.lifecycle.DisablePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}

func (s *IngestService) EnqueueFailed(payload newapi.WebhookLogPayload, ingestErr error) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.store.Relay().EnqueueWebhookOutbox(store.WebhookOutboxEntry{
		ID:        fmt.Sprintf("wh-%d", time.Now().UnixNano()),
		Payload:   raw,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}
