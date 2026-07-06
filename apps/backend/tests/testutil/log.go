package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/memory"
)

func SeedConsumeLog(t *testing.T, st store.Store, raw store.RawConsumeLog) {
	t.Helper()
	mem, ok := st.(*memory.Store)
	if !ok {
		t.Fatal("SeedConsumeLog requires memory store")
	}
	mem.PutConsumeLog(raw)
}

func DefaultConsumeLog(logID, tokenID int64) store.RawConsumeLog {
	return store.RawConsumeLog{
		ID:        logID,
		TokenID:   tokenID,
		Quota:     500000,
		ModelName: "gpt-4o",
		CreatedAt: 1,
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
	mem, ok := st.(*memory.Store)
	if !ok {
		t.Fatal("AssertIngestFailure requires memory store")
	}
	f, found := mem.IngestFailureByLogID(logID)
	if !found {
		t.Fatalf("expected ingest failure for log %d", logID)
	}
	if wantSource != "" && f.Source != wantSource {
		t.Fatalf("failure source = %q, want %q", f.Source, wantSource)
	}
	return f
}
