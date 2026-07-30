package common

import "math"

const ModelNotInDeptMessage = "该模型不在您部门的可用范围内"

const DefaultPersonalBudget = 0

// DefaultBillingCurrency is the only hardcoded billing currency code.
// Empty company.BillingCurrency resolves here; never substitute for ledger/lot rows.
const DefaultBillingCurrency = "CNY"

// DefaultQuotaPerUnit is the seed QPU for DefaultBillingCurrency (currencies table).
// 1 CNY = 10000 quota.
const DefaultQuotaPerUnit int64 = 10000

// ResolveBillingCurrency returns code, or DefaultBillingCurrency when empty.
func ResolveBillingCurrency(code string) string {
	if code == "" {
		return DefaultBillingCurrency
	}
	return code
}

const NewAPIGroupPrefix = "dept-"

// MoneyToQuota converts a currency amount (e.g. CNY) to quota using the given quotaPerUnit.
func MoneyToQuota(amount float64, quotaPerUnit int64) int64 {
	return int64(math.Round(amount * float64(quotaPerUnit)))
}

// QuotaToMoney converts quota to currency amount using the given quotaPerUnit.
func QuotaToMoney(quota int64, quotaPerUnit int64) float64 {
	if quotaPerUnit <= 0 {
		return 0
	}
	return float64(quota) / float64(quotaPerUnit)
}
