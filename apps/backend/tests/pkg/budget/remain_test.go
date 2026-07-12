package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
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
		projects     []types.Project
		departmentID string
		want         float64
	}{
		{
			name: "key remaining only (no project or member)",
			key: types.PlatformKey{
				ID:       "k1",
				Budget:   100,
				Consumed: 30,
			},
			tree:         nil,
			members:      nil,
			platformKeys: nil,
			projects:     nil,
			departmentID: "d-unknown",
			want:         70,
		},
		{
			name: "key remaining limited by project",
			key: types.PlatformKey{
				ID:        "k1",
				Budget:    200,
				Consumed:  50,
				ProjectID: &groupID,
			},
			tree: nil,
			projects: []types.Project{
				{ID: "g1", Budget: 80, Consumed: 60},
			},
			departmentID: "d-unknown",
			want:         20,
		},
		{
			name: "key remaining limited by department budget",
			key: types.PlatformKey{
				ID:       "k1",
				Budget:   1000,
				Consumed: 0,
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
				ID:       "k1",
				Budget:   1000,
				Consumed: 0,
			},
			tree: []types.BudgetNode{
				{ID: "d1", Budget: 500, Consumed: 400, ReservedPool: budgetfix.FloatPtr(30)},
			},
			departmentID: "d1",
			want:         70,
		},
		{
			name: "negative remaining clamped to zero",
			key: types.PlatformKey{
				ID:       "k1",
				Budget:   10,
				Consumed: 20,
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
				Consumed: 0,
				MemberID: &memberID,
			},
			tree: nil,
			members: []types.Member{
				{ID: "m1", PersonalBudget: 100},
			},
			platformKeys: []types.PlatformKey{
				{ID: "k1", MemberID: &memberID, Status: "active", Consumed: 80},
			},
			departmentID: "d-unknown",
			want:         20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkgbudget.ComputeRemainBudget(
				tt.key, tt.tree, tt.members, tt.platformKeys, tt.projects, tt.departmentID, nil, nil,
			)
			if got != tt.want {
				t.Errorf("ComputeRemainBudget() = %v, want %v", got, tt.want)
			}
		})
	}
}
