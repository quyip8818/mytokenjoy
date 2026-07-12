//go:build testhook

package gatewayfix

import (
	"context"
	"fmt"
	"sync"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
)

func gatewaySoftKey(companyID int64, keyHash string) string {
	return fmt.Sprintf("gateway:budget_check:%d:%s", companyID, keyHash)
}

// FakeBudgetCheck is an in-memory GatewaySoftCache for soft-block tests.
type FakeBudgetCheck struct {
	mu      sync.Mutex
	entries map[string]domainbudget.GatewaySoftEntry
	gets    int
	enabled bool
}

func NewFakeBudgetCheck() *FakeBudgetCheck {
	return &FakeBudgetCheck{entries: map[string]domainbudget.GatewaySoftEntry{}, enabled: true}
}

func (f *FakeBudgetCheck) Enabled() bool { return f.enabled }

func (f *FakeBudgetCheck) Get(_ context.Context, companyID int64, keyHash string) (domainbudget.GatewaySoftEntry, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gets++
	entry, ok := f.entries[gatewaySoftKey(companyID, keyHash)]
	return entry, ok, nil
}

func (f *FakeBudgetCheck) Set(_ context.Context, companyID int64, keyHash string, entry domainbudget.GatewaySoftEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[gatewaySoftKey(companyID, keyHash)] = entry
	return nil
}

// Gets returns the number of Get calls (for asserting the +1 Redis GET).
func (f *FakeBudgetCheck) Gets() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.gets
}

var _ domainbudget.GatewaySoftCache = (*FakeBudgetCheck)(nil)
