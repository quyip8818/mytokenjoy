package policy_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
)

func TestChannelPolicyLocal(t *testing.T) {
	t.Parallel()
	p := policy.NewLocalChannelPolicy()
	group := p.ResolveNewAPIGroup(nil, "dept-123")
	if group == "" {
		t.Error("expected non-empty newapi group")
	}
}

func TestChannelPolicySaaSShared(t *testing.T) {
	t.Parallel()
	p := policy.NewSaaSSharedChannelPolicy("shared-group")
	group := p.ResolveNewAPIGroup(nil, "dept-123")
	if group != "shared-group" {
		t.Errorf("expected 'shared-group', got %q", group)
	}
}

func TestNewChannelPolicy(t *testing.T) {
	t.Parallel()
	t.Run("saas mode returns shared policy", func(t *testing.T) {
		cfg := config.Config{PlatformConfig: config.PlatformConfig{SupportSaas: true, PlatformSharedNewAPIGroup: "my-shared"}}
		p := policy.NewChannelPolicy(cfg)
		group := p.ResolveNewAPIGroup(nil, "any-dept")
		if group != "my-shared" {
			t.Errorf("expected 'my-shared', got %q", group)
		}
	})

	t.Run("local mode returns local policy", func(t *testing.T) {
		cfg := config.Config{PlatformConfig: config.PlatformConfig{SupportSaas: false}}
		p := policy.NewChannelPolicy(cfg)
		group := p.ResolveNewAPIGroup(nil, "dept-abc")
		if group == "" {
			t.Error("expected non-empty newapi group for local policy")
		}
	})
}
