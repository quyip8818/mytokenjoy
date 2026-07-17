//go:build testhook

package gatewayfix

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
)

func gatewaySoftKey(companyID uuid.UUID, keyHash string) string {
	return fmt.Sprintf("gateway:budget_check:%d:%s", companyID, keyHash)
}

// FakeBudgetCheck is an in-memory CombinedKeyCache for combined key budget tests.
type FakeBudgetCheck struct {
	mu      sync.Mutex
	entries map[string]domainbudget.CombinedKeyEntry
	gets    int
	enabled bool
}

func NewFakeBudgetCheck() *FakeBudgetCheck {
	return &FakeBudgetCheck{entries: map[string]domainbudget.CombinedKeyEntry{}, enabled: true}
}

func (f *FakeBudgetCheck) Enabled() bool { return f.enabled }

func (f *FakeBudgetCheck) Get(_ context.Context, companyID uuid.UUID, keyHash string) (domainbudget.CombinedKeyEntry, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gets++
	entry, ok := f.entries[gatewaySoftKey(companyID, keyHash)]
	return entry, ok, nil
}

func (f *FakeBudgetCheck) Set(_ context.Context, companyID uuid.UUID, keyHash string, entry domainbudget.CombinedKeyEntry) error {
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

var _ domainbudget.CombinedKeyCache = (*FakeBudgetCheck)(nil)
