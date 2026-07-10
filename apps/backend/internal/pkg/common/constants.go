package common

const ModelNotInDeptMessage = "该模型不在您部门的可用范围内"

const DefaultPersonalQuota = 5000

const DefaultModelPricePoint = 1000

const DefaultPointsPerUnit = 1000

const QuotaPerUnit = 500000

// WalletSyncDriftEpsilon is the max allowed Postgres vs NewAPI point drift before reconcile.
const WalletSyncDriftEpsilon = 0.01 * float64(DefaultPointsPerUnit)

// WalletSyncDebounceSecs delays wallet_sync execution after ingest/recharge bursts.
const WalletSyncDebounceSecs = 5

// WalletSyncRetryAfterSecs is returned when gateway rejects due to pending wallet_sync lag.
const WalletSyncRetryAfterSecs = WalletSyncDebounceSecs

const NewAPIGroupPrefix = "dept-"
