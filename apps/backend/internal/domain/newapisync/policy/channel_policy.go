package policy

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
)

type ChannelPolicy interface {
	ResolveNewAPIGroup(ctx context.Context, departmentID uuid.UUID) string
}

type LocalChannelPolicy struct{}

func NewLocalChannelPolicy() ChannelPolicy {
	return LocalChannelPolicy{}
}

func (LocalChannelPolicy) ResolveNewAPIGroup(_ context.Context, departmentID uuid.UUID) string {
	return newapiunits.NewAPIGroupForDepartment(departmentID)
}

type SaaSSharedChannelPolicy struct {
	group string
}

func NewSaaSSharedChannelPolicy(group string) ChannelPolicy {
	return SaaSSharedChannelPolicy{group: group}
}

func (p SaaSSharedChannelPolicy) ResolveNewAPIGroup(_ context.Context, _ uuid.UUID) string {
	return p.group
}

func NewChannelPolicy(cfg config.Config) ChannelPolicy {
	if cfg.SupportSaas {
		return NewSaaSSharedChannelPolicy(cfg.PlatformSharedNewAPIGroup)
	}
	return NewLocalChannelPolicy()
}

var groupDisplayNames = map[uuid.UUID]string{}

// GroupDisplayName returns a human-readable NewAPI group label for known demo departments.
func GroupDisplayName(departmentID uuid.UUID) string {
	if name, ok := groupDisplayNames[departmentID]; ok {
		return name
	}
	return departmentID.String()
}

// ResolveProviderChannelGroup picks the NewAPI group for the provider channel.
// In local (non-SaaS) mode, returns empty so all department tokens can access the shared channel.
// In SaaS mode, returns the platform-wide shared group.
func ResolveProviderChannelGroup(cfg config.Config) string {
	if cfg.SupportSaas {
		return cfg.PlatformSharedNewAPIGroup
	}
	return ""
}
