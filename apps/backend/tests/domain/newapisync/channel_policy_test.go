package newapisync_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	domainnewapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
)

func TestChannelPolicyLocal(t *testing.T) {
	t.Parallel()
	policy := domainnewapisync.NewLocalChannelPolicy()
	group := policy.ResolveNewAPIGroup(nil, "dept-123")
	if group == "" {
		t.Error("expected non-empty newapi group")
	}
}

func TestChannelPolicySaaSShared(t *testing.T) {
	t.Parallel()
	policy := domainnewapisync.NewSaaSSharedChannelPolicy("shared-group")
	group := policy.ResolveNewAPIGroup(nil, "dept-123")
	if group != "shared-group" {
		t.Errorf("expected 'shared-group', got %q", group)
	}
}

func TestNewChannelPolicy(t *testing.T) {
	t.Parallel()
	t.Run("saas mode returns shared policy", func(t *testing.T) {
		cfg := config.Config{SupportSaas: true, PlatformSharedNewAPIGroup: "my-shared"}
		policy := domainnewapisync.NewChannelPolicy(cfg)
		group := policy.ResolveNewAPIGroup(nil, "any-dept")
		if group != "my-shared" {
			t.Errorf("expected 'my-shared', got %q", group)
		}
	})

	t.Run("local mode returns local policy", func(t *testing.T) {
		cfg := config.Config{SupportSaas: false}
		policy := domainnewapisync.NewChannelPolicy(cfg)
		group := policy.ResolveNewAPIGroup(nil, "dept-abc")
		if group == "" {
			t.Error("expected non-empty newapi group for local policy")
		}
	})
}
