package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *TokenLifecycle) EnqueueModelLimitsForDepartment(ctx context.Context, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpdateModelLimitsOutboxPayload{
		CompanyID:    company.CompanyID(ctx),
		DepartmentID: departmentID,
	})
	return l.relayOutbox.EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindUpdateModelLimits,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}

func (l *TokenLifecycle) EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []string) error {
	seen := make(map[string]struct{}, len(departmentIDs))
	for _, deptID := range departmentIDs {
		if deptID == "" {
			continue
		}
		if _, ok := seen[deptID]; ok {
			continue
		}
		seen[deptID] = struct{}{}
		if err := l.EnqueueModelLimitsForDepartment(ctx, deptID); err != nil {
			return err
		}
	}
	return nil
}

func (l *TokenLifecycle) SyncModelLimitsForDepartment(ctx context.Context, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	mappings, err := l.mappings.ListMappingsByDepartmentID(ctx, departmentID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if mapping.SyncStatus != store.RelaySyncStatusSynced || mapping.NewAPITokenID == nil {
			continue
		}
		if err := l.SyncUpdatePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}
