package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryRelayRepo struct {
	store *Store
	mu    sync.Mutex
	data  struct {
		mappings      map[string]store.RelayMapping
		tokenIndex    map[int64]string
		relayOutbox   []store.RelayOutboxEntry
		webhookOutbox []store.WebhookOutboxEntry
		ingestedLogs  map[int64]struct{}
		lastLogID     int64
		rebalance     []store.RebalanceQueueEntry
	}
}

func newMemoryRelayRepo(m *Store) *memoryRelayRepo {
	r := &memoryRelayRepo{store: m}
	r.data.mappings = make(map[string]store.RelayMapping)
	r.data.tokenIndex = make(map[int64]string)
	r.data.ingestedLogs = make(map[int64]struct{})
	return r
}

func mappingBelongsToCompany(mapping store.RelayMapping, companyID int64) bool {
	return mapping.CompanyID == companyID
}

func (r *memoryRelayRepo) GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok || !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return nil, nil
	}
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) GetMappingByFullKey(ctx context.Context, fullKey string) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for companyID, snap := range r.store.dataByCompany {
		for _, key := range snap.PlatformKeys {
			if key.FullKey == nil || *key.FullKey != fullKey {
				continue
			}
			r.mu.Lock()
			m, ok := r.data.mappings[key.ID]
			r.mu.Unlock()
			if !ok || m.CompanyID != companyID {
				continue
			}
			copy := m
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *memoryRelayRepo) GetMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.data.tokenIndex[tokenID]
	if !ok {
		return nil, nil
	}
	m := r.data.mappings[id]
	if !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return nil, nil
	}
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) FindMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.data.tokenIndex[tokenID]
	if !ok {
		return nil, nil
	}
	m := r.data.mappings[id]
	copy := m
	return &copy, nil
}

func (r *memoryRelayRepo) ListMappingsByMemberID(ctx context.Context, memberID string) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.MemberID != nil && *m.MemberID == memberID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.DepartmentID == departmentID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.BudgetGroupID != nil && *m.BudgetGroupID == budgetGroupID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListActiveMappings(ctx context.Context) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tid := store.CompanyID(ctx)
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, tid) {
			continue
		}
		if m.SyncStatus == store.RelaySyncStatusSynced {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]store.RelayMapping, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]store.RelayMapping, 0)
	for _, m := range r.data.mappings {
		if !mappingBelongsToCompany(m, companyID) {
			continue
		}
		if m.SyncStatus == store.RelaySyncStatusSynced {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) UpsertMapping(ctx context.Context, mapping store.RelayMapping) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	mapping.CompanyID = store.CompanyID(ctx)
	r.data.mappings[mapping.PlatformKeyID] = mapping
	if mapping.NewAPITokenID != nil {
		r.data.tokenIndex[*mapping.NewAPITokenID] = mapping.PlatformKeyID
	}
	return nil
}

func (r *memoryRelayRepo) UpdateMappingSync(
	ctx context.Context,
	platformKeyID string,
	tokenID int64,
	status string,
	remainQuota *int64,
	syncedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok || !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return fmt.Errorf("mapping not found: %s", platformKeyID)
	}
	m.NewAPITokenID = &tokenID
	m.SyncStatus = status
	m.SyncedAt = &syncedAt
	m.RelayRemainQuota = remainQuota
	r.data.mappings[platformKeyID] = m
	r.data.tokenIndex[tokenID] = platformKeyID
	return nil
}

func (r *memoryRelayRepo) UpdateMappingRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data.mappings[platformKeyID]
	if !ok || !mappingBelongsToCompany(m, store.CompanyID(ctx)) {
		return fmt.Errorf("mapping not found: %s", platformKeyID)
	}
	m.RelayRemainQuota = &remainQuota
	r.data.mappings[platformKeyID] = m
	return nil
}

func (r *memoryRelayRepo) EnqueueRelayOutbox(ctx context.Context, entry store.RelayOutboxEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.relayOutbox = append(r.data.relayOutbox, entry)
	return nil
}

