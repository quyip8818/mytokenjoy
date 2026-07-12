package budgetcheck_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/store"
)

type recordingCache struct {
	sets int
}

func (r *recordingCache) Enabled() bool { return true }

func (r *recordingCache) Get(context.Context, int64, string) (budgetcheck.Entry, bool, error) {
	return budgetcheck.Entry{}, false, nil
}

func (r *recordingCache) Set(context.Context, int64, string, budgetcheck.Entry) error {
	r.sets++
	return nil
}

func TestRefreshSummariesSetsWithoutStoreReads(t *testing.T) {
	cache := &recordingCache{}
	budgetcheck.RefreshSummaries(context.Background(), cache, nil, 1, []store.GatewaySoftSummary{
		{
			PlatformKeyID: "pk-1",
			KeyHash:       "hash-1",
			SoftRemain:    12.5,
			UpdatedAt:     time.Unix(1, 0).UTC(),
			Version:       3,
		},
	})
	if cache.sets != 1 {
		t.Fatalf("expected 1 SET, got %d", cache.sets)
	}
}

func TestBlocksWithVersion(t *testing.T) {
	entry := budgetcheck.Entry{SoftRemain: 0, Version: 2}
	if !budgetcheck.BlocksWithVersion(entry, 2) {
		t.Fatal("expected block when versions match and remain <= 0")
	}
	if budgetcheck.BlocksWithVersion(entry, 3) {
		t.Fatal("expected allow when redis version is stale")
	}
	if budgetcheck.BlocksWithVersion(budgetcheck.Entry{SoftRemain: 0, Version: 1}, 0) {
		t.Fatal("expected allow when pg version is unset")
	}
}
