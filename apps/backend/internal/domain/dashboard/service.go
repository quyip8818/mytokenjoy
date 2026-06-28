package dashboard

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/pkg/periodutil"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	UsageSeries(ctx context.Context, q types.UsageSeriesQuery, scope domainusage.SessionScope) (types.UsageSeriesResponse, error)
	CostSummary(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) (types.CostSummary, error)
	DepartmentCosts(ctx context.Context, parentID string, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCost, error)
	DepartmentMemberCosts(ctx context.Context, deptID string, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCostMember, error)
	DailyCosts(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DailyCost, error)
	TopConsumers(ctx context.Context, limit int, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.TopConsumer, error)
	ModelUsage(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.ModelUsage, error)
	TeamUsage(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.TeamUsage, error)
}

type service struct {
	cfg          config.Config
	store        store.Store
	logAggregator *domainusage.LogAggregator
	now          func() time.Time
}

func NewService(cfg config.Config, st store.Store, logAggregator *domainusage.LogAggregator) Service {
	return &service{
		cfg:           cfg,
		store:         st,
		logAggregator: logAggregator,
		now:           time.Now,
	}
}

func (s *service) UsageSeries(ctx context.Context, q types.UsageSeriesQuery, scope domainusage.SessionScope) (types.UsageSeriesResponse, error) {
	if q.Timezone == "" {
		q.Timezone = domainusage.ResolveTimezone("")
	}
	if err := domainusage.ValidateGroupBy(q.GroupBy); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	if err := domainusage.ValidateWindow(q.Start, q.End, q.Granularity); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, q.DepartmentID)
	if err != nil {
		return types.UsageSeriesResponse{}, err
	}
	q.ScopeDeptIDs = scopeDeptIDs

	switch q.Granularity {
	case domainusage.GranularityMinute:
		return s.logAggregator.Series(ctx, q)
	case domainusage.GranularityDay, domainusage.GranularityHour:
		points, err := s.store.Usage().QuerySeries(ctx, q)
		if err != nil {
			return types.UsageSeriesResponse{}, err
		}
		if err := domainusage.ValidateSeriesPointLimit(len(points)); err != nil {
			return types.UsageSeriesResponse{}, err
		}
		return types.UsageSeriesResponse{
			Granularity: q.Granularity,
			Source:      domainusage.SourceBuckets,
			Timezone:    q.Timezone,
			Approximate: false,
			MappingAsOf: domainusage.MappingAsOfIngestTime,
			Points:      points,
		}, nil
	default:
		return types.UsageSeriesResponse{}, domainusage.ValidateGroupBy("invalid")
	}
}

func (s *service) CostSummary(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) (types.CostSummary, error) {
	current, err := s.resolveRange(params)
	if err != nil {
		return types.CostSummary{}, err
	}
	prev := periodutil.PreviousRange(current)
	scopeDeptIDs, err := s.resolveScope(ctx, scope, "")
	if err != nil {
		return types.CostSummary{}, err
	}
	base := types.UsageAggregateQuery{Timezone: current.Timezone, ScopeDeptIDs: scopeDeptIDs}
	currentTotals, err := s.store.Usage().QuerySummary(ctx, withRange(base, current))
	if err != nil {
		return types.CostSummary{}, err
	}
	prevTotals, err := s.store.Usage().QuerySummary(ctx, withRange(base, prev))
	if err != nil {
		return types.CostSummary{}, err
	}
	memberCount := float64(len(s.store.Org().Members()))
	if memberCount == 0 {
		memberCount = 1
	}
	avgCostPerRequest := safeDiv(currentTotals.CostCNY, float64(currentTotals.CallCount))
	prevAvgCostPerRequest := safeDiv(prevTotals.CostCNY, float64(prevTotals.CallCount))
	avgCostPerMember := currentTotals.CostCNY / memberCount
	prevAvgCostPerMember := prevTotals.CostCNY / memberCount
	return types.CostSummary{
		TotalCost:            currentTotals.CostCNY,
		TotalCostMom:         mom(currentTotals.CostCNY, prevTotals.CostCNY),
		TotalTokens:          0,
		TotalRequests:        float64(currentTotals.CallCount),
		TotalRequestsMom:     mom(float64(currentTotals.CallCount), float64(prevTotals.CallCount)),
		AvgCostPerRequest:    avgCostPerRequest,
		AvgCostPerRequestMom: mom(avgCostPerRequest, prevAvgCostPerRequest),
		AvgCostPerMember:     avgCostPerMember,
		AvgCostPerMemberMom:  mom(avgCostPerMember, prevAvgCostPerMember),
	}, nil
}

