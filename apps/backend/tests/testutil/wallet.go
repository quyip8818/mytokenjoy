package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
)

func MustWallet(t *testing.T, cfg config.Config, client interface {
	GetUserQuota(ctx context.Context, userID int64) (int64, error)
}) company.WalletService {
	t.Helper()
	wallet := company.NewWalletService(cfg, client)
	return wallet
}
