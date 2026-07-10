//go:build testhook

package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

const defaultConsumeLogUnix = 1781866800 // 2026-06-19T11:00:00Z, within UsageBucketRows query window

func SeedConsumeLog(t *testing.T, st store.Store, raw store.RawConsumeLog) {
	t.Helper()
	if err := postgres.InsertConsumeLog(context.Background(), st, raw); err != nil {
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

func PendingIngestJobCount(t *testing.T, st store.Store) int {
	t.Helper()
	n, err := st.Logs().CountPendingIngestJobs(Ctx())
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func AssertIngestJob(t *testing.T, st store.Store, logID int64, wantSource string) store.IngestJob {
	t.Helper()
	job, found, err := postgres.GetIngestJobByLogID(context.Background(), st, logID)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatalf("expected ingest job for log %d", logID)
	}
	if wantSource != "" && job.Source != wantSource {
		t.Fatalf("job source = %q, want %q", job.Source, wantSource)
	}
	return job
}
