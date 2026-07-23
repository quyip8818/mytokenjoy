package gateway

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/store"
)

// PrecheckResult carries metadata from a successful precheck (e.g. company type for routing decisions).
type PrecheckResult struct {
	CompanyType string
}

type Prechecker interface {
	Run(ctx context.Context, keyHash string, model string, opts PrecheckOpts) (PrecheckResult, error)
}

// PrecheckOpts controls optional gateway precheck skips.
type PrecheckOpts struct {
	SkipModelCheck     bool // /v1/models listing
	SkipModelAllowlist bool // modelcatalog.IsTestOnlyCallType when allowDev
}

// PrecheckForRequest builds precheck options from the gateway request context.
func PrecheckForRequest(path, model string) PrecheckOpts {
	return PrecheckOpts{
		SkipModelCheck:     path == "/v1/models",
		SkipModelAllowlist: modelcatalog.IsTestOnlyCallType(model),
	}
}

type PrecheckService struct {
	cache       *PrecheckCache
	clock       clock.Clock
	budgetCheck domainbudget.CombinedKeyCache
}

// NewPrecheckService creates a precheck service.
// Use NewPrecheckCache to create the cache from a GatewayPrecheckRepository.
func NewPrecheckService(cache *PrecheckCache, clk clock.Clock, budgetCheck domainbudget.CombinedKeyCache) *PrecheckService {
	if budgetCheck == nil {
		budgetCheck = domainbudget.NoopCombinedKeyCache
	}
	return &PrecheckService{
		cache:       cache,
		clock:       clock.OrDefault(clk),
		budgetCheck: budgetCheck,
	}
}

// NewPrecheckServiceLegacy creates a precheck service with a raw loader (no cache).
// Used in tests that don't need caching.
func NewPrecheckServiceLegacy(loader store.GatewayPrecheckRepository, clk clock.Clock, budgetCheck domainbudget.CombinedKeyCache) *PrecheckService {
	return NewPrecheckService(NewPrecheckCache(loader), clk, budgetCheck)
}

func (p *PrecheckService) Run(ctx context.Context, keyHash string, model string, opts PrecheckOpts) (PrecheckResult, error) {
	row, err := p.cache.Get(ctx, keyHash)
	if err != nil {
		return PrecheckResult{}, err
	}
	if row == nil {
		return PrecheckResult{}, fmt.Errorf("platform key not found")
	}
	if err := EvaluateAt(PrecheckContextFromStore(row), model, opts, p.clock.Now()); err != nil {
		return PrecheckResult{}, err
	}
	if err := p.budgetRemainCheck(ctx, row.CompanyID, keyHash); err != nil {
		return PrecheckResult{}, err
	}
	return PrecheckResult{CompanyType: row.CompanyType}, nil
}

// budgetRemainCheck queries Redis directly for the remain value.
// No PG version comparison — Ingest SET always overwrites Redis with the precise value,
// and Rebalance refreshes after budget changes. Fail-open on cache miss or Redis error.
func (p *PrecheckService) budgetRemainCheck(ctx context.Context, companyID uuid.UUID, keyHash string) error {
	if !p.budgetCheck.Enabled() {
		return nil
	}
	entry, ok, err := p.budgetCheck.Get(ctx, companyID, keyHash)
	if err != nil || !ok {
		return nil // fail-open
	}
	if entry.Remain <= 0 {
		return ErrBudgetExhausted
	}
	return nil
}

// Cache returns the underlying PrecheckCache for invalidation by other services.
func (p *PrecheckService) Cache() *PrecheckCache {
	return p.cache
}

var _ Prechecker = (*PrecheckService)(nil)
