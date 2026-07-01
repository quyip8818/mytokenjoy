package memory

import (
	"context"

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

func (r *memoryAuditRepo) CallLogs(ctx context.Context) ([]types.CallLog, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneCallLogs(r.store.companySnapshot(store.CompanyID(ctx)).CallLogs), nil
}
