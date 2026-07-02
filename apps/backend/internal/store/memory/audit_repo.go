package memory

import (
	"context"
	"sort"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryAuditRepo struct{ store *Store }

func (r *memoryAuditRepo) Settings(ctx context.Context) (types.AuditSettings, error) {
	if err := ctx.Err(); err != nil {
		return types.AuditSettings{}, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.companySnapshot(store.CompanyID(ctx)).AuditSettings, nil
}

func (r *memoryAuditRepo) SetSettings(ctx context.Context, settings types.AuditSettings) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.AuditSettings = settings
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryAuditRepo) OperationLogs(ctx context.Context) ([]types.OperationLog, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneOperationLogs(r.store.companySnapshot(store.CompanyID(ctx)).OperationLogs), nil
}

func (r *memoryAuditRepo) ListOperationsPage(ctx context.Context, filter store.AuditOperationFilter) ([]types.OperationLog, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	items := make([]types.OperationLog, 0)
	for _, item := range r.store.companySnapshot(store.CompanyID(ctx)).OperationLogs {
		if !matchesAuditOperationFilter(item, filter) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	total := len(items)
	page, pageSize := normalizeAuditPage(filter.Page, filter.PageSize)
	start := (page - 1) * pageSize
	if start >= total {
		return []types.OperationLog{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return store.CloneOperationLogs(items[start:end]), total, nil
}

func matchesAuditOperationFilter(item types.OperationLog, filter store.AuditOperationFilter) bool {
	if filter.Action != "" && item.Action != filter.Action {
		return false
	}
	if filter.OperatorID != "" && item.OperatorID != filter.OperatorID {
		return false
	}
	day := item.CreatedAt
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
		fields := []string{item.Detail, item.Target, item.Operator}
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

func normalizeAuditPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

func (r *memoryAuditRepo) AppendOperationLog(ctx context.Context, log types.OperationLog) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.OperationLogs = append([]types.OperationLog{log}, snap.OperationLogs...)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

var _ store.AuditRepository = (*memoryAuditRepo)(nil)
