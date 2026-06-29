package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/memory"
)

func NewMemoryStoreFromSnapshot(snapshot store.Snapshot) store.Store {
	return memory.New(snapshot)
}

func MustMemoryStore(t *testing.T, st store.Store) *memory.Store {
	t.Helper()
	mem, ok := st.(*memory.Store)
	if !ok {
		t.Fatal("expected memory store")
	}
	return mem
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
