package budget

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

type Rebalancer interface {
	ProcessAxis(ctx context.Context, axisKind, axisID string) error
}

type RebalanceService struct {
	cfg    config.Config
	store  store.Store
	client adminport.Port
}

func NewRebalanceService(cfg config.Config, st store.Store, client adminport.Port) *RebalanceService {
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
	newRemain := allocated
	walletID, err := s.newAPIWalletUserID(ctx, mapping.CompanyID)
	if err != nil {
		return err
	}
	if walletID > 0 {
		walletUnits, walletErr := s.walletAvailable(ctx, mapping, allocated)
		if walletErr != nil {
			return walletErr
		}
		if walletUnits < newRemain {
			newRemain = walletUnits
		}
	}
	if newRemain == token.RemainQuota {
		return nil
	}
	remain := newRemain
	req := adminport.UpdateTokenInput{
		ID:          token.ID,
		RemainQuota: &remain,
	}
	updated, err := s.client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	if err := s.store.PlatformKeyMappings().UpdateMappingNewAPIKeyRemainQuota(ctx, mapping.PlatformKeyID, updated.RemainQuota); err != nil {
		return err
	}
	return RefreshPlatformKeySoft(ctx, s.store, mapping.PlatformKeyID, s.cfg.Clock(), nil)
}

func (s *RebalanceService) newAPIWalletUserID(ctx context.Context, companyID int64) (int64, error) {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletUserID > 0 {
		return companyCtx.NewAPIWalletUserID, nil
	}
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return 0, err
	}
	if co == nil {
		return 0, nil
	}
	id, ok := store.ConfiguredNewAPIWalletUserID(co)
	if !ok {
		return 0, nil
	}
	return id, nil
}

func (s *RebalanceService) walletAvailable(ctx context.Context, mapping store.PlatformKeyMapping, allocated int64) (int64, error) {
	co, err := s.store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		return 0, err
	}
	if co == nil {
		return 0, fmt.Errorf("company not found: %d", mapping.CompanyID)
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return 0, err
	}
	walletUnits := newapiunits.ToNewAPIUnits(co.WalletRemain, models, nil)
	mappings, err := s.store.PlatformKeyMappings().ListActiveMappingsByCompany(ctx, mapping.CompanyID)
	if err != nil {
		return 0, err
	}
	var used int64
	for _, m := range mappings {
		if m.PlatformKeyID == mapping.PlatformKeyID || m.NewAPIKeyRemainQuota == nil {
			continue
		}
		used = newapiunits.AddSat(used, *m.NewAPIKeyRemainQuota)
	}
	available := newapiunits.SubFloor0(walletUnits, used)
	if allocated < available {
		return allocated, nil
	}
	return available, nil
}

var _ Rebalancer = (*RebalanceService)(nil)
