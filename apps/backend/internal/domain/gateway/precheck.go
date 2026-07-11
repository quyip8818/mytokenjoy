package gateway

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

type Prechecker interface {
	Run(ctx context.Context, keyHash string, model string, skipModelCheck bool) error
}

type PrecheckService struct {
	loader      store.GatewayPrecheckRepository
	clock       clock.Clock
	budgetCheck budgetcheck.Store
}

func NewPrecheckService(loader store.GatewayPrecheckRepository, clk clock.Clock, budgetCheck budgetcheck.Store) *PrecheckService {
	if budgetCheck == nil {
		budgetCheck = budgetcheck.Noop{}
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
	return p.softBudgetCheck(ctx, row.CompanyID, keyHash)
}

// softBudgetCheck is a best-effort +1 Redis GET. A miss or any error degrades to
// allow (never falls back to Postgres). It only bridges the sub-second window
// before Overrun disables an over-budget key downstream.
func (p *PrecheckService) softBudgetCheck(ctx context.Context, companyID int64, keyHash string) error {
	if !p.budgetCheck.Enabled() {
		return nil
	}
	entry, ok, err := p.budgetCheck.Get(ctx, companyID, keyHash)
	if err != nil || !ok {
		return nil
	}
	if budgetcheck.Blocks(entry) {
		return fmt.Errorf("budget exhausted")
	}
	return nil
}

var _ Prechecker = (*PrecheckService)(nil)
