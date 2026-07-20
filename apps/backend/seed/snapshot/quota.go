package snapshot

import "github.com/tokenjoy/backend/internal/pkg/common"

func seedQuota(display float64) int64 {
	return common.QuotaFromAmount(display, common.DefaultQuotaPerUnit)
}
