// Package quota provides billing currency and quota conversion helpers.
// These are pure functions with no external dependencies, safe to import
// from any layer (domain, store, infra, support).
package quota

import "math"

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
