package relayfix

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
)

type StubWallet struct {
	Quota int64
}

func NewStubWallet(quota int64) *StubWallet {
	return &StubWallet{Quota: quota}
}

func (s *StubWallet) AvailableQuota(_ context.Context, _ int64) (int64, error) {
	return s.Quota, nil
}

var _ company.WalletService = (*StubWallet)(nil)
