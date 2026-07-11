// Package budgetcheck defines the optional GatewayBudgetCheck soft-block cache.
//
// It is a best-effort, non-authoritative signal written by the async budget
// projector/reconciler and read by the Gateway precheck. A miss MUST degrade to
// "allow"; the Gateway never falls back to Postgres for this signal.
package budgetcheck

import (
	"context"
	"fmt"
	"time"
)

// Entry is the value stored per (company, key_hash).
//
// The period_key is carried in the value (not the Redis key) so the Gateway
// never needs to recompute the budget period (which would require joining
// org_nodes). See docs/实现-异步预算投影.md §10.
type Entry struct {
	PeriodKey  string    `json:"periodKey"`
	SoftRemain float64   `json:"softRemain"`
	KeyStatus  string    `json:"keyStatus"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Store is the GatewayBudgetCheck cache contract.
type Store interface {
	// Enabled reports whether a real backing store is configured. Callers on the
	// write path MUST skip all computation when this is false.
	Enabled() bool
	// Get returns the cached entry. ok=false means miss (degrade to allow).
	Get(ctx context.Context, companyID int64, keyHash string) (Entry, bool, error)
	// Set writes the entry with the store's configured TTL.
	Set(ctx context.Context, companyID int64, keyHash string, entry Entry) error
}

// Blocks reports whether a cached entry should soft-block the Gateway request.
func Blocks(entry Entry) bool {
	return entry.SoftRemain <= 0
}

// Key builds the Redis key: gateway:budget_check:{company_id}:{key_hash}.
func Key(companyID int64, keyHash string) string {
	return fmt.Sprintf("gateway:budget_check:%d:%s", companyID, keyHash)
}

// Noop is the default store used when REDIS_URL is empty. Get always misses and
// Set is a cheap no-op.
type Noop struct{}

func (Noop) Enabled() bool { return false }

func (Noop) Get(context.Context, int64, string) (Entry, bool, error) {
	return Entry{}, false, nil
}

func (Noop) Set(context.Context, int64, string, Entry) error { return nil }

var _ Store = Noop{}
