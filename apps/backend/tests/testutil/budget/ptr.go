//go:build testhook

package budgetfix

import "github.com/tokenjoy/backend/internal/pkg/common"

func FloatPtr(v float64) *float64 { return &v }

func Int64Ptr(v int64) *int64 { return &v }

// QuotaFromDisplay converts a display-currency amount (e.g. CNY) to int64 quota.
func QuotaFromDisplay(display float64) int64 {
	return common.QuotaFromAmount(display, common.DefaultQuotaPerUnit)
}
