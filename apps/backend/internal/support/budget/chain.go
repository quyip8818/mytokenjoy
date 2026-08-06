package budget

import (
	"context"

	"github.com/google/uuid"
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
	PersonalCap      float64
	PersonalConsumed float64
	ProjectCap       float64
	ProjectConsumed  float64
	MemberBudget     float64
	SubConsumed      float64
}

// GatewayChainRemain returns the tightest budget remain for the key scope.
// When uncapped is true, no management budget constraints apply — callers must
// leave combined_key_remain as NULL (allow) rather than persisting a sentinel.
func GatewayChainRemain(scope string, in ChainInputs) (remain float64, limiting string, uncapped bool) {
	type candidate struct {
		val  float64
		name string
	}
	// wallet_remain_quota is checked independently in the precheck path (real-time from PG).
	// This chain only evaluates budget-control constraints: key, member, project.
	var candidates []candidate

	if in.KeyBudget > 0 {
		candidates = append(candidates, candidate{
			clampNonNegative(in.KeyBudget - in.KeyConsumed),
			LimitingPlatformKey,
		})
	}

	switch scope {
	case types.PlatformKeyScopeMember:
		candidates = append(candidates, candidate{
			clampNonNegative(in.PersonalCap - in.PersonalConsumed),
			LimitingMember,
		})
	case types.PlatformKeyScopeProject:
		candidates = append(candidates, candidate{
			clampNonNegative(in.ProjectCap - in.ProjectConsumed),
			LimitingProject,
		})
	case types.PlatformKeyScopeProjectMember:
		candidates = append(candidates,
			candidate{
				clampNonNegative(in.MemberBudget - in.SubConsumed),
				LimitingProjectMember,
			},
			candidate{
				clampNonNegative(in.ProjectCap - in.ProjectConsumed),
				LimitingProject,
			},
		)
	}

	if len(candidates) == 0 {
		return 0, "", true
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.val < best.val {
			best = c
		}
	}
	return best.val, best.name, false
}

func clampNonNegative(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func SumProjectMemberKeyConsumed(keys []types.PlatformKey, projectID, memberID uuid.UUID) float64 {
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
	projectID, memberID uuid.UUID, periodKey string,
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
