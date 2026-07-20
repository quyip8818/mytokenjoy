package budget

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

// MappingStores groups repositories needed to compute remain budget for a platform key mapping.
type MappingStores struct {
	Consumed store.BudgetConsumedRepository
	OrgNodes store.OrgNodeRepository
	Org      store.OrgRepository
	Budget   store.BudgetRepository
	Keys     store.KeysRepository
	Clock    clock.Clock
}

// ComputeRemainForMapping uses a preloaded budget context for batch gateway summary writes.
func ComputeRemainForMapping(
	ctx context.Context,
	budgetCtx BudgetContext,
	consumed store.BudgetConsumedRepository,
	org store.OrgRepository,
	budgetRepo store.BudgetRepository,
	mapping store.PlatformKeyMapping,
	periodKey string,
) (int64, error) {
	if mapping.DepartmentID == uuid.Nil {
		return 0, fmt.Errorf("department not found")
	}
	key, ok := budgetCtx.FindPlatformKey(mapping.PlatformKeyID)
	if !ok {
		return 0, fmt.Errorf("platform key not found")
	}
	if key.Scope == "" {
		return 0, fmt.Errorf("platform key scope missing")
	}
	inputs, err := BuildChainInputs(ctx, budgetCtx, consumed, org, budgetRepo, key, mapping, periodKey)
	if err != nil {
		return 0, err
	}
	remain, _ := GatewayChainRemain(key.Scope, inputs)
	return remain, nil
}

func BuildChainInputs(
	ctx context.Context,
	budgetCtx BudgetContext,
	consumed store.BudgetConsumedRepository,
	org store.OrgRepository,
	budgetRepo store.BudgetRepository,
	key types.PlatformKey,
	mapping store.PlatformKeyMapping,
	periodKey string,
) (ChainInputs, error) {
	inputs := ChainInputs{}
	if key.Budget > 0 {
		keyUsed, found, err := consumed.GetConsumed(ctx, store.AxisKindPlatformKey, key.ID, periodKey)
		if err != nil {
			return ChainInputs{}, err
		}
		if found {
			key.Consumed = keyUsed
		}
		inputs.KeyBudget = key.Budget
		inputs.KeyConsumed = key.Consumed
	}

	switch key.Scope {
	case types.PlatformKeyScopeMember:
		if mapping.MemberID == nil {
			return ChainInputs{}, fmt.Errorf("member mapping required")
		}
		capacity, found, err := org.MemberPersonalBudget(ctx, *mapping.MemberID)
		if err != nil {
			return ChainInputs{}, err
		}
		if found {
			inputs.PersonalCap = capacity
			memberConsumed, _, err := consumed.GetConsumed(ctx, store.AxisKindMember, *mapping.MemberID, periodKey)
			if err != nil {
				return ChainInputs{}, err
			}
			inputs.PersonalConsumed = memberConsumed
		}
	case types.PlatformKeyScopeProject, types.PlatformKeyScopeProjectMember:
		if mapping.ProjectID == nil {
			return ChainInputs{}, fmt.Errorf("project mapping required")
		}
		project, ok := FindProject(budgetCtx.Projects, *mapping.ProjectID)
		if !ok {
			return ChainInputs{}, fmt.Errorf("project not found")
		}
		inputs.ProjectCap = project.Budget
		projectConsumed, _, err := consumed.GetConsumed(ctx, store.AxisKindProject, *mapping.ProjectID, periodKey)
		if err != nil {
			return ChainInputs{}, err
		}
		inputs.ProjectConsumed = projectConsumed
		if key.Scope == types.PlatformKeyScopeProjectMember {
			if mapping.MemberID == nil {
				return ChainInputs{}, fmt.Errorf("member mapping required for project_member")
			}
			memberBudget, found, err := budgetRepo.GetProjectMemberBudget(ctx, *mapping.ProjectID, *mapping.MemberID)
			if err != nil {
				return ChainInputs{}, err
			}
			if !found {
				return ChainInputs{}, fmt.Errorf("project member roster not found")
			}
			inputs.MemberBudget = memberBudget
			inputs.SubConsumed = SumProjectMemberKeyConsumed(budgetCtx.PlatformKeys, *mapping.ProjectID, *mapping.MemberID)
		}
	}
	return inputs, nil
}

// --- Batch path (used by ComputeGatewaySummaryUpdates) ---

// PreloadedConsumed holds batch-loaded consumed data for all three axes, keyed by periodKey then axisID.
type PreloadedConsumed struct {
	Key     map[string]map[uuid.UUID]int64 // periodKey → keyID → consumed
	Member  map[string]map[uuid.UUID]int64 // periodKey → memberID → consumed
	Project map[string]map[uuid.UUID]int64 // periodKey → projectID → consumed
}

