package seed_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestApplyUsageBucketsSeedsMemoryStore(t *testing.T) {
	cfg := testutil.TestConfig()
	st := store.NewMemory(seed.Load(cfg))
	ctx := context.Background()

	if err := seed.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	if count := testutil.UsageBucketCount(st); count != 6 {
		t.Fatalf("expected 6 seeded buckets, got %d", count)
	}
	if err := seed.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	if count := testutil.UsageBucketCount(st); count != 6 {
		t.Fatalf("expected idempotent seed to keep 6 buckets, got %d", count)
	}
}

func TestApplyUsageBucketsProducesNonZeroDashboardSummary(t *testing.T) {
	cfg := testutil.TestConfig()
	st := store.NewMemory(seed.Load(cfg))
	ctx := context.Background()
	if err := seed.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	totals, err := st.Usage().QuerySummary(ctx, types.UsageAggregateQuery{
		Start:    time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Timezone: types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if totals.CostCNY <= 0 {
		t.Fatalf("expected seeded cost > 0, got %v", totals.CostCNY)
	}
}
