package newapisync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *NewAPISync) EnqueueModelLimitsForDepartment(ctx context.Context, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(UpdateModelLimitsOutboxPayload{
		CompanyID:    company.CompanyID(ctx),
		DepartmentID: departmentID,
	})
	return l.outbox.EnqueueNewAPISyncOutbox(ctx, store.AsyncJob{
		ID:      fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:    store.OutboxKindUpdateModelLimits,
		Payload: payload,
		Status:  store.JobStatusPending,
	})
}

func (l *NewAPISync) EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []string) error {
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

func (l *NewAPISync) SyncModelLimitsForDepartment(ctx context.Context, departmentID string) error {
	if !l.Enabled() {
		return nil
	}
	mappings, err := l.mappings.ListMappingsByDepartmentID(ctx, departmentID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if mapping.SyncStatus != store.MappingSyncStatusSynced || mapping.NewAPIKeyID == nil {
			continue
		}
		if err := l.SyncUpdatePlatformKey(ctx, mapping.PlatformKeyID, nil); err != nil {
			return err
		}
	}
	return nil
}
