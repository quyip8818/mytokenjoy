package memory

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

type txSnapshot struct {
	companies     map[int64]store.Company
	dataByCompany map[int64]store.Snapshot
	invites       map[string]store.CompanyInvite
}

func (m *Store) captureTxSnapshot() txSnapshot {
	companies := make(map[int64]store.Company, len(m.companies))
	for k, v := range m.companies {
		companies[k] = v
	}
	dataByCompany := make(map[int64]store.Snapshot, len(m.dataByCompany))
	for k, v := range m.dataByCompany {
		dataByCompany[k] = store.CloneSnapshot(v)
	}
	var invites map[string]store.CompanyInvite
	if m.invites != nil {
		invites = make(map[string]store.CompanyInvite, len(m.invites))
		for k, v := range m.invites {
			invites[k] = v
		}
	}
	return txSnapshot{companies: companies, dataByCompany: dataByCompany, invites: invites}
}

func (m *Store) restoreTxSnapshot(snap txSnapshot) {
	m.companies = snap.companies
	m.dataByCompany = snap.dataByCompany
	m.invites = snap.invites
}

func (m *Store) WithTx(ctx context.Context, fn func(store.Store) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	snap := m.captureTxSnapshot()
	m.mu.Unlock()

	if err := fn(m); err != nil {
		m.mu.Lock()
		m.restoreTxSnapshot(snap)
		m.mu.Unlock()
		return err
	}
	return nil
}
