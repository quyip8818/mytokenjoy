package snapshot

import "github.com/tokenjoy/backend/internal/support/quota"

func seedQuota(money float64) int64 {
	return quota.MoneyToQuota(money, quota.DefaultQuotaPerUnit)
}
