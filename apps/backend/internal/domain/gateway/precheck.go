package gateway

import (
	"context"
	"fmt"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/store"
)

type Prechecker interface {
	Run(ctx context.Context, keyHash string, model string, opts PrecheckOpts) error
}

// PrecheckOpts controls optional gateway precheck skips.
type PrecheckOpts struct {
	SkipModelCheck     bool // /v1/models listing
	SkipModelAllowlist bool // modelcatalog.IsLocalOnlyCallType when allowDev
}

// PrecheckForRequest builds precheck options from the gateway request context.
func PrecheckForRequest(path, model string, allowDev bool) PrecheckOpts {
	return PrecheckOpts{
		SkipModelCheck:     path == "/v1/models",
		SkipModelAllowlist: allowDev && modelcatalog.IsLocalOnlyCallType(model),
	}
}

type PrecheckService struct {
	loader      store.GatewayPrecheckRepository
	clock       clock.Clock
	budgetCheck domainbudget.CombinedKeyCache
}

func NewPrecheckService(loader store.GatewayPrecheckRepository, clk clock.Clock, budgetCheck domainbudget.CombinedKeyCache) *PrecheckService {
	if budgetCheck == nil {
		budgetCheck = domainbudget.NoopCombinedKeyCache
	}
	return &PrecheckService{
		loader:      loader,
		clock:       clock.OrDefault(clk),
		budgetCheck: budgetCheck,
	}
}

func (p *PrecheckService) Run(ctx context.Context, keyHash string, model string, opts PrecheckOpts) error {
	row, err := p.loader.LoadPrecheckContext(ctx, keyHash)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("platform key not found")
	}
	if err := EvaluateAt(PrecheckContextFromStore(row), model, opts, p.clock.Now()); err != nil {
		return err
	}
	return p.budgetRemainCheck(ctx, row.CompanyID, keyHash, row.CombinedKeyRemainVersion)
}

func (p *PrecheckService) budgetRemainCheck(ctx context.Context, companyID int64, keyHash string, pgVersion int64) error {
	if !p.budgetCheck.Enabled() {
		return nil
	}
	entry, ok, err := p.budgetCheck.Get(ctx, companyID, keyHash)
	if err != nil || !ok {
		return nil
	}
	if domainbudget.BlocksCombinedKey(entry, pgVersion) {
		return ErrBudgetExhausted
	}
	return nil
}

var _ Prechecker = (*PrecheckService)(nil)
