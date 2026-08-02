package usage

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type EntryBuildSnapshot struct {
	Catalog      []types.ModelInfo
	OrgTree      []types.OrgNode
	Discounts    []store.ModelDiscountRow // per-company discount coefficients
	QuotaPerUnit int64                    // company billing QPU
}

func LoadEntryBuildSnapshot(ctx context.Context, deps EntryBuildReader, tokenJoyCompanyID uuid.UUID) (EntryBuildSnapshot, error) {
	catalog, err := deps.Models().Models(ctx)
	if err != nil {
		return EntryBuildSnapshot{}, err
	}
	tree, err := deps.Org().Nodes().Tree(ctx)
	if err != nil {
		return EntryBuildSnapshot{}, err
	}

	companyID := company.CompanyID(ctx)

	discounts, _ := deps.ModelDiscount().CurrentDiscounts(ctx, companyID)

	qpu := resolveQPU(ctx, deps, companyID)

	return EntryBuildSnapshot{
		Catalog:      catalog,
		OrgTree:      tree,
		Discounts:    discounts,
		QuotaPerUnit: qpu,
	}, nil
}

// resolveQPU looks up the company's billing QPU without importing domain/billing (avoid cycle).
func resolveQPU(ctx context.Context, deps EntryBuildReader, companyID uuid.UUID) int64 {
	co, err := deps.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		return common.DefaultQuotaPerUnit
	}
	currency := common.ResolveBillingCurrency(co.BillingCurrency)
	cur, err := deps.Billing().GetCurrency(ctx, currency)
	if err != nil || cur == nil || cur.QuotaPerUnit <= 0 {
		return common.DefaultQuotaPerUnit
	}
	return cur.QuotaPerUnit
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
	return effectiveWhitelistIDs(keyIDs, deptAllowed)
}

func effectiveWhitelistIDs(keyWhitelist, deptAllowed []uuid.UUID) []uuid.UUID {
	if len(keyWhitelist) == 0 {
		return append([]uuid.UUID{}, deptAllowed...)
	}
	allowed := make(map[uuid.UUID]struct{}, len(deptAllowed))
	for _, id := range deptAllowed {
		allowed[id] = struct{}{}
	}
	out := make([]uuid.UUID, 0, len(keyWhitelist))
	for _, id := range keyWhitelist {
		if _, ok := allowed[id]; ok {
			out = append(out, id)
		}
	}
	return out
}
