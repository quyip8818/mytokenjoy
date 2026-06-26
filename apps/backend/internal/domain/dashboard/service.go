package dashboard

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/dashboardcalc"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	CostSummary(period string) types.CostSummary
	DepartmentCosts(parentID, period string) []types.DepartmentCost
	DepartmentMemberCosts(deptID, period string) []types.DepartmentCostMember
	DailyCosts(period string) []types.DailyCost
	TopConsumers(limit int, period string) []types.TopConsumer
	ModelUsage() []types.ModelUsage
	TeamUsage() []types.TeamUsage
}

type service struct {
	cfg   config.Config
	store store.Store
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{cfg: cfg, store: st}
}

func (s *service) CostSummary(period string) types.CostSummary {
	resolved := dashboardcalc.ResolvePeriod(period)
	return dashboardcalc.BuildCostSummary(resolved, s.store.Org().Members())
}

func (s *service) DepartmentCosts(parentID, period string) []types.DepartmentCost {
	resolved := dashboardcalc.ResolvePeriod(period)
	return dashboardcalc.GetDepartmentCostsForParent(parentID, resolved)
}

func (s *service) DepartmentMemberCosts(deptID, period string) []types.DepartmentCostMember {
	resolved := dashboardcalc.ResolvePeriod(period)
	return dashboardcalc.GetDepartmentMemberCosts(deptID, resolved)
}

func (s *service) DailyCosts(period string) []types.DailyCost {
	resolved := dashboardcalc.ResolvePeriod(period)
	return dashboardcalc.BuildDailyCosts(resolved)
}

func (s *service) TopConsumers(limit int, period string) []types.TopConsumer {
	resolved := dashboardcalc.ResolvePeriod(period)
	return dashboardcalc.GetTopConsumers(limit, resolved, s.store.Org().Members())
}

func (s *service) ModelUsage() []types.ModelUsage {
	return s.store.Dashboard().ModelUsage()
}

func (s *service) TeamUsage() []types.TeamUsage {
	return s.store.Dashboard().TeamUsage()
}
