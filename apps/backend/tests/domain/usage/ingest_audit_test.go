package usage_test

import (
	"testing"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestIngestVisibleInAuditCalls(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ingest := testutil.NewIngestService(t, cfg, st)
	callLogQuerier := domainusage.NewCallLogQuerier(st.Ledger())
	ctx := testutil.Ctx()

	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	const logID int64 = 9100
	const input = "audit e2e snippet"
	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{
		ID: logID, TokenID: 99, Quota: 500000, ModelName: "gpt-4o", CreatedAt: 1717200000,
		PromptTokens: 100, CompletionTokens: 50, UseTime: 250, Content: input,
	})
	if err := ingest.IngestByLogID(ctx, logID, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	result, err := callLogQuerier.ListCalls(ctx, types.AuditCallsQueryParams{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	key := domainusage.NewAPIIdempotencyKey(logID)
	var found *types.CallLog
	for i := range result.Items {
		if result.Items[i].PreviewSnippet == input {
			found = &result.Items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected audit call with snippet %q, ledger key=%s total=%d", input, key, result.Total)
	}
	if found.InputTokens != 100 || found.OutputTokens != 50 {
		t.Fatalf("unexpected tokens in audit: in=%v out=%v", found.InputTokens, found.OutputTokens)
	}
	if found.LatencyMs != 250 {
		t.Fatalf("unexpected latency %v", found.LatencyMs)
	}
}
