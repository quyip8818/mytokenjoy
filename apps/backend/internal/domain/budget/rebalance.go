package budget

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

type Rebalancer interface {
	ProcessAxis(ctx context.Context, axisKind, axisID string) error
}

// RebalanceStore is the narrow store surface the rebalance processor needs.
type RebalanceStore interface {
	BudgetConsumed() store.BudgetConsumedRepository
	Org() store.OrgRepository
	Budget() store.BudgetRepository
	Keys() store.KeysRepository
	PlatformKeyMappings() store.PlatformKeyMappingRepository
	Company() store.CompanyRepository
	Models() store.ModelsRepository
	CombinedKeySummaries() store.CombinedKeySummaryRepository
}

type RebalanceService struct {
	cfg    config.Config
	store  RebalanceStore
	client adminport.Port
}

func NewRebalanceService(cfg config.Config, st RebalanceStore, client adminport.Port) *RebalanceService {
	return &RebalanceService{cfg: cfg, store: st, client: client}
}

// rebalanceContext holds preloaded data shared across all mappings in a single ProcessAxis call.
type rebalanceContext struct {
	budgetCtx   pkgbudget.BudgetContext
	departments []types.Department
	rules       []types.RoutingRule
	models      []types.ModelInfo
}

func (s *RebalanceService) ProcessAxis(ctx context.Context, axisKind, axisID string) error {
	if s.client == nil {
		return fmt.Errorf("newapi admin client required")
	}
	var mappings []store.PlatformKeyMapping
	var err error
	switch axisKind {
	case store.RebalanceAxisMember:
		mappings, err = s.store.PlatformKeyMappings().ListMappingsByMemberID(ctx, axisID)
	case store.RebalanceAxisProject:
		mappings, err = s.store.PlatformKeyMappings().ListMappingsByProjectID(ctx, axisID)
	case store.RebalanceAxisCompany:
		companyID, parseErr := strconv.ParseInt(axisID, 10, 64)
		if parseErr != nil {
			return parseErr
		}
		mappings, err = s.store.PlatformKeyMappings().ListActiveMappingsByCompany(ctx, companyID)
	default:
		return nil
	}
	if err != nil {
		return err
	}

	// Filter to actionable mappings first.
	active := mappings[:0]
	for _, m := range mappings {
		if m.NewAPIKeyID != nil && m.SyncStatus == store.MappingSyncStatusSynced {
			active = append(active, m)
		}
	}
	if len(active) == 0 {
		return nil
	}

	// Preload shared data once for all mappings.
	rctx, err := s.loadRebalanceContext(ctx)
	if err != nil {
		return err
	}

	for _, mapping := range active {
		if err := s.rebalanceKey(ctx, mapping, rctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *RebalanceService) loadRebalanceContext(ctx context.Context) (*rebalanceContext, error) {
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return nil, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return nil, err
	}
	return &rebalanceContext{
		budgetCtx:   budgetCtx,
		departments: departments,
		rules:       rules,
		models:      models,
	}, nil
}

func (s *RebalanceService) rebalanceKey(ctx context.Context, mapping store.PlatformKeyMapping, rctx *rebalanceContext) error {
	key, ok := rctx.budgetCtx.FindPlatformKey(mapping.PlatformKeyID)
	if !ok || key.Status != "active" {
		return nil
	}
	token, err := s.client.GetToken(ctx, *mapping.NewAPIKeyID)
	if err != nil {
		return err
	}

	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, rctx.departments, rctx.rules, rctx.models)
	effectiveIDs := newapiunits.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, s.store.Org().Nodes(), mapping.DepartmentID, s.cfg.Clock())
	if err != nil {
		return err
	}
	remainPoint, err := pkgbudget.ComputeRemainForMapping(
		ctx, rctx.budgetCtx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Company(), mapping, open.String(),
	)
	if err != nil {
		return err
	}
	allocated := newapiunits.ToNewAPIUnits(
		remainPoint,
		rctx.models,
		effectiveIDs,
	)
	if allocated == token.RemainQuota {
		return nil
	}
	remain := allocated
	req := adminport.UpdateTokenInput{
		ID:          token.ID,
		RemainQuota: &remain,
	}
	if _, err := s.client.UpdateToken(ctx, req); err != nil {
		return err
	}
	return RefreshPlatformKeyCombined(ctx, s.store, mapping.PlatformKeyID, s.cfg.Clock(), nil)
}

var _ Rebalancer = (*RebalanceService)(nil)
