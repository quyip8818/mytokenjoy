package common

import "math"

const ModelNotInDeptMessage = "该模型不在您部门的可用范围内"

const DefaultPersonalBudget = 0

// DefaultBillingCurrency is the only hardcoded billing currency code.
// Empty company.BillingCurrency resolves here; never substitute for ledger/lot rows.
const DefaultBillingCurrency = "CNY"

// DefaultQuotaPerUnit is the seed QPU for DefaultBillingCurrency (currencies table).
// 1 CNY = 500000 quota. This aligns with NewAPI's internal QuotaPerUnit so that
// model_ratio=1 traffic flows through without conversion.
const DefaultQuotaPerUnit int64 = 500000

// ResolveBillingCurrency returns code, or DefaultBillingCurrency when empty.
func ResolveBillingCurrency(code string) string {
	if code == "" {
		return DefaultBillingCurrency
	}
	return code
}

const NewAPIGroupPrefix = "dept-"

// QuotaFromAmount converts a display amount (e.g. CNY) to quota using the given quotaPerUnit.
func QuotaFromAmount(amount float64, quotaPerUnit int64) int64 {
	return int64(math.Round(amount * float64(quotaPerUnit)))
}

// QuotaToDisplay converts quota to display amount using the given quotaPerUnit.
func QuotaToDisplay(quota int64, quotaPerUnit int64) float64 {
	if quotaPerUnit <= 0 {
		return 0
	}
	return float64(quota) / float64(quotaPerUnit)
}
