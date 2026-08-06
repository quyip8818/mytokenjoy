package gateway

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/clock"
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
	SkipModelAllowlist bool // test model (source=="test") — not in allowlist, skip check
	SkipWalletCheck    bool // selfhosted with non-platform channels — wallet not required
}

// BuildPrecheckOpts builds precheck options from request context.
// isTestModel should be resolved by the caller via catalog source lookup.
func BuildPrecheckOpts(path string, isTestModel bool) PrecheckOpts {
	return PrecheckOpts{
		SkipModelCheck:     path == "/v1/models",
		SkipModelAllowlist: isTestModel,
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
	// Selfhosted companies may have non-platform channels; skip wallet check
	// so requests can route to self-managed channels even when wallet=0.
	// SaaS Gateway is the backstop for platform-channel traffic.
	if row.CompanyType == store.CompanyTypeSelfhosted {
		opts.SkipWalletCheck = true
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
