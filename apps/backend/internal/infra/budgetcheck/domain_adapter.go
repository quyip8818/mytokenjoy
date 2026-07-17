package budgetcheck

import (
	"context"

	"github.com/google/uuid"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
)

type domainStoreAdapter struct {
	store Store
}

// WrapStore adapts infra budgetcheck.Store to domain/budget.CombinedKeyCache.
func WrapStore(store Store) domainbudget.CombinedKeyCache {
	if store == nil {
		return domainbudget.NoopCombinedKeyCache
	}
	return domainStoreAdapter{store: store}
}

func (a domainStoreAdapter) Enabled() bool {
	return a.store.Enabled()
}

func (a domainStoreAdapter) Get(ctx context.Context, companyID uuid.UUID, keyHash string) (domainbudget.CombinedKeyEntry, bool, error) {
	entry, ok, err := a.store.Get(ctx, companyID, keyHash)
	if err != nil || !ok {
		return domainbudget.CombinedKeyEntry{}, ok, err
	}
	return domainbudget.CombinedKeyEntry{
		Remain:    entry.Remain,
		UpdatedAt: entry.UpdatedAt,
		Version:   entry.Version,
	}, ok, nil
}

func (a domainStoreAdapter) Set(ctx context.Context, companyID uuid.UUID, keyHash string, entry domainbudget.CombinedKeyEntry) error {
	return a.store.Set(ctx, companyID, keyHash, Entry{
		Remain:    entry.Remain,
		UpdatedAt: entry.UpdatedAt,
		Version:   entry.Version,
	})
}

var _ domainbudget.CombinedKeyCache = domainStoreAdapter{}
