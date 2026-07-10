package newapisync

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

type ChannelPolicy interface {
	ResolveNewAPIGroup(ctx context.Context, departmentID string) string
}

type LocalChannelPolicy struct{}

func NewLocalChannelPolicy() ChannelPolicy {
	return LocalChannelPolicy{}
}

func (LocalChannelPolicy) ResolveNewAPIGroup(_ context.Context, departmentID string) string {
	return newapi.NewAPIGroupForDepartment(departmentID)
}

type SaaSSharedChannelPolicy struct {
	group string
}

func NewSaaSSharedChannelPolicy(group string) ChannelPolicy {
	return SaaSSharedChannelPolicy{group: group}
}

func (p SaaSSharedChannelPolicy) ResolveNewAPIGroup(_ context.Context, _ string) string {
	return p.group
}

func NewChannelPolicy(cfg config.Config) ChannelPolicy {
	if cfg.SupportSaas {
		return NewSaaSSharedChannelPolicy(cfg.PlatformSharedNewAPIGroup)
	}
	return NewLocalChannelPolicy()
}
