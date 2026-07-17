package usage

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

// IngestJobEnqueuer enqueues side-effect jobs within the ingest transaction.
type IngestJobEnqueuer interface {
	EnqueueAfterIngest(ctx context.Context, tx store.Tx, companyID uuid.UUID, effects *IngestEffects) error
}

// BudgetOps is the port for budget domain operations needed during ingest.
type BudgetOps interface {
	// ConsumptionDeltas computes budget axis increments for a settled entry.
	ConsumptionDeltas(ctx context.Context, entry types.UsageLedgerEntry, open pkgbudget.OpenBudgetPeriod) ([]ConsumedDelta, error)
	// RefreshCombinedKeySummaries writes PG-derived summaries into the optional cache (best-effort).
	RefreshCombinedKeySummaries(ctx context.Context, companyID uuid.UUID, summaries []store.CombinedKeySummary)
	// CheckBudgetAlerts evaluates percentage alert rules for touched departments (best-effort post-commit).
	CheckBudgetAlerts(ctx context.Context, st store.Store, companyID uuid.UUID, touchedDepts map[uuid.UUID]struct{})
	// ComputeGatewaySummaryUpdates recomputes combined key remain for touched platform keys.
	ComputeGatewaySummaryUpdates(ctx context.Context, st store.Store, keyIDs map[uuid.UUID]struct{}, clk clock.Clock) ([]store.CombinedKeySummaryUpdate, error)
}

// ConsumedDelta represents a single axis budget increment.
type ConsumedDelta struct {
	Kind      string
	AxisID    uuid.UUID
	PeriodKey string
	Amount    float64
}

// LotConsumer is the port for billing lot consumption during ingest.
type LotConsumer interface {
	// ConsumeLotsLocked consumes lots assuming the company row is already locked.
	ConsumeLotsLocked(ctx context.Context, st store.Store, co *store.Company, amountPoint float64) (LotConsumeResult, error)
	// LedgerSegmentsFromEntry builds ledger segment entries from lot consumption segments.
	LedgerSegmentsFromEntry(base types.UsageLedgerEntry, segs []LotSegment) []types.UsageLedgerEntry
}

// LotConsumeResult is the outcome of lot consumption.
type LotConsumeResult struct {
	Segments       []LotSegment
	OverdraftUsed  bool
	OverdraftDelta float64
}

// LotSegment represents a single lot consumption segment.
type LotSegment struct {
	LotID           uuid.UUID
	Points          float64
	DisplayAmount   float64
	BillingCurrency string
}
