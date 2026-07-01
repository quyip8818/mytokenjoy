package testutil

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/memory"
)

func NewMemoryStoreFromSnapshot(snapshot store.Snapshot) store.Store {
	return memory.New(snapshot)
}

func UsageBucketRows(st store.Store) []types.UsageBucketRow {
	mem, ok := st.(*memory.Store)
	if !ok {
		return nil
	}
	return mem.UsageBucketRows()
}

func NotificationLogs(st store.Store) []types.NotificationLogEntry {
	mem, ok := st.(*memory.Store)
	if !ok {
		return nil
	}
	return mem.NotificationLogs()
}

func RelayOutboxEntry(st store.Store, id string) (store.RelayOutboxEntry, bool) {
	mem, ok := st.(*memory.Store)
	if !ok {
		return store.RelayOutboxEntry{}, false
	}
	return mem.RelayOutboxEntry(id)
}

func PendingRebalanceCount(st store.Store, companyID int64) int {
	ctx := CtxForCompany(companyID)
	entries, err := st.Relay().ClaimPendingRebalance(ctx, 100)
	if err != nil {
		return 0
	}
	return len(entries)
}
