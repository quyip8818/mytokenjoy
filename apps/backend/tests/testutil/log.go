package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

const defaultConsumeLogUnix = 1718794800 // 2026-06-19T11:00:00Z, within UsageBucketRows query window

func SeedConsumeLog(t *testing.T, st store.Store, raw store.RawConsumeLog) {
	t.Helper()
	if err := postgres.InsertConsumeLog(context.Background(), postgres.LogPool(st), raw); err != nil {
		t.Fatal(err)
	}
}

func DefaultConsumeLog(logID, tokenID int64) store.RawConsumeLog {
	return store.RawConsumeLog{
		ID:        logID,
		TokenID:   tokenID,
		Quota:     500000,
		ModelName: "gpt-4o",
		CreatedAt: defaultConsumeLogUnix,
	}
}

func PendingIngestFailureCount(t *testing.T, st store.Store) int {
	t.Helper()
	n, err := st.Logs().CountPendingIngestFailures(Ctx())
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func AssertIngestFailure(t *testing.T, st store.Store, logID int64, wantSource string) store.IngestFailure {
	t.Helper()
	f, found, err := postgres.GetIngestFailureByLogID(context.Background(), postgres.LogPool(st), logID)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatalf("expected ingest failure for log %d", logID)
	}
	if wantSource != "" && f.Source != wantSource {
		t.Fatalf("failure source = %q, want %q", f.Source, wantSource)
	}
	return f
}
