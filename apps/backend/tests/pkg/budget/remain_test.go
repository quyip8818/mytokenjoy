package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
)

func TestComputeRemainBudget(t *testing.T) {
	t.Parallel()
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
				ID:     "k1",
				Budget: 100,
				Used:   30,
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
				Budget:        200,
				Used:          50,
				BudgetGroupID: &groupID,
			},
			tree: nil,
			groups: []types.BudgetGroup{
				{ID: "g1", Budget: 80, Consumed: 60},
			},
			departmentID: "d-unknown",
			want:         20,
		},
		{
			name: "key remaining limited by department budget",
			key: types.PlatformKey{
				ID:     "k1",
				Budget: 1000,
				Used:   0,
			},
			tree: []types.BudgetNode{
				{ID: "d1", Budget: 500, Consumed: 450},
			},
			departmentID: "d1",
			want:         50,
		},
		{
			name: "department reserved pool reduces available",
			key: types.PlatformKey{
				ID:     "k1",
				Budget: 1000,
				Used:   0,
			},
			tree: []types.BudgetNode{
				{ID: "d1", Budget: 500, Consumed: 400, ReservedPool: floatPtr(30)},
			},
			departmentID: "d1",
			want:         70,
		},
		{
			name: "negative remaining clamped to zero",
			key: types.PlatformKey{
				ID:     "k1",
				Budget: 10,
				Used:   20,
			},
			tree:         nil,
			departmentID: "d-unknown",
			want:         0,
		},
		{
			name: "member budget limits result",
			key: types.PlatformKey{
				ID:       "k1",
				Budget:   500,
				Used:     0,
				MemberID: &memberID,
			},
			tree: nil,
			members: []types.Member{
				{ID: "m1", PersonalBudget: 100},
			},
			platformKeys: []types.PlatformKey{
				{ID: "k1", MemberID: &memberID, Status: "active", Used: 80},
			},
			departmentID: "d-unknown",
			want:         20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkgbudget.ComputeRemainBudget(
				tt.key, tt.tree, tt.members, tt.platformKeys, tt.groups, tt.departmentID, nil, nil,
			)
			if got != tt.want {
				t.Errorf("ComputeRemainBudget() = %v, want %v", got, tt.want)
			}
		})
	}
}
