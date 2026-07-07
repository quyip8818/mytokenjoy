package relay_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestComputeRemainQuotaCNY(t *testing.T) {
	memberID := "m1"
	groupID := "g1"

	tests := []struct {
		name         string
		key          types.PlatformKey
		tree         []types.BudgetNode
		members      []types.Member
		platformKeys []types.PlatformKey
		groups       []types.BudgetGroup
		departmentID string
		want         float64
	}{
		{
			name: "key remaining only (no group or member)",
			key: types.PlatformKey{
				ID:    "k1",
				Quota: 100,
				Used:  30,
			},
			tree:         nil,
			members:      nil,
			platformKeys: nil,
			groups:       nil,
			departmentID: "d-unknown",
			want:         70,
		},
		{
			name: "key remaining limited by budget group",
			key: types.PlatformKey{
				ID:            "k1",
				Quota:         200,
				Used:          50,
				BudgetGroupID: &groupID,
			},
			tree: nil,
			groups: []types.BudgetGroup{
				{ID: "g1", Budget: 80, Consumed: 60},
			},
			departmentID: "d-unknown",
			want:         20, // min(150, 20) = 20
		},
		{
			name: "key remaining limited by department budget",
			key: types.PlatformKey{
				ID:    "k1",
				Quota: 1000,
				Used:  0,
			},
			tree: []types.BudgetNode{
				{ID: "d1", Budget: 500, Consumed: 450},
			},
			departmentID: "d1",
			want:         50, // min(1000, 50) = 50
		},
		{
			name: "department reserved pool reduces available",
			key: types.PlatformKey{
				ID:    "k1",
				Quota: 1000,
				Used:  0,
			},
			tree: []types.BudgetNode{
				{ID: "d1", Budget: 500, Consumed: 400, ReservedPool: floatPtr(30)},
			},
			departmentID: "d1",
			want:         70, // 500-400-30 = 70; min(1000, 70) = 70
		},
		{
			name: "negative remaining clamped to zero",
			key: types.PlatformKey{
				ID:    "k1",
				Quota: 10,
				Used:  20,
			},
			tree:         nil,
			departmentID: "d-unknown",
			want:         0,
		},
		{
			name: "member quota limits result",
			key: types.PlatformKey{
				ID:       "k1",
				Quota:    500,
				Used:     0,
				MemberID: &memberID,
			},
			tree: nil,
			members: []types.Member{
				{ID: "m1", PersonalQuota: 100},
			},
			platformKeys: []types.PlatformKey{
				{ID: "k1", MemberID: &memberID, Status: "active", Used: 80},
			},
			departmentID: "d-unknown",
			want:         20, // memberCap=100, memberUsed=80, memberRemaining=20; min(500, 20) = 20
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domainrelay.ComputeRemainQuotaCNY(
				tt.key, tt.tree, tt.members, tt.platformKeys, tt.groups, tt.departmentID,
			)
			if got != tt.want {
				t.Errorf("ComputeRemainQuotaCNY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannelPolicyLocal(t *testing.T) {
	policy := domainrelay.NewLocalChannelPolicy()
	group := policy.ResolveRelayGroup(nil, "dept-123")
	if group == "" {
		t.Error("expected non-empty relay group")
	}
}

func TestChannelPolicySaaSShared(t *testing.T) {
	policy := domainrelay.NewSaaSSharedChannelPolicy("shared-group")
	group := policy.ResolveRelayGroup(nil, "dept-123")
	if group != "shared-group" {
		t.Errorf("expected 'shared-group', got %q", group)
	}
}

func TestNewChannelPolicy(t *testing.T) {
	t.Run("saas mode returns shared policy", func(t *testing.T) {
		cfg := config.Config{SupportSaas: true, PlatformSharedRelayGroup: "my-shared"}
		policy := domainrelay.NewChannelPolicy(cfg)
		group := policy.ResolveRelayGroup(nil, "any-dept")
		if group != "my-shared" {
			t.Errorf("expected 'my-shared', got %q", group)
		}
	})

	t.Run("local mode returns local policy", func(t *testing.T) {
		cfg := config.Config{SupportSaas: false}
		policy := domainrelay.NewChannelPolicy(cfg)
		group := policy.ResolveRelayGroup(nil, "dept-abc")
		if group == "" {
			t.Error("expected non-empty relay group for local policy")
		}
	})
}

func floatPtr(v float64) *float64 { return &v }
