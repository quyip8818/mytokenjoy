package dashboard

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
)

func (s *service) ModelUsage(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.ModelUsage, error) {
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, "")
	if err != nil {
		return nil, err
	}
	rows, err := s.store.Usage().QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByModel, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, row := range rows {
		total += row.CostCNY
	}
	catalog := s.store.Models().Models()
	result := make([]types.ModelUsage, 0, len(rows))
	for _, row := range rows {
		provider := ""
		displayName := row.Model
		for _, model := range catalog {
			if model.Name == row.Model {
				provider = string(model.Provider)
				displayName = model.DisplayName
				if displayName == "" {
					displayName = model.Name
				}
				break
			}
		}
		pct := 0.0
		if total > 0 {
			pct = row.CostCNY / total * 100
		}
		result = append(result, types.ModelUsage{
			ModelID: row.Model, ModelName: displayName, Provider: provider,
			Requests: float64(row.CallCount), Tokens: 0, Cost: row.CostCNY, Percentage: pct,
		})
	}
	return result, nil
}

func (s *service) TeamUsage(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.TeamUsage, error) {
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, "")
	if err != nil {
		return nil, err
	}
	departments := orgutil.FlattenDepartmentTree(s.store.Org().Departments())
	tree := s.store.Budget().Tree()
	members := s.store.Org().Members()
	rows, err := s.store.Usage().QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByDepartment, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	consumedByDept := make(map[string]float64, len(rows))
	for _, row := range rows {
		consumedByDept[row.DepartmentID] = row.CostCNY
	}
	result := make([]types.TeamUsage, 0, len(departments))
	for _, dept := range departments {
		quota := 0.0
		if node := findBudgetNode(tree, dept.ID); node != nil {
			quota = node.Budget
		}
		memberCount := 0.0
		for _, member := range members {
			if member.DepartmentID == dept.ID {
				memberCount++
			}
		}
		topModel := ""
		topRows, err := s.store.Usage().QueryAggregates(ctx, types.UsageAggregateQuery{
			Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
			GroupBy: types.UsageGroupByModel, DepartmentID: dept.ID, Limit: 1,
			ScopeDeptIDs: scopeDeptIDs,
		})
		if err == nil && len(topRows) > 0 {
			topModel = topRows[0].Model
		}
		result = append(result, types.TeamUsage{
			DepartmentID: dept.ID, DepartmentName: dept.Name,
			Quota: quota, Consumed: consumedByDept[dept.ID],
			MemberCount: memberCount, TopModel: topModel,
		})
	}
	return result, nil
}

func findBudgetNode(tree []types.BudgetNode, id string) *types.BudgetNode {
	for i := range tree {
		if tree[i].ID == id {
			return &tree[i]
		}
		if len(tree[i].Children) > 0 {
			if found := findBudgetNode(tree[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}