func (s *service) DepartmentCosts(ctx context.Context, parentID string, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCost, error) {
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, "")
	if err != nil {
		return nil, err
	}
	departments := s.store.Org().Departments()
	childIDs := storeChildDepartmentIDs(departments, parentID)
	if len(childIDs) == 0 {
		return []types.DepartmentCost{}, nil
	}
	rows, err := s.store.Usage().QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: domainusage.GroupByDepartment, DepartmentIDs: childIDs, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, row := range rows {
		total += row.CostCNY
	}
	result := make([]types.DepartmentCost, 0, len(rows))
	for _, row := range rows {
		dept := orgutil.FindDepartment(departments, row.DepartmentID)
		name := row.DepartmentID
		hasChildren := false
		if dept != nil {
			name = dept.Name
			hasChildren = len(dept.Children) > 0
		}
		pct := 0.0
		if total > 0 {
			pct = row.CostCNY / total * 100
		}
		result = append(result, types.DepartmentCost{
			DepartmentID: row.DepartmentID, DepartmentName: name,
			Cost: row.CostCNY, Percentage: pct, HasChildren: hasChildren,
		})
	}
	return result, nil
}

func (s *service) DepartmentMemberCosts(ctx context.Context, deptID string, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCostMember, error) {
	if !domainusage.IsDepartmentAccessible(s.store.Org().Departments(), scope, deptID) {
		return nil, domain.NewDomainError(domain.StatusForbidden, "Department not accessible")
	}
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, deptID)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.Usage().QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Timezone: rng.Timezone,
		GroupBy: domainusage.GroupByMember, DepartmentID: deptID, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	members := s.store.Org().Members()
	result := make([]types.DepartmentCostMember, 0, len(rows))
	for _, row := range rows {
		name := row.MemberID
		if member, ok := queryutil.FindMemberByID(members, row.MemberID); ok {
			name = member.Name
		}
		result = append(result, types.DepartmentCostMember{
			MemberID: row.MemberID, MemberName: name,
			Cost: row.CostCNY, Requests: float64(row.CallCount), Tokens: 0,
		})
	}
	return result, nil
}

func (s *service) DailyCosts(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DailyCost, error) {
	rng, err := s.resolveRange(params)
	if err != nil {
		return nil, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, "")
	if err != nil {
		return nil, err
	}
	rows, err := s.store.Usage().QueryAggregates(ctx, types.UsageAggregateQuery{
		Start: rng.Start, End: rng.End, Granularity: rng.Granularity, Timezone: rng.Timezone,
		GroupBy: domainusage.GroupByNone, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]types.DailyCost, 0, len(rows))
	for _, row := range rows {
		result = append(result, types.DailyCost{
			Date: row.Bucket, Cost: row.CostCNY, Requests: float64(row.CallCount), Tokens: 0,
		})
	}
	return result, nil
}

func (s *service) TopConsumers(ctx context.Context, limit int, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.TopConsumer, error) {
	if limit <= 0 {
		limit = 5
	}
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
		GroupBy: domainusage.GroupByMember, Limit: limit, ScopeDeptIDs: scopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	result := make([]types.TopConsumer, 0, len(rows))
	for _, row := range rows {
		name := row.MemberID
		deptName := ""
		if member, ok := queryutil.FindMemberByID(members, row.MemberID); ok {
			name = member.Name
			if path := orgutil.GetDeptPath(departments, member.DepartmentID); path != nil {
				deptName = *path
			}
		}
		result = append(result, types.TopConsumer{
			MemberID: row.MemberID, MemberName: name, Department: deptName,
			Cost: row.CostCNY, Requests: float64(row.CallCount), Tokens: 0,
		})
	}
	return result, nil
}

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
		GroupBy: domainusage.GroupByModel, ScopeDeptIDs: scopeDeptIDs,
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
		GroupBy: domainusage.GroupByDepartment, ScopeDeptIDs: scopeDeptIDs,
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
			GroupBy: domainusage.GroupByModel, DepartmentID: dept.ID, Limit: 1,
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

func (s *service) resolveRange(params types.CostQueryParams) (periodutil.ResolvedRange, error) {
	if err := domainusage.ValidateCostGranularity(params.Granularity); err != nil {
		return periodutil.ResolvedRange{}, err
	}
	normalized := params
	normalized.Granularity = domainusage.NormalizeCostGranularity(params.Granularity)
	return periodutil.Resolve(normalized, s.now(), domainusage.ResolveTimezone(""))
}

func (s *service) resolveScope(_ context.Context, scope domainusage.SessionScope, requestedDeptID string) ([]string, error) {
	return domainusage.ResolveScopeDepartments(s.store.Org().Departments(), scope, requestedDeptID)
}

func withRange(base types.UsageAggregateQuery, rng periodutil.ResolvedRange) types.UsageAggregateQuery {
	base.Start = rng.Start
	base.End = rng.End
	base.Timezone = rng.Timezone
	base.Granularity = rng.Granularity
	return base
}

func storeChildDepartmentIDs(departments []types.Department, parentID string) []string {
	if parentID == "" {
		ids := make([]string, 0, len(departments))
		for _, dept := range departments {
			ids = append(ids, dept.ID)
		}
		return ids
	}
	parent := orgutil.FindDepartment(departments, parentID)
	if parent == nil {
		return nil
	}
	ids := make([]string, 0, len(parent.Children))
	for _, child := range parent.Children {
		ids = append(ids, child.ID)
	}
	return ids
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

func mom(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return (current - previous) / previous * 100
}

func safeDiv(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
