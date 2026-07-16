package usage_test

import (
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestIngestByLogID(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)

	testutil.SeedConsumeLog(t, fix.Store, testutil.DefaultConsumeLog(2002, 99))
	if err := fix.Ingest.IngestByLogID(testutil.Ctx(), 2002, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	ingested, err := testutil.HasLedgerLogID(fix.Store, 2002)
	if err != nil || !ingested {
		t.Fatalf("expected log 2002 in ledger, err=%v ingested=%v", err, ingested)
	}
}

func TestIngestDoesNotWriteUsageBucketDirectly(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t, withBudgetAmount(100_000))
	beforeBuckets := testutil.UsageBucketCount(fix.Store)

	testutil.SeedConsumeLog(t, fix.Store, store.RawConsumeLog{
		ID: 4001, TokenID: 99, Quota: 100000, ModelName: "gpt-4o", CreatedAt: 1717200000,
	})
	if err := fix.Ingest.IngestByLogID(testutil.Ctx(), 4001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	if testutil.UsageBucketCount(fix.Store) != beforeBuckets {
		t.Fatalf("expected ingest to skip usage_buckets, before=%d after=%d", beforeBuckets, testutil.UsageBucketCount(fix.Store))
	}
}

func TestIngestRaw(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)

	raw := testutil.DefaultConsumeLog(3003, 99)
	if err := fix.Ingest.IngestRaw(testutil.Ctx(), raw, types.SourceReconcile); err != nil {
		t.Fatal(err)
	}
	ingested, err := testutil.HasLedgerLogID(fix.Store, 3003)
	if err != nil || !ingested {
		t.Fatalf("expected log 3003 in ledger, err=%v ingested=%v", err, ingested)
	}
}

func TestIngestStoresLedgerPeriodKey(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)
	raw := testutil.DefaultConsumeLog(8801, 99)
	testutil.SeedConsumeLog(t, fix.Store, raw)
	if err := fix.Ingest.IngestByLogID(testutil.Ctx(), 8801, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	want := contract.DemoBudgetPeriod
	entries, _, err := fix.Store.Ledger().ListCallSettledPage(testutil.Ctx(), store.LedgerCallFilter{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	var found *types.UsageLedgerEntry
	for i := range entries {
		if entries[i].IdempotencyKey == usage.NewAPIIdempotencyKey(8801) {
			found = &entries[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected ledger entry for log 8801")
	}
	if found.PeriodKey != want {
		t.Fatalf("PeriodKey = %q, want %q", found.PeriodKey, want)
	}
}

func TestIngestByLogIDNotFound(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t, withoutMapping())

	err := fix.Ingest.IngestByLogID(testutil.Ctx(), 9999, types.SourceWebhook)
	if err == nil {
		t.Fatal("expected error for missing consume log")
	}
	if !errors.Is(err, store.ErrConsumeLogNotFound) {
		t.Fatalf("expected ErrConsumeLogNotFound, got %v", err)
	}
}

func TestIngestByLogIDMappingMissing(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t, withoutMapping())
	testutil.SeedConsumeLog(t, fix.Store, testutil.DefaultConsumeLog(5005, 999999))

	err := fix.Ingest.IngestByLogID(testutil.Ctx(), 5005, types.SourceWebhook)
	if err == nil {
		t.Fatal("expected mapping missing error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusNotFound {
		t.Fatalf("expected not found domain error, got %v", err)
	}
}
