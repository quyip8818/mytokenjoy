package modellimits

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
	"github.com/tokenjoy/backend/internal/domain/newapisync/platformkey"
	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/store"
)

func EnqueueModelLimitsForDepartment(ctx context.Context, d syncdeps.Deps, departmentID uuid.UUID) error {
	if !syncdeps.Enabled(d) {
		return nil
	}
	return d.Enqueuer.InsertNewAPISync(ctx, ports.SyncJob{
		CompanyID:    company.CompanyID(ctx),
		SubKind:      outbox.KindUpdateModelLimits,
		DepartmentID: departmentID,
	})
}

func EnqueueModelLimitsForDepartments(ctx context.Context, d syncdeps.Deps, departmentIDs []uuid.UUID) error {
	seen := make(map[uuid.UUID]struct{}, len(departmentIDs))
	for _, deptID := range departmentIDs {
		if deptID == uuid.Nil {
			continue
		}
		if _, ok := seen[deptID]; ok {
			continue
		}
		seen[deptID] = struct{}{}
		if err := EnqueueModelLimitsForDepartment(ctx, d, deptID); err != nil {
			return err
		}
	}
	return nil
}

func SyncModelLimitsForDepartment(ctx context.Context, d syncdeps.Deps, departmentID uuid.UUID) error {
	if !syncdeps.Enabled(d) {
		return nil
	}
	mappings, err := d.Mappings.ListMappingsByDepartmentID(ctx, departmentID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if mapping.SyncStatus != store.MappingSyncStatusSynced || mapping.NewAPIKeyID == nil {
			continue
		}
		if err := platformkey.SyncUpdatePlatformKey(ctx, d, mapping.PlatformKeyID, nil); err != nil {
			return err
		}
	}
	return nil
}
