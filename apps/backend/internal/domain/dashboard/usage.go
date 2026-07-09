package dashboard

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
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
	rows, err := s.reader.QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByModel, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, row := range rows {
		total += row.Cost
	}
	catalog, err := s.store.Models().Models(ctx)
	if err != nil {
		return nil, err
	}
	catalogByType := make(map[string]types.ModelInfo, len(catalog))
	for _, model := range catalog {
		existing, ok := catalogByType[model.Type]
		if !ok {
			catalogByType[model.Type] = model
			continue
		}
		if model.IsCustom() && !existing.IsCustom() {
			catalogByType[model.Type] = model
		}
	}
	result := make([]types.ModelUsage, 0, len(rows))
	for _, row := range rows {
		provider := ""
		displayName := row.Model
		if model, ok := catalogByType[row.Model]; ok {
			provider = string(model.Provider)
			displayName = model.Name
			if displayName == "" {
				displayName = model.Type
			}
		}
		pct := 0.0
		if total > 0 {
			pct = row.Cost / total * 100
		}
		result = append(result, types.ModelUsage{
			CallType: row.Model, ModelName: displayName, Provider: provider,
			Requests: float64(row.CallCount), Tokens: 0, Cost: row.Cost, Percentage: pct,
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
	deptTree, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	departments := org.FlattenDepartmentTree(deptTree)
	tree, err := common.LoadBudgetTree(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.reader.QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByDepartment, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	consumedByDept := make(map[string]float64, len(rows))
	for _, row := range rows {
		consumedByDept[row.DepartmentID] = row.Cost
	}
	deptIDs := make([]string, 0, len(departments))
	for _, dept := range departments {
		deptIDs = append(deptIDs, dept.ID)
	}
	topModels, err := s.reader.TopModelsByDepartments(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone, ScopeDeptIDs: scopeDeptIDs,
	}, deptIDs)
	if err != nil {
		return nil, err
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
		result = append(result, types.TeamUsage{
			DepartmentID: dept.ID, DepartmentName: dept.Name,
			Quota: quota, Consumed: consumedByDept[dept.ID],
			MemberCount: memberCount, TopModel: topModels[dept.ID],
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
