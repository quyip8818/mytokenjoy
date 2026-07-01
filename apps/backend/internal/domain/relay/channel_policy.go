package relay

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

type ChannelPolicy interface {
	ResolveRelayGroup(ctx context.Context, departmentID string) string
}

type LocalChannelPolicy struct{}

func NewLocalChannelPolicy() ChannelPolicy {
	return LocalChannelPolicy{}
}

func (LocalChannelPolicy) ResolveRelayGroup(_ context.Context, departmentID string) string {
	return newapi.RelayGroupForDepartment(departmentID)
}

type SaaSSharedChannelPolicy struct {
	group string
}

func NewSaaSSharedChannelPolicy(group string) ChannelPolicy {
	return SaaSSharedChannelPolicy{group: group}
}

func (p SaaSSharedChannelPolicy) ResolveRelayGroup(_ context.Context, _ string) string {
	return p.group
}

func NewChannelPolicy(cfg config.Config) ChannelPolicy {
	if cfg.MultiCompany {
		return NewSaaSSharedChannelPolicy(cfg.PlatformSharedRelayGroup)
	}
	return NewLocalChannelPolicy()
}