func (r *memoryRelayRepo) ClaimPendingRelayOutbox(ctx context.Context, limit int) ([]store.RelayOutboxEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := make([]store.RelayOutboxEntry, 0, limit)
	for i := range r.data.relayOutbox {
		if len(out) >= limit {
			break
		}
		e := r.data.relayOutbox[i]
		if e.Status == store.OutboxStatusPending && !e.NextRetry.After(now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkRelayOutboxDone(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.relayOutbox {
		if r.data.relayOutbox[i].ID == id {
			r.data.relayOutbox[i].Status = store.OutboxStatusDone
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) MarkRelayOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.relayOutbox {
		if r.data.relayOutbox[i].ID == id {
			r.data.relayOutbox[i].Attempts++
			r.data.relayOutbox[i].NextRetry = nextRetry
			errMsg := lastError
			r.data.relayOutbox[i].LastError = &errMsg
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) relayOutboxEntry(id string) (store.RelayOutboxEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range r.data.relayOutbox {
		if entry.ID == id {
			return entry, true
		}
	}
	return store.RelayOutboxEntry{}, false
}

func (r *memoryRelayRepo) EnqueueWebhookOutbox(ctx context.Context, entry store.WebhookOutboxEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.webhookOutbox = append(r.data.webhookOutbox, entry)
	return nil
}

func (r *memoryRelayRepo) ClaimPendingWebhookOutbox(ctx context.Context, limit int) ([]store.WebhookOutboxEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := make([]store.WebhookOutboxEntry, 0, limit)
	for i := range r.data.webhookOutbox {
		if len(out) >= limit {
			break
		}
		e := r.data.webhookOutbox[i]
		if e.Status == store.OutboxStatusPending && !e.NextRetry.After(now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkWebhookOutboxDone(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.webhookOutbox {
		if r.data.webhookOutbox[i].ID == id {
			r.data.webhookOutbox[i].Status = store.OutboxStatusDone
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) MarkWebhookOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.webhookOutbox {
		if r.data.webhookOutbox[i].ID == id {
			r.data.webhookOutbox[i].Attempts++
			r.data.webhookOutbox[i].NextRetry = nextRetry
			errMsg := lastError
			r.data.webhookOutbox[i].LastError = &errMsg
			return nil
		}
	}
	return nil
}

func (r *memoryRelayRepo) HasIngestedLogID(ctx context.Context, logID int64) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data.ingestedLogs[logID]
	return ok, nil
}

func (r *memoryRelayRepo) InsertIngestedLogID(ctx context.Context, logID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.ingestedLogs[logID] = struct{}{}
	return nil
}

func (r *memoryRelayRepo) GetLastLogID(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data.lastLogID, nil
}

func (r *memoryRelayRepo) SetLastLogID(ctx context.Context, logID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.lastLogID = logID
	return nil
}

func (r *memoryRelayRepo) EnqueueRebalance(ctx context.Context, axisKind, axisID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	companyID := store.CompanyID(ctx)
	for _, e := range r.data.rebalance {
		if e.CompanyID == companyID && e.AxisKind == axisKind && e.AxisID == axisID && e.Status == store.OutboxStatusPending {
			return nil
		}
	}
	r.data.rebalance = append(r.data.rebalance, store.RebalanceQueueEntry{
		ID:        fmt.Sprintf("rb-%d-%s-%s-%d", companyID, axisKind, axisID, time.Now().UnixNano()),
		CompanyID: companyID,
		AxisKind:  axisKind,
		AxisID:    axisID,
		Status:    store.OutboxStatusPending,
	})
	return nil
}

func (r *memoryRelayRepo) ClaimPendingRebalance(ctx context.Context, limit int) ([]store.RebalanceQueueEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]store.RebalanceQueueEntry, 0, limit)
	for _, e := range r.data.rebalance {
		if len(out) >= limit {
			break
		}
		if e.Status == store.OutboxStatusPending {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memoryRelayRepo) MarkRebalanceDone(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.rebalance {
		if r.data.rebalance[i].ID == id {
			r.data.rebalance[i].Status = store.OutboxStatusDone
			return nil
		}
	}
	return nil
}

var _ store.RelayRepository = (*memoryRelayRepo)(nil)
