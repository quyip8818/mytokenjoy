//go:build testhook

package gatewayfix

import (
	"context"
	"sync"

	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
)

// FakeBudgetCheck is an in-memory budgetcheck.Store for soft-block tests.
type FakeBudgetCheck struct {
	mu      sync.Mutex
	entries map[string]budgetcheck.Entry
	gets    int
	enabled bool
}

func NewFakeBudgetCheck() *FakeBudgetCheck {
	return &FakeBudgetCheck{entries: map[string]budgetcheck.Entry{}, enabled: true}
}

func (f *FakeBudgetCheck) Enabled() bool { return f.enabled }

func (f *FakeBudgetCheck) Get(_ context.Context, companyID int64, keyHash string) (budgetcheck.Entry, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gets++
	entry, ok := f.entries[budgetcheck.Key(companyID, keyHash)]
	return entry, ok, nil
}

func (f *FakeBudgetCheck) Set(_ context.Context, companyID int64, keyHash string, entry budgetcheck.Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[budgetcheck.Key(companyID, keyHash)] = entry
	return nil
}

// Gets returns the number of Get calls (for asserting the +1 Redis GET).
func (f *FakeBudgetCheck) Gets() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.gets
}

var _ budgetcheck.Store = (*FakeBudgetCheck)(nil)
