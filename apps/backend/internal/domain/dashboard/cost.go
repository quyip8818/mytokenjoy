package dashboard

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

// roundCost rounds a monetary value to 2 decimal places.
func roundCost(v float64) float64 {
	return math.Round(v*100) / 100
}

func (s *service) CostSummary(ctx context.Context, params types.CostQueryParams, deptID uuid.UUID, scope domainusage.SessionScope) (types.CostSummary, error) {
	current, err := s.resolveRange(params)
	if err != nil {
		return types.CostSummary{}, err
	}
	prev := pkgbudget.PreviousRange(current)
	scopeDeptIDs, err := s.resolveScope(ctx, scope, deptID)
	if err != nil {
		return types.CostSummary{}, err
	}
	base := types.UsageAggregateQuery{Timezone: current.Timezone, ScopeDeptIDs: scopeDeptIDs}
	currentTotals, err := s.reader.QuerySummary(ctx, withRange(base, current))
	if err != nil {
		return types.CostSummary{}, err
	}
	prevTotals, err := s.reader.QuerySummary(ctx, withRange(base, prev))
	if err != nil {
		return types.CostSummary{}, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.CostSummary{}, err
	}
	memberCount := float64(len(members))
	avgCostPerRequest := safeDiv(currentTotals.Spend(), float64(currentTotals.CallCount))
	prevAvgCostPerRequest := safeDiv(prevTotals.Spend(), float64(prevTotals.CallCount))
	avgCostPerMember := safeDiv(currentTotals.Spend(), memberCount)
	prevAvgCostPerMember := safeDiv(prevTotals.Spend(), memberCount)
	return types.CostSummary{
		TotalCost:            roundCost(currentTotals.Spend()),
		TotalCostMom:         roundCost(mom(currentTotals.Spend(), prevTotals.Spend())),
		TotalTokens:          0,
		TotalRequests:        float64(currentTotals.CallCount),
		TotalRequestsMom:     roundCost(mom(float64(currentTotals.CallCount), float64(prevTotals.CallCount))),
		AvgCostPerRequest:    roundCost(avgCostPerRequest),
		AvgCostPerRequestMom: roundCost(mom(avgCostPerRequest, prevAvgCostPerRequest)),
		AvgCostPerMember:     roundCost(avgCostPerMember),
		AvgCostPerMemberMom:  roundCost(mom(avgCostPerMember, prevAvgCostPerMember)),
	}, nil
}

func (s *service) DepartmentCosts(ctx context.Context, parentID uuid.UUID, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCost, error) {
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, uuid.Nil)
	if err != nil {
		return nil, err
	}
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	childIDs := storeChildOwnerDepartmentID(departments, parentID)
	if len(childIDs) == 0 {
		return []types.DepartmentCost{}, nil
	}
	rows, err := s.reader.QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByDepartment, OwnerDepartmentID: childIDs, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, row := range rows {
		total += row.Spend()
	}
	result := make([]types.DepartmentCost, 0, len(rows))
	for _, row := range rows {
		dept := pkgorg.FindDepartment(departments, row.DepartmentID)
		name := row.DepartmentID.String()
		hasChildren := false
		if dept != nil {
			name = dept.Name
			hasChildren = len(dept.Children) > 0
		}
		pct := 0.0
		if total > 0 {
			pct = row.Spend() / total * 100
		}
		result = append(result, types.DepartmentCost{
			DepartmentID: row.DepartmentID, DepartmentName: name,
			Cost: roundCost(row.Spend()), Percentage: roundCost(pct), HasChildren: hasChildren,
		})
	}
	return result, nil
}

func (s *service) DepartmentMemberCosts(ctx context.Context, deptID uuid.UUID, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCostMember, error) {
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	if !domainusage.IsDepartmentAccessible(departments, scope, deptID, s.scopeConfig) {
		return nil, domain.Forbidden("Department not accessible")
	}
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, deptID)
	if err != nil {
		return nil, err
	}
	rows, err := s.reader.QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByMember, DepartmentID: deptID, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]types.DepartmentCostMember, 0, len(rows))
	for _, row := range rows {
		name := row.MemberID.String()
		if member, ok := pkgorg.FindMemberByID(members, row.MemberID); ok {
			name = member.Alias
		}
		result = append(result, types.DepartmentCostMember{
			MemberID: row.MemberID, MemberName: name,
			Cost: roundCost(row.Spend()), Requests: float64(row.CallCount), Tokens: 0,
		})
	}
	return result, nil
}

func (s *service) DailyCosts(ctx context.Context, params types.CostQueryParams, deptID uuid.UUID, scope domainusage.SessionScope) ([]types.DailyCost, error) {
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, deptID)
	if err != nil {
		return nil, err
	}
	rows, err := s.reader.QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Granularity: rng.Granularity, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByNone, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]types.DailyCost, 0, len(rows))
	for _, row := range rows {
		result = append(result, types.DailyCost{
			Date: row.Bucket, Cost: roundCost(row.Spend()), Requests: float64(row.CallCount), Tokens: 0,
		})
	}
	return result, nil
}

func (s *service) TopConsumers(ctx context.Context, limit int, params types.CostQueryParams, deptID uuid.UUID, scope domainusage.SessionScope) ([]types.TopConsumer, error) {
	if limit <= 0 {
		limit = 5
	}
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, deptID)
	if err != nil {
		return nil, err
	}
	rows, err := s.reader.QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: types.UsageGroupByMember, Limit: limit, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	result := make([]types.TopConsumer, 0, len(rows))
	for _, row := range rows {
		name := row.MemberID.String()
		deptName := ""
		if member, ok := pkgorg.FindMemberByID(members, row.MemberID); ok {
			name = member.Alias
			if path := pkgorg.GetDeptPath(departments, member.DepartmentID); path != nil {
				deptName = *path
			}
		}
		result = append(result, types.TopConsumer{
			MemberID: row.MemberID, MemberName: name, Department: deptName,
			Cost: roundCost(row.Spend()), Requests: float64(row.CallCount),
			Tokens: float64(row.InputTokens + row.OutputTokens),
		})
	}
	return result, nil
}

func storeChildOwnerDepartmentID(departments []types.Department, parentID uuid.UUID) []uuid.UUID {
	if parentID == uuid.Nil {
		ids := make([]uuid.UUID, 0, len(departments))
		for _, dept := range departments {
			ids = append(ids, dept.ID)
		}
		return ids
	}
	parent := pkgorg.FindDepartment(departments, parentID)
	if parent == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(parent.Children))
	for _, child := range parent.Children {
		ids = append(ids, child.ID)
	}
	return ids
}
