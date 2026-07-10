package usage_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestNewAPIIdempotencyKeyAndParse(t *testing.T) {
	t.Parallel()
	key := domainusage.NewAPIIdempotencyKey(42)
	if key != "newapi:42" {
		t.Fatalf("unexpected key %q", key)
	}
	logID, ok := domainusage.ParseNewAPILogID(key)
	if !ok || logID != 42 {
		t.Fatalf("parse failed: ok=%v id=%d", ok, logID)
	}
}

func TestTruncatePreview(t *testing.T) {
	t.Parallel()
	short := "hello"
	if domainusage.TruncatePreview(short) != short {
		t.Fatal("short preview should be unchanged")
	}
	long := strings.Repeat("字", types.PreviewSnippetMaxLen+10)
	truncated := domainusage.TruncatePreview(long)
	if len([]rune(truncated)) != types.PreviewSnippetMaxLen {
		t.Fatalf("expected %d runes, got %d", types.PreviewSnippetMaxLen, len([]rune(truncated)))
	}
}

func TestOccurredAtFromPayloadRejectsMissing(t *testing.T) {
	t.Parallel()
	if _, err := domainusage.OccurredAtFromPayload(0); err == nil {
		t.Fatal("expected error for createdAt=0")
	}
	if _, err := domainusage.OccurredAtFromPayload(-1); err == nil {
		t.Fatal("expected error for createdAt<0")
	}
	got, err := domainusage.OccurredAtFromPayload(1717200000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Unix() != 1717200000 {
		t.Fatalf("OccurredAt = %v", got)
	}
}

func TestBuildCallSettledEntryRejectsMissingOccurredAt(t *testing.T) {
	t.Parallel()
	_, err := domainusage.BuildCallSettledEntry(domainusage.EntryBuildInput{
		Raw: store.RawConsumeLog{
			ID: 9000, TokenID: 99, Quota: 100, ModelName: "gpt-4o", CreatedAt: 0,
		},
		Mapping: &store.RelayMapping{
			PlatformKeyID: contract.IDPlatformKey1,
			DepartmentID:  contract.IDDept3,
		},
		Source: types.SourceWebhook,
	})
	if err == nil {
		t.Fatal("expected error when CreatedAt is missing")
	}
}

func TestBuildCallSettledEntryPreviewSnippetRespectsRetention(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()

	if err := st.Audit().SetSettings(ctx, types.AuditSettings{ContentRetentionEnabled: false}); err != nil {
		t.Fatal(err)
	}
	models, err := st.Models().Models(ctx)
	if err != nil {
		t.Fatal(err)
	}
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	memberID := contract.IDMember1
	member := findMember(members, contract.IDMember1)
	key := findPlatformKey(keys, contract.IDPlatformKey1)
	entry, err := domainusage.BuildCallSettledEntry(domainusage.EntryBuildInput{
		Raw: store.RawConsumeLog{
			ID: 9001, TokenID: 99, Quota: 100, ModelName: "gpt-4o", CreatedAt: 1,
			Content: "should not be stored",
		},
		Mapping: &store.RelayMapping{
			PlatformKeyID: contract.IDPlatformKey1,
			MemberID:      &memberID,
			DepartmentID:  contract.IDDept3,
		},
		Source:      types.SourceWebhook,
		Catalog:     models,
		Settings:    types.AuditSettings{ContentRetentionEnabled: false},
		Member:      member,
		PlatformKey: key,
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.CallDetail.PreviewSnippet != "" {
		t.Fatalf("expected empty snippet when retention disabled, got %q", entry.CallDetail.PreviewSnippet)
	}

	entry, err = domainusage.BuildCallSettledEntry(domainusage.EntryBuildInput{
		Raw: store.RawConsumeLog{
			ID: 9002, TokenID: 99, Quota: 100, ModelName: "gpt-4o", CreatedAt: 1,
			Content: "preview text", PromptTokens: 10, CompletionTokens: 5, UseTime: 120,
		},
		Mapping: &store.RelayMapping{
			PlatformKeyID: contract.IDPlatformKey1,
			MemberID:      &memberID,
			DepartmentID:  contract.IDDept3,
		},
		Source:      types.SourceWebhook,
		Catalog:     models,
		Settings:    types.AuditSettings{ContentRetentionEnabled: true},
		Member:      member,
		PlatformKey: key,
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.CallDetail.PreviewSnippet != "preview text" {
		t.Fatalf("expected snippet, got %q", entry.CallDetail.PreviewSnippet)
	}
	if entry.InputTokens != 10 || entry.OutputTokens != 5 {
		t.Fatalf("unexpected tokens in=%d out=%d", entry.InputTokens, entry.OutputTokens)
	}
	if entry.CallDetail.LatencyMs != 120 {
		t.Fatalf("unexpected latency %v", entry.CallDetail.LatencyMs)
	}
}

func findMember(members []types.Member, id string) *types.Member {
	for i := range members {
		if members[i].ID == id {
			return &members[i]
		}
	}
	return nil
}

func findPlatformKey(keys []types.PlatformKey, id string) *types.PlatformKey {
	for i := range keys {
		if keys[i].ID == id {
			return &keys[i]
		}
	}
	return nil
}
