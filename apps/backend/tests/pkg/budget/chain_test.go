package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
)

func TestGatewayChainRemain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		scope string
		in    pkgbudget.ChainInputs
		want  int64
		limit string
	}{
		{
			name:  "member limited by personal",
			scope: types.PlatformKeyScopeMember,
			in: pkgbudget.ChainInputs{
				KeyBudget: 500, KeyConsumed: 100,
				WalletRemain: 10000,
				PersonalCap:  1000, PersonalConsumed: 950,
			},
			want:  50,
			limit: pkgbudget.LimitingMember,
		},
		{
			name:  "project_member sub bottleneck",
			scope: types.PlatformKeyScopeProjectMember,
			in: pkgbudget.ChainInputs{
				KeyBudget: 300, KeyConsumed: 0,
				WalletRemain: 10000,
				ProjectCap:   5000, ProjectConsumed: 1200,
				MemberBudget: 500, SubConsumed: 480,
			},
			want:  20,
			limit: pkgbudget.LimitingProjectMember,
		},
		{
			name:  "wallet no longer caps chain",
			scope: types.PlatformKeyScopeProject,
			in: pkgbudget.ChainInputs{
				KeyBudget: 1000, KeyConsumed: 0,
				WalletRemain: 10,
				ProjectCap:   5000, ProjectConsumed: 0,
			},
			want:  1000,
			limit: pkgbudget.LimitingPlatformKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, limiting := pkgbudget.GatewayChainRemain(tt.scope, tt.in)
			if got != tt.want {
				t.Fatalf("remain = %v, want %v", got, tt.want)
			}
			if limiting != tt.limit {
				t.Fatalf("limiting = %q, want %q", limiting, tt.limit)
			}
		})
	}
}
