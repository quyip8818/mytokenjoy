package common

const ModelNotInDeptMessage = "该模型不在您部门的可用范围内"

const DefaultPersonalBudget = 0

const DefaultModelPricePoint = 1000

// DefaultBillingCurrency is the only hardcoded billing currency code.
// Empty company.BillingCurrency resolves here; never substitute for ledger/lot rows.
const DefaultBillingCurrency = "CNY"

// DefaultPointsPerUnit is the seed PPU for DefaultBillingCurrency (currencies table).
const DefaultPointsPerUnit = 1000

// ResolveBillingCurrency returns code, or DefaultBillingCurrency when empty.
func ResolveBillingCurrency(code string) string {
	if code == "" {
		return DefaultBillingCurrency
	}
	return code
}

const QuotaPerUnit = 500000

// WalletSyncDriftEpsilon is the max allowed Postgres vs NewAPI point drift before reconcile.
const WalletSyncDriftEpsilon = 0.01 * float64(DefaultPointsPerUnit)

// WalletSyncDebounceSecs delays wallet_sync execution after ingest/recharge bursts.
const WalletSyncDebounceSecs = 5

const NewAPIGroupPrefix = "dept-"
