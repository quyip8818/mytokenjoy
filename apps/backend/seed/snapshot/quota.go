package snapshot

import "github.com/tokenjoy/backend/internal/pkg/common"

func seedQuota(money float64) int64 {
	return common.MoneyToQuota(money, common.DefaultQuotaPerUnit)
}
