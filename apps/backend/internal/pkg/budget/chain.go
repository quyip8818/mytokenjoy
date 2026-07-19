package budget

import (
	"context"
	"math"

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
	KeyBudget        int64
	KeyConsumed      int64
	PersonalCap      int64
	PersonalConsumed int64
	ProjectCap       int64
	ProjectConsumed  int64
	MemberBudget     int64
	SubConsumed      int64
}

func GatewayChainRemain(scope string, in ChainInputs) (remain int64, limiting string) {
	type candidate struct {
		val  int64
		name string
	}
	// wallet_remain is checked independently in the precheck path (real-time from PG).
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
		// No budget constraints configured — uncapped by management rules.
		return math.MaxInt64, ""
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.val < best.val {
			best = c
		}
	}
	return best.val, best.name
}

func clampNonNegative(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}

func SumProjectMemberKeyConsumed(keys []types.PlatformKey, projectID, memberID uuid.UUID) int64 {
	var sum int64
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
) (int64, error) {
	var sum int64
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
