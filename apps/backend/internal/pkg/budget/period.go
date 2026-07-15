package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

// OpenBudgetPeriod is the open-budget (gate) period_key derived from business Clock.
// It must not be constructed from OccurredAt.
type OpenBudgetPeriod struct{ key string }

// OccurrencePeriod is the ledger occurrence period_key derived from OccurredAt.
// It must not be used as a budget_consumed open key.
type OccurrencePeriod struct{ key string }

func (p OpenBudgetPeriod) String() string { return p.key }
func (p OccurrencePeriod) String() string { return p.key }

func (p OpenBudgetPeriod) IsZero() bool { return p.key == "" }
func (p OccurrencePeriod) IsZero() bool { return p.key == "" }

// TestOpenBudgetPeriod creates an OpenBudgetPeriod with a given key for testing.
func TestOpenBudgetPeriod(key string) OpenBudgetPeriod { return OpenBudgetPeriod{key: key} }

// OpenDepartmentPeriod resolves the open-budget period for a department from the business clock.
func OpenDepartmentPeriod(ctx context.Context, nodes store.OrgNodeRepository, departmentID string, clk clock.Clock) (OpenBudgetPeriod, error) {
	key, err := departmentPeriodKey(ctx, nodes, departmentID, clock.NowUTC(clk))
	if err != nil {
		return OpenBudgetPeriod{}, err
	}
	return OpenBudgetPeriod{key: key}, nil
}

// OpenDepartmentPeriodAt resolves the open-budget period for a department at an
// explicit point in time. Used by reconcile to reproduce the period that was active
// when an entry was originally ingested.
func OpenDepartmentPeriodAt(ctx context.Context, nodes store.OrgNodeRepository, departmentID string, at time.Time) (OpenBudgetPeriod, error) {
	key, err := departmentPeriodKey(ctx, nodes, departmentID, at)
	if err != nil {
		return OpenBudgetPeriod{}, err
	}
	return OpenBudgetPeriod{key: key}, nil
}

// OccurrenceDepartmentPeriodFromTree resolves the ledger occurrence period from a pre-loaded org tree.
func OccurrenceDepartmentPeriodFromTree(orgTree []types.OrgNode, departmentID string, occurredAt time.Time) (OccurrencePeriod, error) {
	orgPeriod := PeriodMonthly
	if departmentID != "" {
		if node := pkgorg.FindOrgNode(orgTree, departmentID); node != nil {
			orgPeriod = node.Period
		}
	}
	return OccurrencePeriod{key: SnapshotKey(orgPeriod, occurredAt)}, nil
}

// OccurrenceDepartmentPeriod resolves the ledger occurrence period from event time.
func OccurrenceDepartmentPeriod(ctx context.Context, nodes store.OrgNodeRepository, departmentID string, occurredAt time.Time) (OccurrencePeriod, error) {
	key, err := departmentPeriodKey(ctx, nodes, departmentID, occurredAt)
	if err != nil {
		return OccurrencePeriod{}, err
	}
	return OccurrencePeriod{key: key}, nil
}

// OpenSnapshotKey resolves an open-budget period from an org period spec and business clock.
func OpenSnapshotKey(orgPeriod string, clk clock.Clock) OpenBudgetPeriod {
	return OpenBudgetPeriod{key: SnapshotKey(orgPeriod, clock.NowUTC(clk))}
}

// OccurrenceSnapshotKey resolves a ledger occurrence period from an org period spec and event time.
func OccurrenceSnapshotKey(orgPeriod string, occurredAt time.Time) OccurrencePeriod {
	return OccurrencePeriod{key: SnapshotKey(orgPeriod, occurredAt)}
}

// RootPeriodKey resolves the open snapshot period_key for the org root at instant at.
// Used by seed apply when SeedAt is already materialised as time.Time.
func RootPeriodKey(nodes []types.OrgNode, at time.Time) string {
	flat := pkgorg.FlattenOrgNodeTree(nodes)
	for _, node := range flat {
		if node.ParentID == nil || *node.ParentID == "" {
			return SnapshotKey(node.Period, at)
		}
	}
	return SnapshotKey(PeriodMonthly, at)
}

func departmentPeriodKey(ctx context.Context, nodes store.OrgNodeRepository, departmentID string, at time.Time) (string, error) {
	if departmentID == "" {
		return SnapshotKey(PeriodMonthly, at), nil
	}
	orgPeriod, found, err := nodes.GetNodePeriod(ctx, departmentID)
	if err != nil {
		return "", err
	}
	if !found {
		return SnapshotKey(PeriodMonthly, at), nil
	}
	return SnapshotKey(orgPeriod, at), nil
}

func buildDeptPeriodMap(ctx context.Context, nodes store.OrgNodeRepository, at time.Time) (map[string]string, string, error) {
	orgNodes, err := nodes.Tree(ctx)
	if err != nil {
		return nil, "", err
	}
	flat := pkgorg.FlattenOrgNodeTree(orgNodes)
	deptPeriod := make(map[string]string, len(flat))
	var rootPeriodKey string
	for _, node := range flat {
		deptPeriod[node.ID] = node.Period
		if node.ParentID == nil || *node.ParentID == "" {
			rootPeriodKey = SnapshotKey(node.Period, at)
		}
	}
	if rootPeriodKey == "" {
		return nil, "", fmt.Errorf("org tree has no root node")
	}
	return deptPeriod, rootPeriodKey, nil
}
