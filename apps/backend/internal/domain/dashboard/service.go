package dashboard

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/budget"
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
	cfg           config.Config
	store         store.Store
	logAggregator *domainusage.LogAggregator
	now           func() time.Time
}

func NewService(cfg config.Config, st store.Store, logAggregator *domainusage.LogAggregator) Service {
	return &service{
		cfg:           cfg,
		store:         st,
		logAggregator: logAggregator,
		now:           time.Now,
	}
}

func (s *service) resolveRange(params types.CostQueryParams) (budget.ResolvedRange, error) {
	if err := domainusage.ValidateCostGranularity(params.Granularity); err != nil {
		return budget.ResolvedRange{}, err
	}
	normalized := params
	normalized.Granularity = domainusage.NormalizeCostGranularity(params.Granularity)
	return budget.Resolve(normalized, s.dashboardNow(), domainusage.ResolveTimezone(""))
}

func (s *service) dashboardNow() time.Time {
	if !s.cfg.IsDemoProfile() {
		return s.now()
	}
	t, err := time.Parse("2006-01-02", s.cfg.DemoToday)
	if err != nil {
		return s.now()
	}
	return t
}

func (s *service) resolveScope(_ context.Context, scope domainusage.SessionScope, requestedDeptID string) ([]string, error) {
	return domainusage.ResolveScopeDepartments(s.store.Org().Departments(), scope, requestedDeptID)
}

func withRange(base types.UsageAggregateQuery, rng budget.ResolvedRange) types.UsageAggregateQuery {
	base.Start = rng.Start
	base.End = rng.End
	base.Timezone = rng.Timezone
	base.Granularity = rng.Granularity
	return base
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
