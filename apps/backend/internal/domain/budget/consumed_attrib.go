package budget

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type AxisKind = string

type AxisKey struct {
	Kind      AxisKind
	AxisID    uuid.UUID
	PeriodKey string
}

type AxisDelta struct {
	Kind      AxisKind
	AxisID    uuid.UUID
	PeriodKey string
	Amount    int64
}

// ConsumptionDeltas computes the budget_consumed axis increments for a single entry.
func ConsumptionDeltas(_ context.Context, _ store.OrgNodeRepository, entry types.UsageLedgerEntry, open pkgbudget.OpenBudgetPeriod) ([]AxisDelta, error) {
	if open.IsZero() {
		return nil, fmt.Errorf("consumption deltas require open budget period")
	}
	periodKey := open.String()
	deltas := []AxisDelta{{
		Kind: store.AxisKindPlatformKey, AxisID: entry.PlatformKeyID, PeriodKey: periodKey, Amount: entry.Amount,
	}}
	scope := entry.PlatformKeyScope
	switch scope {
	case types.PlatformKeyScopeMember:
		if entry.MemberID != nil {
			deltas = append(deltas, AxisDelta{
				Kind: store.AxisKindMember, AxisID: *entry.MemberID, PeriodKey: periodKey, Amount: entry.Amount,
			})
		}
	case types.PlatformKeyScopeProject, types.PlatformKeyScopeProjectMember:
		if entry.ProjectID != nil {
			deltas = append(deltas, AxisDelta{
				Kind: store.AxisKindProject, AxisID: *entry.ProjectID, PeriodKey: periodKey, Amount: entry.Amount,
			})
		}
	default:
		return nil, fmt.Errorf("unknown platform key scope %q", scope)
	}
	return deltas, nil
}

func ConsumedDrift(expected, actual int64) bool {
	return expected != actual
}
