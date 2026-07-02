package memory_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestLedgerListCallSettledPageFilters(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.Ctx()

	entry := types.UsageLedgerEntry{
		ID:             "ul-test-1",
		EventType:      types.EventTypeCallSettled,
		IdempotencyKey: "newapi:99001",
		AmountCNY:      1.5,
		DepartmentID:   seed.IDDept3,
		PlatformKeyID:  seed.IDPlatformKey1,
		Source:         types.SourceWebhook,
		OccurredAt:     time.Date(2026, 6, 10, 9, 3, 0, 0, time.UTC),
		Model:          "gpt-4o",
		InputTokens:    10,
		OutputTokens:   5,
		CallDetail: types.UsageCallDetail{
			Caller: "张三", CallerID: seed.IDMember1, CallerType: types.CallerTypeMember,
			Status: types.CallStatusSuccess, PreviewSnippet: "hello ledger",
		},
		CreatedAt: time.Date(2026, 6, 10, 9, 3, 0, 0, time.UTC),
	}
	if _, err := st.Ledger().InsertOnConflict(ctx, entry); err != nil {
		t.Fatal(err)
	}

	filtered, total, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{
		Model: "gpt-4o", Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total < 1 || len(filtered) == 0 {
		t.Fatalf("expected at least one gpt-4o entry, total=%d items=%d", total, len(filtered))
	}

	byKeyword, total, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{
		Keyword: "hello ledger", Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(byKeyword) != 1 || byKeyword[0].ID != entry.ID {
		t.Fatalf("unexpected keyword filter: total=%d items=%+v", total, byKeyword)
	}

	empty, total, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{
		Model: "nonexistent-model", Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(empty) != 0 {
		t.Fatalf("expected no matches, total=%d items=%d", total, len(empty))
	}
}

func TestLedgerQueryMinuteSeries(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.Ctx()

	occurredAt := time.Date(2026, 6, 10, 9, 3, 0, 0, time.UTC)
	if _, err := st.Ledger().InsertOnConflict(ctx, types.UsageLedgerEntry{
		ID: "ul-minute-1", EventType: types.EventTypeCallSettled, IdempotencyKey: "newapi:99002",
		AmountCNY: 2, DepartmentID: seed.IDDept3, PlatformKeyID: seed.IDPlatformKey1,
		Source: types.SourceWebhook, OccurredAt: occurredAt, Model: "gpt-4o-mini",
		InputTokens: 1, OutputTokens: 1, CreatedAt: occurredAt,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Ledger().InsertOnConflict(ctx, types.UsageLedgerEntry{
		ID: "ul-minute-2", EventType: types.EventTypeCallSettled, IdempotencyKey: "newapi:99003",
		AmountCNY: 3, DepartmentID: seed.IDDept3, PlatformKeyID: seed.IDPlatformKey1,
		Source: types.SourceWebhook,
		OccurredAt: occurredAt.Add(2 * time.Minute), Model: "gpt-4o-mini",
		InputTokens: 2, OutputTokens: 2, CreatedAt: occurredAt.Add(2 * time.Minute),
	}); err != nil {
		t.Fatal(err)
	}

	points, err := st.Ledger().QueryMinuteSeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityMinute,
		Start:       time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC),
		End:         time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
		ScopeDeptIDs: []string{seed.IDDept3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) < 2 {
		t.Fatalf("expected at least two minute buckets, got %+v", points)
	}
}
