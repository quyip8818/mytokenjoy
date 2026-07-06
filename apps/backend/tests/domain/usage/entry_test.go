package usage_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestNewAPIIdempotencyKeyAndParse(t *testing.T) {
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

func TestBuildCallSettledEntryPreviewSnippetRespectsRetention(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
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
	memberID := seed.IDMember1
	model := findModel(models, "gpt-4o")
	member := findMember(members, seed.IDMember1)
	key := findPlatformKey(keys, seed.IDPlatformKey1)
	entry, err := domainusage.BuildCallSettledEntry(domainusage.EntryBuildInput{
		Raw: store.RawConsumeLog{
			ID: 9001, TokenID: 99, Quota: 100, ModelName: "gpt-4o", CreatedAt: 1,
			Content: "should not be stored",
		},
		Mapping: &store.RelayMapping{
			PlatformKeyID: seed.IDPlatformKey1,
			MemberID:      &memberID,
			DepartmentID:  seed.IDDept3,
		},
		Source:      types.SourceWebhook,
		Model:       model,
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
			PlatformKeyID: seed.IDPlatformKey1,
			MemberID:      &memberID,
			DepartmentID:  seed.IDDept3,
		},
		Source:      types.SourceWebhook,
		Model:       model,
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

func findModel(models []types.ModelInfo, name string) *types.ModelInfo {
	for i := range models {
		if models[i].Name == name {
			return &models[i]
		}
	}
	return nil
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
