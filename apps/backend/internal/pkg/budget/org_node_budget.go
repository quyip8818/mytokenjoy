package budget

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func OrgNodeBudgetRowFromNode(node types.OrgNode) store.OrgNodeBudgetRow {
	period := node.Period
	if period == "" {
		period = PeriodMonthly
	}
	return store.OrgNodeBudgetRow{
		NodeID:          node.ID,
		Budget:          node.Budget,
		ReservedPool:    node.ReservedPool,
		Period:          period,
		MemberAvgBudget: node.MemberAvgBudget,
	}
}

func OrgNodeBudgetRowsFromNodes(nodes []types.OrgNode) []store.OrgNodeBudgetRow {
	flat := pkgorg.FlattenOrgNodeTree(nodes)
	rows := make([]store.OrgNodeBudgetRow, len(flat))
	for i, node := range flat {
		rows[i] = OrgNodeBudgetRowFromNode(node)
	}
	return rows
}

func PersistNodeBudget(ctx context.Context, repo store.OrgNodeBudgetRepository, nodeID uuid.UUID, node types.BudgetNode) error {
	existing, _, err := repo.Get(ctx, nodeID)
	if err != nil {
		return err
	}
	period := existing.Period
	if period == "" {
		period = PeriodMonthly
	}
	return repo.Upsert(ctx, nodeID, store.OrgNodeBudgetRow{
		Budget:          node.Budget,
		ReservedPool:    node.ReservedPool,
		Period:          period,
		MemberAvgBudget: existing.MemberAvgBudget,
	})
}

func PersistMemberAvgBudget(ctx context.Context, repo store.OrgNodeBudgetRepository, nodeID uuid.UUID, memberAvg float64) error {
	existing, found, err := repo.Get(ctx, nodeID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("org node budget not found: %s", nodeID)
	}
	existing.MemberAvgBudget = memberAvg
	return repo.Upsert(ctx, nodeID, existing)
}
