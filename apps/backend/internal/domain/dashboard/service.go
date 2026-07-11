package dashboard

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	UsageSeries(ctx context.Context, q types.UsageSeriesQuery, scope domainusage.SessionScope) (types.UsageSeriesResponse, error)
	UsageSeriesFromQuery(ctx context.Context, rawGranularity, rawStart, rawEnd, groupBy, deptID, memberID string, scope domainusage.SessionScope) (types.UsageSeriesResponse, error)
	CostSummary(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) (types.CostSummary, error)
	DepartmentCosts(ctx context.Context, parentID string, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCost, error)
	DepartmentMemberCosts(ctx context.Context, deptID string, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DepartmentCostMember, error)
	DailyCosts(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.DailyCost, error)
	TopConsumers(ctx context.Context, limit int, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.TopConsumer, error)
	ModelUsage(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.ModelUsage, error)
	TeamUsage(ctx context.Context, params types.CostQueryParams, scope domainusage.SessionScope) ([]types.TeamUsage, error)
}

type service struct {
	cfg    config.Config
	store  store.Store
	reader domainusage.Reader
	clock  clock.Clock
}

func NewService(cfg config.Config, st store.Store, reader domainusage.Reader) Service {
	return &service{
		cfg:    cfg,
		store:  st,
		reader: reader,
		clock:  cfg.Clock(),
	}
}

func (s *service) resolveRange(params types.CostQueryParams) (budget.ResolvedRange, error) {
	if err := domainusage.ValidateCostGranularity(params.Granularity); err != nil {
		return budget.ResolvedRange{}, err
	}
	normalized := params
	normalized.Granularity = domainusage.NormalizeCostGranularity(params.Granularity)
	return budget.Resolve(normalized, s.clock.Now(), domainusage.ResolveTimezone(""))
}

func (s *service) resolveScope(ctx context.Context, scope domainusage.SessionScope, requestedDeptID string) ([]string, error) {
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	return domainusage.ResolveScopeDepartments(departments, scope, requestedDeptID)
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
