package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
)

func MustWallet(t *testing.T, cfg config.Config, reader company.QuotaReader) company.WalletService {
	t.Helper()
	return company.NewWalletService(cfg, reader)
}
