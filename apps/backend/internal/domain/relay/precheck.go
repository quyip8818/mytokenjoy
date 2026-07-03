package relay

import (
	"context"
	"fmt"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

const minEstimateCNY = 0.01

type PrecheckInput struct {
	Mapping *store.RelayMapping
	Company *store.Company
	Model   string
}

type Prechecker interface {
	Run(ctx context.Context, in PrecheckInput) error
}

type PrecheckService struct {
	store  store.Store
	wallet domaincompany.WalletService
}

func NewPrecheckService(st store.Store, wallet domaincompany.WalletService) *PrecheckService {
	return &PrecheckService{store: st, wallet: wallet}
}

func (p *PrecheckService) Run(ctx context.Context, in PrecheckInput) error {
	if domaincompany.IsRelayBlocked(in.Company.Status) {
		return fmt.Errorf("company not active")
	}
	if err := p.checkWallet(ctx, in.Company); err != nil {
		return err
	}
	if err := p.checkDepartmentBudget(ctx, in.Mapping); err != nil {
		return err
	}
	if err := p.checkTokenRemainQuota(in.Mapping); err != nil {
		return err
	}
	return p.checkPlatformKey(ctx, in.Mapping, in.Model)
}

func (p *PrecheckService) checkWallet(ctx context.Context, company *store.Company) error {
	if company.NewAPIWalletUserID == nil || p.wallet == nil {
		return nil
	}
	quota, err := p.wallet.AvailableQuota(ctx, *company.NewAPIWalletUserID)
	if err != nil {
		return fmt.Errorf("wallet unavailable")
	}
	balanceCNY := newapi.FromNewAPIUnits(quota, nil, nil)
	if balanceCNY < minEstimateCNY {
		return fmt.Errorf("insufficient wallet balance")
	}
	return nil
}

func (p *PrecheckService) checkDepartmentBudget(ctx context.Context, mapping *store.RelayMapping) error {
	tree, err := common.LoadBudgetTree(ctx, p.store)
	if err != nil {
		return err
	}
	node := pkgbudget.FindBudgetNode(tree, mapping.DepartmentID)
	if node == nil {
		return fmt.Errorf("department not found")
	}
	if node.Budget <= 0 {
		return fmt.Errorf("budget exceeded")
	}
	if node.Consumed+minEstimateCNY > node.Budget {
		return fmt.Errorf("budget exceeded")
	}
	return nil
}

func (p *PrecheckService) checkTokenRemainQuota(mapping *store.RelayMapping) error {
	if mapping.NewAPITokenRemainQuota == nil || *mapping.NewAPITokenRemainQuota <= 0 {
		return fmt.Errorf("insufficient token quota")
	}
	return nil
}

func (p *PrecheckService) checkPlatformKey(ctx context.Context, mapping *store.RelayMapping, modelName string) error {
	keys, err := p.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	var key *types.PlatformKey
	for i := range keys {
		if keys[i].ID == mapping.PlatformKeyID {
			key = &keys[i]
			break
		}
	}
	if key == nil {
		return fmt.Errorf("platform key not found")
	}
	if key.Status != "active" {
		return fmt.Errorf("platform key inactive")
	}
	if modelName == "" || len(key.ModelWhitelist) == 0 {
		return nil
	}
	for _, allowed := range key.ModelWhitelist {
		if allowed == modelName {
			return nil
		}
	}
	return fmt.Errorf("model not allowed")
}
