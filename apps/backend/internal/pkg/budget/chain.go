package budget

import (
	"context"
	"math"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

const (
	LimitingPlatformKey   = "platform_key"
	LimitingMember        = "member"
	LimitingProject       = "project"
	LimitingProjectMember = "project_member"
	LimitingWallet        = "wallet"
)

type ChainInputs struct {
	KeyBudget        float64
	KeyConsumed      float64
	WalletRemain     float64
	PersonalCap      float64
	PersonalConsumed float64
	ProjectCap       float64
	ProjectConsumed  float64
	MemberBudget     float64
	SubConsumed      float64
}

func GatewayChainRemain(scope string, in ChainInputs) (remain float64, limiting string) {
	type candidate struct {
		val  float64
		name string
	}
	candidates := []candidate{{in.WalletRemain, LimitingWallet}}

	if in.KeyBudget > 0 {
		candidates = append(candidates, candidate{
			ClampNonNegative(in.KeyBudget - in.KeyConsumed),
			LimitingPlatformKey,
		})
	}

	switch scope {
	case types.PlatformKeyScopeMember:
		candidates = append(candidates, candidate{
			ClampNonNegative(in.PersonalCap - in.PersonalConsumed),
			LimitingMember,
		})
	case types.PlatformKeyScopeProject:
		candidates = append(candidates, candidate{
			ClampNonNegative(in.ProjectCap - in.ProjectConsumed),
			LimitingProject,
		})
	case types.PlatformKeyScopeProjectMember:
		candidates = append(candidates,
			candidate{
				ClampNonNegative(in.MemberBudget - in.SubConsumed),
				LimitingProjectMember,
			},
			candidate{
				ClampNonNegative(in.ProjectCap - in.ProjectConsumed),
				LimitingProject,
			},
		)
	}

	if len(candidates) == 0 {
		return 0, LimitingWallet
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.val < best.val {
			best = c
		}
	}
	if math.IsInf(best.val, 1) {
		return in.WalletRemain, LimitingWallet
	}
	return best.val, best.name
}

func SumProjectMemberKeyConsumed(keys []types.PlatformKey, projectID, memberID string) float64 {
	var sum float64
	for _, key := range keys {
		if key.Scope != types.PlatformKeyScopeProjectMember {
			continue
		}
		if key.ProjectID == nil || *key.ProjectID != projectID {
			continue
		}
		if key.MemberID == nil || *key.MemberID != memberID {
			continue
		}
		sum += key.Consumed
	}
	return sum
}

func SumProjectMemberKeyConsumedFromRepo(
	ctx context.Context,
	consumed store.BudgetConsumedRepository,
	keys []types.PlatformKey,
	projectID, memberID, periodKey string,
) (float64, error) {
	var sum float64
	for _, key := range keys {
		if key.Scope != types.PlatformKeyScopeProjectMember {
			continue
		}
		if key.ProjectID == nil || *key.ProjectID != projectID {
			continue
		}
		if key.MemberID == nil || *key.MemberID != memberID {
			continue
		}
		used, found, err := consumed.GetConsumed(ctx, store.AxisKindPlatformKey, key.ID, periodKey)
		if err != nil {
			return 0, err
		}
		if found {
			sum += used
		}
	}
	return sum, nil
}
