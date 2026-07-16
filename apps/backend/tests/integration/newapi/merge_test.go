package newapi_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestMergeTokenPutPreservesAndHealsExpired(t *testing.T) {
	t.Parallel()
	cur := newapi.Token{
		ID: 9, Name: "tokenjoy:plk-1", Status: 1, RemainQuota: 10,
		ModelLimitsEnabled: true, ModelLimits: "local-test-model",
		Group: "dept-dept-3", ExpiredTime: -1,
	}
	remain := int64(42)
	got := newapi.MergeTokenPut(cur, newapi.UpdateTokenRequest{ID: 9, RemainQuota: &remain})
	if got.ExpiredTime != newapi.TokenExpiredNever || got.Name != cur.Name || got.Group != cur.Group || got.RemainQuota != 42 {
		t.Fatalf("unexpected merge %#v", got)
	}

	broken := cur
	broken.ExpiredTime = 0
	healed := newapi.MergeTokenPut(broken, newapi.UpdateTokenRequest{ID: 9, RemainQuota: &remain})
	if healed.ExpiredTime != newapi.TokenExpiredNever {
		t.Fatalf("expected heal 0→-1, got %d", healed.ExpiredTime)
	}

	explicit := int64(1_700_000_000)
	forced := newapi.MergeTokenPut(cur, newapi.UpdateTokenRequest{ID: 9, ExpiredTime: &explicit})
	if forced.ExpiredTime != explicit {
		t.Fatalf("expected explicit override, got %d", forced.ExpiredTime)
	}
}

func TestMergeChannelPutPreservesUnspecified(t *testing.T) {
	t.Parallel()
	cur := newapi.Channel{ID: 12, Type: 1, Name: "pk-openai", Key: "sk-old", Status: 1, Group: "platform_shared"}
	got := newapi.MergeChannelPut(cur, newapi.UpsertChannelRequest{ID: 12, Status: 2, Key: "sk-new"})
	if got.Name != "pk-openai" || got.Group != "platform_shared" || got.Type != 1 {
		t.Fatalf("expected preserve, got %#v", got)
	}
	if got.Status != 2 || got.Key != "sk-new" {
		t.Fatalf("expected overrides, got %#v", got)
	}
}
