package policy

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
)

type ChannelPolicy interface {
	ResolveNewAPIGroup(ctx context.Context, departmentID string) string
}

type LocalChannelPolicy struct{}

func NewLocalChannelPolicy() ChannelPolicy {
	return LocalChannelPolicy{}
}

func (LocalChannelPolicy) ResolveNewAPIGroup(_ context.Context, departmentID string) string {
	return newapiunits.NewAPIGroupForDepartment(departmentID)
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

var groupDisplayNames = map[string]string{
	"dept-3": "后端组",
	"dept-5": "测试组",
}

// GroupDisplayName returns a human-readable NewAPI group label for known demo departments.
func GroupDisplayName(departmentID string) string {
	if name, ok := groupDisplayNames[departmentID]; ok {
		return name
	}
	return departmentID
}

// ResolveProviderChannelGroup picks the NewAPI group for the single provider channel.
func ResolveProviderChannelGroup(cfg config.Config) string {
	if cfg.SupportSaas {
		return cfg.PlatformSharedNewAPIGroup
	}
	deptID := cfg.DefaultProviderDeptID
	if deptID == "" {
		deptID = "dept-3"
	}
	return newapiunits.NewAPIGroupForDepartment(deptID)
}
