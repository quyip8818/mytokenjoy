package policy

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
)

// ChannelPolicy resolves the NewAPI group for platform key tokens and provider channels.
// ponytail: company-level grouping — no department isolation. Model visibility is
// controlled by TokenJoy Gateway precheck routing whitelist, not NewAPI groups.
type ChannelPolicy interface {
	ResolveNewAPIGroup(ctx context.Context, departmentID uuid.UUID) string
}

// CompanyChannelPolicy returns companyID as the group for all tokens and custom channels.
// Platform (tokenjoy) channels use group="" so all tokens can access them.
type CompanyChannelPolicy struct{}

func NewCompanyChannelPolicy() ChannelPolicy {
	return CompanyChannelPolicy{}
}

func (CompanyChannelPolicy) ResolveNewAPIGroup(ctx context.Context, _ uuid.UUID) string {
	return company.CompanyID(ctx).String()
}

func NewChannelPolicy(_ config.Config) ChannelPolicy {
	return NewCompanyChannelPolicy()
}

// ResolveProviderChannelGroup returns the group for provider (custom) channels.
// Uses companyID so all company tokens (same group) can access the custom channels.
func ResolveProviderChannelGroup(ctx context.Context) string {
	return company.CompanyID(ctx).String()
}
