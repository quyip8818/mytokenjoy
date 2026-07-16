package budget

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
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
	for _, mapping := range mappings {
		if mapping.NewAPIKeyID == nil || mapping.SyncStatus != store.MappingSyncStatusSynced {
			continue
		}
		if err := s.rebalanceKey(ctx, mapping); err != nil {
			return err
		}
	}
	return nil
}

func (s *RebalanceService) rebalanceKey(ctx context.Context, mapping store.PlatformKeyMapping) error {
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return err
	}
	key, ok := budgetCtx.FindPlatformKey(mapping.PlatformKeyID)
	if !ok || key.Status != "active" {
		return nil
	}
	token, err := s.client.GetToken(ctx, *mapping.NewAPIKeyID)
	if err != nil {
		return err
	}

	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, departments, rules, models)
	effectiveIDs := newapiunits.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, s.store.Org().Nodes(), mapping.DepartmentID, s.cfg.Clock())
	if err != nil {
		return err
	}
	remainPoint, err := pkgbudget.ComputeRemainForMapping(
		ctx, budgetCtx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Company(), mapping, open.String(),
	)
	if err != nil {
		return err
	}
	allocated := newapiunits.ToNewAPIUnits(
		remainPoint,
		models,
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
