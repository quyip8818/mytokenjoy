//go:build testhook

package budgetfix

import "github.com/tokenjoy/backend/internal/pkg/common"

func FloatPtr(v float64) *float64 { return &v }

func Int64Ptr(v int64) *int64 { return &v }

// QuotaFromMoney converts a currency amount (e.g. CNY) to int64 quota.
func QuotaFromMoney(money float64) int64 {
	return common.MoneyToQuota(money, common.DefaultQuotaPerUnit)
}
