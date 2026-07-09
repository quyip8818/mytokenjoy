package budget

import (
	"context"
	"strconv"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Rebalancer interface {
	ProcessAxis(ctx context.Context, axisKind, axisID string) error
}

type RebalanceService struct {
	cfg    config.Config
	store  store.Store
	client newapi.AdminClient
}

func NewRebalanceService(cfg config.Config, st store.Store, client newapi.AdminClient) *RebalanceService {
	return &RebalanceService{cfg: cfg, store: st, client: client}
}

func (s *RebalanceService) ProcessAxis(ctx context.Context, axisKind, axisID string) error {
	if s.client == nil || !s.cfg.NewAPIEnabled {
		return nil
	}
	var mappings []store.RelayMapping
	var err error
	switch axisKind {
	case store.RebalanceAxisMember:
		mappings, err = s.store.Relay().ListMappingsByMemberID(ctx, axisID)
	case store.RebalanceAxisDepartment:
		mappings, err = s.store.Relay().ListMappingsByDepartmentID(ctx, axisID)
	case store.RebalanceAxisBudgetGroup:
		mappings, err = s.store.Relay().ListMappingsByBudgetGroupID(ctx, axisID)
	case store.RebalanceAxisCompany:
		companyID, parseErr := strconv.ParseInt(axisID, 10, 64)
		if parseErr != nil {
			return parseErr
		}
		mappings, err = s.store.Relay().ListActiveMappingsByCompany(ctx, companyID)
	default:
		return nil
	}
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if mapping.NewAPITokenID == nil || mapping.SyncStatus != store.RelaySyncStatusSynced {
			continue
		}
		if err := s.rebalanceKey(ctx, mapping); err != nil {
			return err
		}
	}
	return nil
}

func (s *RebalanceService) rebalanceKey(ctx context.Context, mapping store.RelayMapping) error {
	platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.NowUTC())
	if err != nil {
		return err
	}
	key, ok := findPlatformKeyByID(platformKeys, mapping.PlatformKeyID)
	if !ok || key.Status != "active" {
		return nil
	}
	token, err := s.client.GetToken(ctx, *mapping.NewAPITokenID)
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
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	groups, err := pkgbudget.LoadBudgetGroupsWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.cfg.NowUTC())
	if err != nil {
		return err
	}
	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes(), s.cfg.NowUTC())
	if err != nil {
		return err
	}

	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, departments, rules, models)
	effectiveIDs := newapi.EffectiveWhitelistIDs(key.ModelWhitelist, deptAllowed)
	allocated := newapi.ToNewAPIUnits(
		relay.ComputeRemainQuota(key, tree, members, platformKeys, groups, mapping.DepartmentID),
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
	req := newapi.UpdateTokenRequest{
		ID:          token.ID,
		RemainQuota: &remain,
	}
	updated, err := s.client.UpdateToken(ctx, req)
	if err != nil {
		return err
	}
	return s.store.Relay().UpdateMappingNewAPITokenRemainQuota(ctx, mapping.PlatformKeyID, updated.RemainQuota)
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

func (s *RebalanceService) walletAvailable(ctx context.Context, mapping store.RelayMapping, allocated int64) (int64, error) {
	co, err := s.store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil || co == nil {
		return allocated, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return allocated, err
	}
	walletUnits := newapi.ToQuotaUnits(co.BalancePoint, models, nil)
	mappings, err := s.store.Relay().ListActiveMappingsByCompany(ctx, mapping.CompanyID)
	if err != nil {
		return allocated, err
	}
	var used int64
	for _, m := range mappings {
		if m.PlatformKeyID == mapping.PlatformKeyID || m.NewAPITokenRemainQuota == nil {
			continue
		}
		used += *m.NewAPITokenRemainQuota
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

func findPlatformKeyByID(platformKeys []types.PlatformKey, id string) (types.PlatformKey, bool) {
	for _, key := range platformKeys {
		if key.ID == id {
			return key, true
		}
	}
	return types.PlatformKey{}, false
}

var _ Rebalancer = (*RebalanceService)(nil)
