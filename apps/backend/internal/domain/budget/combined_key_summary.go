package budget

import (
	"context"
	"math"

	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

// ComputeGatewaySummaryUpdates loads budget context once and returns combined key summary
// updates for touched platform keys.
func ComputeGatewaySummaryUpdates(
	ctx context.Context,
	st store.Store,
	keyIDs map[string]struct{},
	clk clock.Clock,
) ([]store.CombinedKeySummaryUpdate, error) {
	if len(keyIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(keyIDs))
	for id := range keyIDs {
		ids = append(ids, id)
	}

	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, st.BudgetConsumed(), st.Org(), st.Budget(), st.Keys(), clk)
	if err != nil {
		return nil, err
	}
	mappings, err := st.PlatformKeyMappings().ListMappingsByPlatformKeyIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	updates := make([]store.CombinedKeySummaryUpdate, 0, len(mappings))
	for _, mapping := range mappings {
		if mapping.DepartmentID == "" {
			continue
		}
		open, err := pkgbudget.OpenDepartmentPeriod(ctx, st.Org().Nodes(), mapping.DepartmentID, clk)
		if err != nil {
			return nil, err
		}
		remain, err := pkgbudget.ComputeRemainForMapping(
			ctx, budgetCtx, st.BudgetConsumed(), st.Org(), st.Budget(), st.Company(), mapping, open.String(),
		)
		if err != nil {
			continue
		}
		if remain >= math.MaxFloat64 {
			continue // No budget constraints — leave combined_key_remain as NULL (allow).
		}
		updates = append(updates, store.CombinedKeySummaryUpdate{
			PlatformKeyID: mapping.PlatformKeyID,
			Remain:    remain,
		})
	}
	return updates, nil
}

// RefreshPlatformKeyCombined recomputes and persists combined_key_remain for one platform key.
func RefreshPlatformKeyCombined(ctx context.Context, st store.Store, keyID string, clk clock.Clock, cache CombinedKeyCache) error {
	updates, err := ComputeGatewaySummaryUpdates(ctx, st, map[string]struct{}{keyID: {}}, clk)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}
	summaries, err := st.CombinedKeySummaries().UpdateBatch(ctx, updates)
	if err != nil {
		return err
	}
	coID := store.CompanyID(ctx)
	RefreshCombinedKeySummaries(ctx, cache, nil, coID, summaries)
	return nil
}

// AffectedPlatformKeyIDs resolves platform keys whose combined key summary may
// change after consumed drift repair on the given axis keys.
func AffectedPlatformKeyIDs(ctx context.Context, st store.Store, repaired map[AxisKey]struct{}) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	for key := range repaired {
		var mappings []store.PlatformKeyMapping
		var err error
		switch key.Kind {
		case store.AxisKindPlatformKey:
			out[key.AxisID] = struct{}{}
			continue
		case store.AxisKindMember:
			mappings, err = st.PlatformKeyMappings().ListMappingsByMemberID(ctx, key.AxisID)
		case store.AxisKindProject:
			mappings, err = st.PlatformKeyMappings().ListMappingsByProjectID(ctx, key.AxisID)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, mapping := range mappings {
			out[mapping.PlatformKeyID] = struct{}{}
		}
	}
	return out, nil
}
