package memory

import (
	"sync"

	"github.com/tokenjoy/backend/internal/store"
)

type memoryRelayRepo struct {
	store *Store
	mu    sync.Mutex
	data  struct {
		mappings      map[string]store.RelayMapping
		tokenIndex    map[int64]string
		relayOutbox   []store.RelayOutboxEntry
		rebalance     []store.RebalanceQueueEntry
		overrun       []store.OverrunQueueEntry
	}
}

func newMemoryRelayRepo(m *Store) *memoryRelayRepo {
	r := &memoryRelayRepo{store: m}
	r.data.mappings = make(map[string]store.RelayMapping)
	r.data.tokenIndex = make(map[int64]string)
	return r
}

func mappingBelongsToCompany(mapping store.RelayMapping, companyID int64) bool {
	return mapping.CompanyID == companyID
}

var _ store.RelayRepository = (*memoryRelayRepo)(nil)
