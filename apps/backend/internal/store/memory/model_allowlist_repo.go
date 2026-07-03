package memory

import (
	"context"
	"sort"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryModelAllowlistRepo struct{ store *Store }

func (r *memoryModelAllowlistRepo) List(ctx context.Context, ownerType, ownerID string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	companyID := store.CompanyID(ctx)
	models := make([]string, 0)
	for _, row := range r.store.companySnapshot(companyID).ModelAllowlist {
		if row.OwnerType == ownerType && row.OwnerID == ownerID {
			models = append(models, row.ModelName)
		}
	}
	sort.Strings(models)
	return models, nil
}

func (r *memoryModelAllowlistRepo) Replace(ctx context.Context, ownerType, ownerID string, models []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	filtered := make([]store.ModelAllowlistRow, 0, len(snap.ModelAllowlist))
	for _, row := range snap.ModelAllowlist {
		if row.OwnerType == ownerType && row.OwnerID == ownerID {
			continue
		}
		filtered = append(filtered, row)
	}
	for _, modelName := range models {
		filtered = append(filtered, store.ModelAllowlistRow{
			OwnerType: ownerType,
			OwnerID:   ownerID,
			ModelName: modelName,
		})
	}
	snap.ModelAllowlist = filtered
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryModelAllowlistRepo) DeleteByOwner(ctx context.Context, ownerType, ownerID string) error {
	return r.Replace(ctx, ownerType, ownerID, nil)
}
