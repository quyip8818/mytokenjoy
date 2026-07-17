package usage

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

type EntryBuildSnapshot struct {
	Catalog []types.ModelInfo
	OrgTree []types.OrgNode
}

func LoadEntryBuildSnapshot(ctx context.Context, deps EntryBuildReader) (EntryBuildSnapshot, error) {
	catalog, err := deps.Models().Models(ctx)
	if err != nil {
		return EntryBuildSnapshot{}, err
	}
	tree, err := deps.Org().Nodes().Tree(ctx)
	if err != nil {
		return EntryBuildSnapshot{}, err
	}
	return EntryBuildSnapshot{Catalog: catalog, OrgTree: tree}, nil
}

func LoadEntryBuildInput(ctx context.Context, deps EntryBuildReader, mapping *store.PlatformKeyMapping, raw store.RawConsumeLog, source string, snap EntryBuildSnapshot) (EntryBuildInput, error) {
	modelName := ResolveConsumeModel(raw)
	settings, err := deps.Audit().Settings(ctx)
	if err != nil {
		return EntryBuildInput{}, err
	}
	platformKey, err := deps.Keys().PlatformKeyByID(ctx, mapping.PlatformKeyID)
	if err != nil {
		return EntryBuildInput{}, err
	}
	allowedIDs := resolveBillingAllowedIDs(ctx, deps, mapping, platformKey, snap)
	input := EntryBuildInput{
		Raw: raw, Mapping: mapping, Source: source,
		Catalog: snap.Catalog, AllowedIDs: allowedIDs, Settings: settings,
		PlatformKey: platformKey,
	}
	if mapping.MemberID != nil {
		member, err := deps.Org().MemberByID(ctx, *mapping.MemberID)
		if err != nil {
			return EntryBuildInput{}, err
		}
		input.Member = member
	}
	_ = modelName
	return input, nil
}

func resolveBillingAllowedIDs(ctx context.Context, deps EntryBuildReader, mapping *store.PlatformKeyMapping, platformKey *types.PlatformKey, snap EntryBuildSnapshot) []uuid.UUID {
	if platformKey == nil {
		return nil
	}
	keyIDs := append([]uuid.UUID{}, platformKey.ModelWhitelist...)
	orgNodes := cachedOrgNodes{tree: snap.OrgTree}
	departments := types.OrgNodesToDepartments(snap.OrgTree)
	rules, err := common.LoadRoutingRules(ctx, orgNodes, deps.Models().Allowlist())
	if err != nil {
		return keyIDs
	}
	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, departments, rules, snap.Catalog)
	return newapiunits.EffectiveWhitelistIDs(keyIDs, deptAllowed)
}
