package gateway

import (
	"context"
	"fmt"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

type Prechecker interface {
	Run(ctx context.Context, keyHash string, model string, skipModelCheck bool) error
}

type PrecheckService struct {
	loader      store.GatewayPrecheckRepository
	clock       clock.Clock
	budgetCheck domainbudget.GatewaySoftCache
}

func NewPrecheckService(loader store.GatewayPrecheckRepository, clk clock.Clock, budgetCheck domainbudget.GatewaySoftCache) *PrecheckService {
	if budgetCheck == nil {
		budgetCheck = domainbudget.NoopGatewaySoftCache
	}
	return &PrecheckService{
		loader:      loader,
		clock:       clock.OrDefault(clk),
		budgetCheck: budgetCheck,
	}
}

func (p *PrecheckService) Run(ctx context.Context, keyHash string, model string, skipModelCheck bool) error {
	row, err := p.loader.LoadPrecheckContext(ctx, keyHash)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("platform key not found")
	}
	if err := Evaluate(PrecheckContextFromStore(row), model, skipModelCheck); err != nil {
		return err
	}
	return p.softBudgetCheck(ctx, row.CompanyID, keyHash, row.GatewaySoftVersion)
}

func (p *PrecheckService) softBudgetCheck(ctx context.Context, companyID int64, keyHash string, pgVersion int64) error {
	if !p.budgetCheck.Enabled() {
		return nil
	}
	entry, ok, err := p.budgetCheck.Get(ctx, companyID, keyHash)
	if err != nil || !ok {
		return nil
	}
	if domainbudget.BlocksGatewaySoft(entry, pgVersion) {
		return ErrBudgetExhausted
	}
	return nil
}

var _ Prechecker = (*PrecheckService)(nil)
