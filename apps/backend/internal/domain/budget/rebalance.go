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
	case store.RebalanceAxisDepartment:
		mappings, err = s.store.PlatformKeyMappings().ListMappingsByDepartmentID(ctx, axisID)
	case store.RebalanceAxisBudgetGroup:
		mappings, err = s.store.PlatformKeyMappings().ListMappingsByBudgetGroupID(ctx, axisID)
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
	allocated := newapiunits.ToNewAPIUnits(
		budgetCtx.ComputeRemain(key, mapping.DepartmentID, nil, nil),
		models,
		effectiveIDs,
	)
	newRemain := allocated
	if walletID := s.newAPIWalletUserID(ctx, mapping.CompanyID); walletID > 0 {
		walletUnits, walletErr := s.walletAvailable(ctx, mapping, allocated)
		if walletErr == nil && walletUnits < newRemain {
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
	return s.store.PlatformKeyMappings().UpdateMappingNewAPIKeyRemainQuota(ctx, mapping.PlatformKeyID, updated.RemainQuota)
}

func (s *RebalanceService) newAPIWalletUserID(ctx context.Context, companyID int64) int64 {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletUserID > 0 {
		return companyCtx.NewAPIWalletUserID
	}
	company, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil || company == nil || company.NewAPIWalletUserID == nil {
		return 0
	}
	return *company.NewAPIWalletUserID
}

func (s *RebalanceService) walletAvailable(ctx context.Context, mapping store.PlatformKeyMapping, allocated int64) (int64, error) {
	co, err := s.store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil || co == nil {
		return allocated, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return allocated, err
	}
	walletUnits := newapiunits.ToQuotaUnits(co.WalletRemain, models, nil)
	mappings, err := s.store.PlatformKeyMappings().ListActiveMappingsByCompany(ctx, mapping.CompanyID)
	if err != nil {
		return allocated, err
	}
	var used int64
	for _, m := range mappings {
		if m.PlatformKeyID == mapping.PlatformKeyID || m.NewAPIKeyRemainQuota == nil {
			continue
		}
		used += *m.NewAPIKeyRemainQuota
	}
	available := walletUnits - used
	if available < 0 {
		return 0, nil
	}
	if allocated < available {
		return allocated, nil
	}
	return available, nil
}

var _ Rebalancer = (*RebalanceService)(nil)