// PreloadConsumed batch-loads consumed for all three axes across the given period keys in 3 SQL calls.
func PreloadConsumed(ctx context.Context, consumed store.BudgetConsumedRepository, periodKeys []string) (PreloadedConsumed, error) {
	keys := uniqueStrings(periodKeys)
	keyConsumed, err := consumed.ListConsumedByPeriods(ctx, store.AxisKindPlatformKey, keys)
	if err != nil {
		return PreloadedConsumed{}, err
	}
	memberConsumed, err := consumed.ListConsumedByPeriods(ctx, store.AxisKindMember, keys)
	if err != nil {
		return PreloadedConsumed{}, err
	}
	projectConsumed, err := consumed.ListConsumedByPeriods(ctx, store.AxisKindProject, keys)
	if err != nil {
		return PreloadedConsumed{}, err
	}
	return PreloadedConsumed{Key: keyConsumed, Member: memberConsumed, Project: projectConsumed}, nil
}

func (p PreloadedConsumed) getConsumed(axisKind string, axisID uuid.UUID, periodKey string) int64 {
	var m map[string]map[uuid.UUID]int64
	switch axisKind {
	case store.AxisKindPlatformKey:
		m = p.Key
	case store.AxisKindMember:
		m = p.Member
	case store.AxisKindProject:
		m = p.Project
	default:
		return 0
	}
	if byAxis, ok := m[periodKey]; ok {
		return byAxis[axisID]
	}
	return 0
}

// BuildChainInputsPreloaded constructs ChainInputs entirely from in-memory preloaded data.
// No SQL queries are issued. Returns an error only for missing structural data (scope/mapping inconsistency).
func BuildChainInputsPreloaded(
	budgetCtx BudgetContext,
	preloaded PreloadedConsumed,
	key types.PlatformKey,
	mapping store.PlatformKeyMapping,
	periodKey string,
) (ChainInputs, error) {
	inputs := ChainInputs{}

	if key.Budget > 0 {
		inputs.KeyBudget = key.Budget
		inputs.KeyConsumed = preloaded.getConsumed(store.AxisKindPlatformKey, key.ID, periodKey)
	}

	switch key.Scope {
	case types.PlatformKeyScopeMember:
		if mapping.MemberID == nil {
			return ChainInputs{}, fmt.Errorf("member mapping required")
		}
		member, ok := pkgorg.FindMemberByID(budgetCtx.Members, *mapping.MemberID)
		if !ok {
			break
		}
		inputs.PersonalCap = member.PersonalBudget
		inputs.PersonalConsumed = preloaded.getConsumed(store.AxisKindMember, *mapping.MemberID, periodKey)

	case types.PlatformKeyScopeProject, types.PlatformKeyScopeProjectMember:
		if mapping.ProjectID == nil {
			return ChainInputs{}, fmt.Errorf("project mapping required")
		}
		project, ok := FindProject(budgetCtx.Projects, *mapping.ProjectID)
		if !ok {
			return ChainInputs{}, fmt.Errorf("project not found")
		}
		inputs.ProjectCap = project.Budget
		inputs.ProjectConsumed = preloaded.getConsumed(store.AxisKindProject, *mapping.ProjectID, periodKey)

		if key.Scope == types.PlatformKeyScopeProjectMember {
			if mapping.MemberID == nil {
				return ChainInputs{}, fmt.Errorf("member mapping required for project_member")
			}
			memberBudget, found := project.MemberBudgets[*mapping.MemberID]
			if !found {
				return ChainInputs{}, fmt.Errorf("project member roster not found")
			}
			inputs.MemberBudget = memberBudget
			inputs.SubConsumed = SumProjectMemberKeyConsumed(budgetCtx.PlatformKeys, *mapping.ProjectID, *mapping.MemberID)
		}
	}
	return inputs, nil
}

// ComputeRemainForMappingPreloaded computes remain from preloaded data without issuing any SQL.
func ComputeRemainForMappingPreloaded(
	budgetCtx BudgetContext,
	preloaded PreloadedConsumed,
	mapping store.PlatformKeyMapping,
	periodKey string,
) (int64, error) {
	if mapping.DepartmentID == uuid.Nil {
		return 0, fmt.Errorf("department not found")
	}
	key, ok := budgetCtx.FindPlatformKey(mapping.PlatformKeyID)
	if !ok {
		return 0, fmt.Errorf("platform key not found")
	}
	if key.Scope == "" {
		return 0, fmt.Errorf("platform key scope missing")
	}
	inputs, err := BuildChainInputsPreloaded(budgetCtx, preloaded, key, mapping, periodKey)
	if err != nil {
		return 0, err
	}
	remain, _ := GatewayChainRemain(key.Scope, inputs)
	return remain, nil
}
