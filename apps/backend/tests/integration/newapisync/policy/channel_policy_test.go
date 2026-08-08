package policy_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/integration/newapisync/policy"
	"github.com/tokenjoy/backend/internal/support/tenant"
)

func TestCompanyChannelPolicy(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000000123")
	ctx := tenant.With(context.Background(), tenant.Info{CompanyID: companyID})
	p := policy.NewCompanyChannelPolicy()
	group := p.ResolveNewAPIGroup(ctx, uuid.New()) // departmentID is ignored
	if group != companyID.String() {
		t.Errorf("expected %q, got %q", companyID.String(), group)
	}
}

func TestNewChannelPolicyAlwaysReturnsCompanyPolicy(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	ctx := tenant.With(context.Background(), tenant.Info{CompanyID: companyID})

	t.Run("saas mode", func(t *testing.T) {
		cfg := config.Config{PlatformConfig: config.PlatformConfig{SupportSaas: true}}
		p := policy.NewChannelPolicy(cfg)
		group := p.ResolveNewAPIGroup(ctx, uuid.New())
		if group != companyID.String() {
			t.Errorf("expected %q, got %q", companyID.String(), group)
		}
	})

	t.Run("local mode", func(t *testing.T) {
		cfg := config.Config{PlatformConfig: config.PlatformConfig{SupportSaas: false}}
		p := policy.NewChannelPolicy(cfg)
		group := p.ResolveNewAPIGroup(ctx, uuid.New())
		if group != companyID.String() {
			t.Errorf("expected %q, got %q", companyID.String(), group)
		}
	})
}

func TestResolveProviderChannelGroup(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000000099")
	ctx := tenant.With(context.Background(), tenant.Info{CompanyID: companyID})
	group := policy.ResolveProviderChannelGroup(ctx)
	if group != companyID.String() {
		t.Errorf("expected %q, got %q", companyID.String(), group)
	}
}
