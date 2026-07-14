package newapi

import "testing"

func TestMergeTokenPutPreservesAndHealsExpired(t *testing.T) {
	t.Parallel()
	cur := Token{
		ID: 9, Name: "tokenjoy:plk-1", Status: 1, RemainQuota: 10,
		ModelLimitsEnabled: true, ModelLimits: "local-test-model",
		Group: "dept-dept-3", ExpiredTime: -1,
	}
	remain := int64(42)
	got := mergeTokenPut(cur, UpdateTokenRequest{ID: 9, RemainQuota: &remain})
	if got.ExpiredTime != tokenExpiredNever || got.Name != cur.Name || got.Group != cur.Group || got.RemainQuota != 42 {
		t.Fatalf("unexpected merge %#v", got)
	}

	broken := cur
	broken.ExpiredTime = 0
	healed := mergeTokenPut(broken, UpdateTokenRequest{ID: 9, RemainQuota: &remain})
	if healed.ExpiredTime != tokenExpiredNever {
		t.Fatalf("expected heal 0→-1, got %d", healed.ExpiredTime)
	}

	explicit := int64(1_700_000_000)
	forced := mergeTokenPut(cur, UpdateTokenRequest{ID: 9, ExpiredTime: &explicit})
	if forced.ExpiredTime != explicit {
		t.Fatalf("expected explicit override, got %d", forced.ExpiredTime)
	}
}

func TestMergeChannelPutPreservesUnspecified(t *testing.T) {
	t.Parallel()
	cur := Channel{ID: 12, Type: 1, Name: "pk-openai", Key: "sk-old", Status: 1, Group: "platform_shared"}
	got := mergeChannelPut(cur, UpsertChannelRequest{ID: 12, Status: 2, Key: "sk-new"})
	if got.Name != "pk-openai" || got.Group != "platform_shared" || got.Type != 1 {
		t.Fatalf("expected preserve, got %#v", got)
	}
	if got.Status != 2 || got.Key != "sk-new" {
		t.Fatalf("expected overrides, got %#v", got)
	}
}
