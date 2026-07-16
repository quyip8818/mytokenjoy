package dashboard

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

const reconcileEpsilon = 0.000001

// ReconcileStore is the narrow store surface the dashboard reconcile service needs.
type ReconcileStore interface {
	Company() store.CompanyRepository
	Ledger() store.LedgerRepository
	Usage() store.UsageRepository
}

type ReconcileService struct {
	cfg      config.Config
	store    ReconcileStore
	enqueuer JobEnqueuer
	logger   *slog.Logger
}

func NewReconcileService(cfg config.Config, st ReconcileStore, enqueuer JobEnqueuer, logger *slog.Logger) *ReconcileService {
	if enqueuer == nil {
		enqueuer = NoopJobEnqueuer
	}
	return &ReconcileService{cfg: cfg, store: st, enqueuer: enqueuer, logger: logger}
}

type bucketKey struct {
	BucketStart  time.Time
	DepartmentID string
	MemberID     string
	Model        string
}

func (s *ReconcileService) RunCompany(ctx context.Context, companyID int64) error {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil {
		return nil
	}
	ctx = company.WithContext(ctx, company.ContextFromStore(*co))

	since := reconcileWindowStart(clock.NowUTC(s.cfg.Clock()))
	entries, err := s.store.Ledger().ListCallSettledSince(ctx, since)
	if err != nil {
		return err
	}
	expected := expectedBuckets(entries)
	actualRows, err := s.store.Usage().ListBucketsSince(ctx, since)
	if err != nil {
		return err
	}
	actual := map[bucketKey]types.UsageBucketRow{}
	for _, row := range actualRows {
		actual[bucketKeyFromRow(row)] = row
	}

	for key, want := range expected {
		got, ok := actual[key]
		if ok && !bucketDrift(want, got) {
			continue
		}
		if err := s.store.Usage().SetBucket(ctx, want); err != nil {
			return err
		}
		if s.logger != nil {
			s.logger.Warn("dashboard reconcile drift repaired",
				"company_id", companyID,
				"bucket_start", key.BucketStart,
				"department_id", key.DepartmentID,
				"member_id", key.MemberID,
				"model", key.Model,
			)
		}
	}
	return nil
}

func expectedBuckets(entries []types.UsageLedgerEntry) map[bucketKey]types.UsageBucketRow {
	out := make(map[bucketKey]types.UsageBucketRow)
	for _, entry := range entries {
		delta := bucketFromLedgerEntry(entry)
		key := bucketKeyFromRow(delta)
		row := out[key]
		mergeBucketDelta(&row, delta)
		out[key] = row
	}
	return out
}

func bucketKeyFromRow(row types.UsageBucketRow) bucketKey {
	return bucketKey{
		BucketStart:  row.BucketStart.UTC().Truncate(time.Hour),
		DepartmentID: row.DepartmentID,
		MemberID:     row.MemberID,
		Model:        row.Model,
	}
}

func bucketDrift(expected, actual types.UsageBucketRow) bool {
	return !floatClose(expected.Cost, actual.Cost) ||
		!floatClose(expected.DisplayCost, actual.DisplayCost) ||
		expected.CallCount != actual.CallCount ||
		expected.InputTokens != actual.InputTokens ||
		expected.OutputTokens != actual.OutputTokens
}

func floatClose(a, b float64) bool {
	return math.Abs(a-b) <= reconcileEpsilon
}

func reconcileWindowStart(now time.Time) time.Time {
	return now.AddDate(0, 0, -90)
}
