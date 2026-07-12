package budgetcheck

import (
	"context"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
)

type domainStoreAdapter struct {
	store Store
}

// WrapStore adapts infra budgetcheck.Store to domain/budget.GatewaySoftCache.
func WrapStore(store Store) domainbudget.GatewaySoftCache {
	if store == nil {
		return domainbudget.NoopGatewaySoftCache
	}
	return domainStoreAdapter{store: store}
}

func (a domainStoreAdapter) Enabled() bool {
	return a.store.Enabled()
}

func (a domainStoreAdapter) Get(ctx context.Context, companyID int64, keyHash string) (domainbudget.GatewaySoftEntry, bool, error) {
	entry, ok, err := a.store.Get(ctx, companyID, keyHash)
	if err != nil || !ok {
		return domainbudget.GatewaySoftEntry{}, ok, err
	}
	return domainbudget.GatewaySoftEntry{
		SoftRemain: entry.SoftRemain,
		UpdatedAt:  entry.UpdatedAt,
		Version:    entry.Version,
	}, ok, nil
}

func (a domainStoreAdapter) Set(ctx context.Context, companyID int64, keyHash string, entry domainbudget.GatewaySoftEntry) error {
	return a.store.Set(ctx, companyID, keyHash, Entry{
		SoftRemain: entry.SoftRemain,
		UpdatedAt:  entry.UpdatedAt,
		Version:    entry.Version,
	})
}

var _ domainbudget.GatewaySoftCache = domainStoreAdapter{}
