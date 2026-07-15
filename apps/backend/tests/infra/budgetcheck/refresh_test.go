package budgetcheck_test

import (
	"context"
	"testing"
	"time"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type recordingCache struct {
	sets int
}

func (r *recordingCache) Enabled() bool { return true }

func (r *recordingCache) Get(context.Context, int64, string) (domainbudget.CombinedKeyEntry, bool, error) {
	return domainbudget.CombinedKeyEntry{}, false, nil
}

func (r *recordingCache) Set(context.Context, int64, string, domainbudget.CombinedKeyEntry) error {
	r.sets++
	return nil
}

func TestRefreshSummariesSetsWithoutStoreReads(t *testing.T) {
	cache := &recordingCache{}
	domainbudget.RefreshCombinedKeySummaries(context.Background(), cache, nil, 1, []store.CombinedKeySummary{
		{
			PlatformKeyID: "pk-1",
			KeyHash:       "hash-1",
			Remain:        12.5,
			UpdatedAt:     time.Unix(1, 0).UTC(),
			Version:       3,
		},
	})
	if cache.sets != 1 {
		t.Fatalf("expected 1 SET, got %d", cache.sets)
	}
}

func TestBlocksCombinedKey(t *testing.T) {
	entry := domainbudget.CombinedKeyEntry{Remain: 0, Version: 2}
	if !domainbudget.BlocksCombinedKey(entry, 2) {
		t.Fatal("expected block when versions match and remain <= 0")
	}
	if domainbudget.BlocksCombinedKey(entry, 3) {
		t.Fatal("expected allow when redis version is stale")
	}
	if domainbudget.BlocksCombinedKey(domainbudget.CombinedKeyEntry{Remain: 0, Version: 1}, 0) {
		t.Fatal("expected allow when pg version is unset")
	}
}
