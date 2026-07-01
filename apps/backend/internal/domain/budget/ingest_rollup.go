package budget

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
)

func rollupDepartmentConsumed(tree []types.BudgetNode, leafDepartmentID string, costCNY float64) error {
	node := pkgbudget.FindBudgetNode(tree, leafDepartmentID)
	if node == nil {
		return domain.NotFound(fmt.Sprintf("department node not found: %s", leafDepartmentID))
	}
	node.Consumed += costCNY
	ancestors := collectAncestorIDs(tree, leafDepartmentID)
	for _, ancestorID := range ancestors {
		ancestor := pkgbudget.FindBudgetNode(tree, ancestorID)
		if ancestor != nil {
			ancestor.Consumed += costCNY
		}
	}
	return nil
}

func collectAncestorIDs(tree []types.BudgetNode, leafID string) []string {
	var ancestors []string
	var walk func(nodes []types.BudgetNode, path []string) bool
	walk = func(nodes []types.BudgetNode, path []string) bool {
		for _, node := range nodes {
			nextPath := append(path, node.ID)
			if node.ID == leafID {
				ancestors = append([]string{}, path...)
				return true
			}
			if len(node.Children) > 0 && walk(node.Children, nextPath) {
				return true
			}
		}
		return false
	}
	walk(tree, nil)
	return ancestors
}

func usageHourFromPayload(payload newapi.WebhookLogPayload) time.Time {
	ts := payload.CreatedAt
	if ts <= 0 {
		ts = time.Now().Unix()
	}
	return time.Unix(ts, 0).UTC().Truncate(time.Hour)
}
