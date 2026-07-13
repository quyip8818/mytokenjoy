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

type ReconcileService struct {
	cfg      config.Config
	store    store.Store
	enqueuer JobEnqueuer
	logger   *slog.Logger
}

func NewReconcileService(cfg config.Config, st store.Store, enqueuer JobEnqueuer, logger *slog.Logger) *ReconcileService {
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
	ctx = company.WithContext(ctx, companyContextFromStore(*co))

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
		var memberID string
		if entry.MemberID != nil {
			memberID = *entry.MemberID
		}
		key := bucketKey{
			BucketStart:  entry.OccurredAt.UTC().Truncate(time.Hour),
			DepartmentID: entry.DepartmentID,
			MemberID:     memberID,
			Model:        entry.Model,
		}
		row := out[key]
		row.BucketStart = key.BucketStart
		row.DepartmentID = key.DepartmentID
		row.MemberID = key.MemberID
		row.Model = key.Model
		row.Cost += entry.Amount
		row.CallCount++
		row.InputTokens += entry.InputTokens
		row.OutputTokens += entry.OutputTokens
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
