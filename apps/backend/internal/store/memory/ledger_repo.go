package memory

import (
	"context"
	"sort"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/usagequery"
)

type memoryLedgerRepo struct{ store *Store }

func (r *memoryLedgerRepo) InsertOnConflict(ctx context.Context, entry types.UsageLedgerEntry) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	for _, existing := range snap.UsageLedger {
		if existing.IdempotencyKey == entry.IdempotencyKey {
			return false, nil
		}
	}
	entry.CompanyID = tid
	snap.UsageLedger = append(snap.UsageLedger, store.CloneUsageLedgerEntry(entry))
	r.store.setCompanySnapshot(tid, snap)
	return true, nil
}

func (r *memoryLedgerRepo) ListCallSettledPage(ctx context.Context, filter store.LedgerCallFilter) ([]types.UsageLedgerEntry, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	snap := r.store.companySnapshot(store.CompanyID(ctx))
	items := make([]types.UsageLedgerEntry, 0)
	for _, entry := range snap.UsageLedger {
		if entry.EventType != types.EventTypeCallSettled {
			continue
		}
		if !matchesLedgerCallFilter(entry, filter) {
			continue
		}
		items = append(items, store.CloneUsageLedgerEntry(entry))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].OccurredAt.After(items[j].OccurredAt)
	})
	total := len(items)
	page, pageSize := normalizeLedgerPage(filter.Page, filter.PageSize)
	start := (page - 1) * pageSize
	if start >= total {
		return []types.UsageLedgerEntry{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func (r *memoryLedgerRepo) QueryMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	loc, err := common.LoadLocation(q.Timezone)
	if err != nil {
		return nil, err
	}
	snap := r.store.companySnapshot(store.CompanyID(ctx))
	bucketRows := make([]types.UsageBucketRow, 0)
	for _, entry := range snap.UsageLedger {
		if entry.EventType != types.EventTypeCallSettled {
			continue
		}
		if entry.OccurredAt.Before(q.Start) || !entry.OccurredAt.Before(q.End) {
			continue
		}
		if q.DepartmentID != "" && entry.DepartmentID != q.DepartmentID {
			continue
		}
		if q.MemberID != "" {
			memberID := ""
			if entry.MemberID != nil {
				memberID = *entry.MemberID
			}
			if memberID != q.MemberID {
				continue
			}
		}
		if len(q.ScopeDeptIDs) > 0 && !usagequery.ContainsString(q.ScopeDeptIDs, entry.DepartmentID) {
			continue
		}
		memberID := ""
		if entry.MemberID != nil {
			memberID = *entry.MemberID
		}
		bucketRows = append(bucketRows, types.UsageBucketRow{
			BucketStart:  usagequery.TruncateBucket(entry.OccurredAt, types.UsageGranularityMinute, loc),
			DepartmentID: entry.DepartmentID,
			MemberID:     memberID,
			Model:        entry.Model,
			CostCNY:      entry.AmountCNY,
			CallCount:    1,
			InputTokens:  entry.InputTokens,
			OutputTokens: entry.OutputTokens,
		})
	}
	aggregated := usagequery.AggregateRows(bucketRows, types.UsageGranularityMinute, q.GroupBy, loc)
	points := make([]types.UsageSeriesPoint, len(aggregated))
	for i, row := range aggregated {
		points[i] = usagequery.AggregateToSeriesPoint(row)
	}
	usagequery.SortSeriesPoints(points)
	return points, nil
}

func matchesLedgerCallFilter(entry types.UsageLedgerEntry, filter store.LedgerCallFilter) bool {
	if filter.Model != "" && entry.Model != filter.Model {
		return false
	}
	if filter.Status != "" && entry.CallDetail.Status != filter.Status {
		return false
	}
	if filter.CallerID != "" && entry.CallDetail.CallerID != filter.CallerID {
		return false
	}
	createdAt := pkgtime.FormatSyncLog(entry.OccurredAt.UTC())
	day := createdAt
	if len(day) > 10 {
		day = day[:10]
	}
	if filter.From != "" && day < filter.From {
		return false
	}
	if filter.To != "" && day > filter.To {
		return false
	}
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		q := strings.ToLower(kw)
		fields := []string{entry.Model, entry.CallDetail.Caller, entry.CallDetail.PreviewSnippet}
		matched := false
		for _, field := range fields {
			if strings.Contains(strings.ToLower(field), q) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func normalizeLedgerPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

var _ store.LedgerRepository = (*memoryLedgerRepo)(nil)
