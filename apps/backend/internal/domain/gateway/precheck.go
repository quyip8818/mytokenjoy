package gateway

import (
	"context"
	"fmt"
	"math"

	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

const minEstimatePoint = 0.01 * float64(common.DefaultPointsPerUnit)

type PrecheckInput struct {
	Mapping        *store.PlatformKeyMapping
	Company        *store.Company
	Model          string
	SkipModelCheck bool
}

type Prechecker interface {
	Run(ctx context.Context, in PrecheckInput) error
}

type PrecheckService struct {
	snapshots  store.BudgetSnapshotRepository
	orgNodes   store.OrgNodeRepository
	org        store.OrgRepository
	budget     store.BudgetRepository
	keys       store.KeysRepository
	models     store.ModelsRepository
	wallet     domaincompany.WalletService
	walletSync store.WalletSyncQueueRepository
	clock      clock.Clock
}

func NewPrecheckService(
	snapshots store.BudgetSnapshotRepository,
	orgNodes store.OrgNodeRepository,
	budget store.BudgetRepository,
	org store.OrgRepository,
	keys store.KeysRepository,
	models store.ModelsRepository,
	wallet domaincompany.WalletService,
	walletSync store.WalletSyncQueueRepository,
	clk clock.Clock,
) *PrecheckService {
	return &PrecheckService{
		snapshots:  snapshots,
		orgNodes:   orgNodes,
		budget:     budget,
		org:        org,
		keys:       keys,
		models:     models,
		wallet:     wallet,
		walletSync: walletSync,
		clock:      clock.OrDefault(clk),
	}
}

func (p *PrecheckService) Run(ctx context.Context, in PrecheckInput) error {
	if !in.SkipModelCheck && in.Model == "" {
		return fmt.Errorf("model field is required")
	}
	if domaincompany.IsGatewayBlocked(in.Company.Status) {
		return fmt.Errorf("company not active")
	}
	if err := p.checkBalancePoint(in.Company); err != nil {
		return err
	}
	if err := p.checkBudgetRemain(ctx, in.Mapping); err != nil {
		return err
	}
	if err := p.checkNewAPIKeyRemainQuota(in.Mapping); err != nil {
		return err
	}
	if err := p.checkNewAPIWalletCap(ctx, in.Company); err != nil {
		return err
	}
	if err := p.checkWalletSyncLag(ctx, in.Company); err != nil {
		return err
	}
	if in.SkipModelCheck {
		return nil
	}
	return p.checkPlatformKey(ctx, in.Mapping, in.Model)
}

func (p *PrecheckService) checkBalancePoint(company *store.Company) error {
	if company.BalancePoint < minEstimatePoint {
		return fmt.Errorf("insufficient wallet balance")
	}
	return nil
}

func (p *PrecheckService) checkBudgetRemain(ctx context.Context, mapping *store.PlatformKeyMapping) error {
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, p.orgNodes, mapping.DepartmentID, p.clock)
	if err != nil {
		return err
	}
	remain, err := pkgbudget.RemainForMapping(ctx, pkgbudget.MappingStores{
		Snapshots: p.snapshots,
		OrgNodes:  p.orgNodes,
		Org:       p.org,
		Budget:    p.budget,
		Keys:      p.keys,
		Clock:     p.clock,
	}, mapping, open.String())
	if err != nil {
		return err
	}
	if remain < minEstimatePoint {
		return fmt.Errorf("budget exceeded")
	}
	return nil
}

func (p *PrecheckService) checkNewAPIWalletCap(ctx context.Context, company *store.Company) error {
	if company.NewAPIWalletUserID == nil || p.wallet == nil {
		return nil
	}
	quota, err := p.wallet.AvailableQuota(ctx, *company.NewAPIWalletUserID)
	if err != nil {
		return fmt.Errorf("wallet unavailable")
	}
	models, err := p.models.Models(ctx)
	if err != nil {
		return err
	}
	balancePoint := newapi.FromNewAPIUnits(quota, models, nil)
	if balancePoint < minEstimatePoint {
		return fmt.Errorf("insufficient wallet balance")
	}
	return nil
}

func (p *PrecheckService) checkWalletSyncLag(ctx context.Context, company *store.Company) error {
	if p.walletSync == nil || company.NewAPIWalletUserID == nil || p.wallet == nil {
		return nil
	}
	pending, err := p.walletSync.HasPendingWalletSync(ctx, company.ID)
	if err != nil {
		return domain.NewDomainErrorWithRetryAfter(
			domain.StatusServiceUnavailable,
			"wallet sync in progress",
			common.WalletSyncRetryAfterSecs,
		)
	}
	if !pending {
		return nil
	}
	quota, err := p.wallet.AvailableQuota(ctx, *company.NewAPIWalletUserID)
	if err != nil {
		return domain.NewDomainErrorWithRetryAfter(
			domain.StatusServiceUnavailable,
			"wallet sync in progress",
			common.WalletSyncRetryAfterSecs,
		)
	}
	models, err := p.models.Models(ctx)
	if err != nil {
		return err
	}
	naPoint := newapi.FromNewAPIUnits(quota, models, nil)
	drift := math.Abs(company.BalancePoint - naPoint)
	if drift > common.WalletSyncDriftEpsilon {
		return domain.NewDomainErrorWithRetryAfter(
			domain.StatusServiceUnavailable,
			"wallet sync in progress",
			common.WalletSyncRetryAfterSecs,
		)
	}
	return nil
}

func (p *PrecheckService) checkNewAPIKeyRemainQuota(mapping *store.PlatformKeyMapping) error {
	if mapping.NewAPIKeyRemainQuota == nil || *mapping.NewAPIKeyRemainQuota <= 0 {
		return fmt.Errorf("insufficient token quota")
	}
	return nil
}

func (p *PrecheckService) checkPlatformKey(ctx context.Context, mapping *store.PlatformKeyMapping, modelName string) error {
	key, err := p.keys.PlatformKeyByID(ctx, mapping.PlatformKeyID)
	if err != nil {
		return err
	}
	if key == nil {
		return fmt.Errorf("platform key not found")
	}
	if key.Status != "active" {
		return fmt.Errorf("platform key inactive")
	}
	if modelName == "" {
		return nil
	}
	hasAny, err := p.models.Allowlist().HasAny(ctx, types.AllowlistOwnerPlatformKey, mapping.PlatformKeyID)
	if err != nil {
		return err
	}
	if !hasAny {
		return nil
	}
	allowed, err := p.models.Allowlist().IsCallTypeAllowed(ctx, types.AllowlistOwnerPlatformKey, mapping.PlatformKeyID, modelName)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("model not allowed")
	}
	return nil
}
