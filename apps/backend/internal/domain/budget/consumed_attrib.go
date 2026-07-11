package budget

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

type AxisKind = string

type AxisKey struct {
	Kind      AxisKind
	AxisID    string
	PeriodKey string
}

type AxisDelta struct {
	Kind      AxisKind
	AxisID    string
	PeriodKey string
	Amount    float64
}

type ConsumedIncrementWriter interface {
	IncrementConsumed(ctx context.Context, axisKind, axisID, periodKey string, amount float64) error
	RollupOrgNodeAncestors(ctx context.Context, leafNodeID, periodKey string, amount float64) error
}

func ConsumptionDeltas(ctx context.Context, nodes store.OrgNodeRepository, entry types.UsageLedgerEntry, open pkgbudget.OpenBudgetPeriod) ([]AxisDelta, error) {
	if open.IsZero() {
		return nil, fmt.Errorf("consumption deltas require open budget period")
	}
	periodKey := open.String()
	deltas := []AxisDelta{{
		Kind: store.AxisKindPlatformKey, AxisID: entry.PlatformKeyID, PeriodKey: periodKey, Amount: entry.Amount,
	}}
	if entry.BudgetGroupID != nil {
		deltas = append(deltas, AxisDelta{
			Kind: store.AxisKindBudgetGroup, AxisID: *entry.BudgetGroupID, PeriodKey: periodKey, Amount: entry.Amount,
		})
	}
	if entry.MemberID != nil {
		deltas = append(deltas, AxisDelta{
			Kind: store.AxisKindMember, AxisID: *entry.MemberID, PeriodKey: periodKey, Amount: entry.Amount,
		})
	}
	if entry.DepartmentID != "" {
		ancestorIDs, err := nodes.ListSelfAndAncestorIDs(ctx, entry.DepartmentID)
		if err != nil {
			return nil, err
		}
		for _, id := range ancestorIDs {
			deltas = append(deltas, AxisDelta{
				Kind: store.AxisKindOrgNode, AxisID: id, PeriodKey: periodKey, Amount: entry.Amount,
			})
		}
	}
	return deltas, nil
}

func ApplyIncrement(ctx context.Context, writer ConsumedIncrementWriter, nodes store.OrgNodeRepository, entry types.UsageLedgerEntry, open pkgbudget.OpenBudgetPeriod) error {
	if open.IsZero() {
		return fmt.Errorf("apply increment requires open budget period")
	}
	periodKey := open.String()
	if err := writer.IncrementConsumed(ctx, store.AxisKindPlatformKey, entry.PlatformKeyID, periodKey, entry.Amount); err != nil {
		return err
	}
	if entry.BudgetGroupID != nil {
		if err := writer.IncrementConsumed(ctx, store.AxisKindBudgetGroup, *entry.BudgetGroupID, periodKey, entry.Amount); err != nil {
			return err
		}
	}
	if entry.MemberID != nil {
		if err := writer.IncrementConsumed(ctx, store.AxisKindMember, *entry.MemberID, periodKey, entry.Amount); err != nil {
			return err
		}
	}
	if entry.DepartmentID != "" {
		if err := writer.RollupOrgNodeAncestors(ctx, entry.DepartmentID, periodKey, entry.Amount); err != nil {
			return err
		}
	}
	return nil
}

func ExpectedConsumed(ctx context.Context, nodes store.OrgNodeRepository, entries []types.UsageLedgerEntry, clk clock.Clock) (map[AxisKey]float64, error) {
	acc := make(map[AxisKey]float64)
	for _, entry := range entries {
		open, err := pkgbudget.OpenDepartmentPeriod(ctx, nodes, entry.DepartmentID, clk)
		if err != nil {
			return nil, err
		}
		deltas, err := ConsumptionDeltas(ctx, nodes, entry, open)
		if err != nil {
			return nil, err
		}
		for _, d := range deltas {
			key := AxisKey{Kind: d.Kind, AxisID: d.AxisID, PeriodKey: d.PeriodKey}
			acc[key] += d.Amount
		}
	}
	return acc, nil
}

const reconcileEpsilon = 0.000001

func ConsumedDrift(expected, actual float64) bool {
	diff := expected - actual
	if diff < 0 {
		diff = -diff
	}
	return diff > reconcileEpsilon
}
