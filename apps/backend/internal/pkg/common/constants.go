package common

const ModelNotInDeptMessage = "该模型不在您部门的可用范围内"

const DefaultPersonalBudget = 0

const DefaultModelPricePoint = 1000

const DefaultPointsPerUnit = 1000

const QuotaPerUnit = 500000

// WalletSyncDriftEpsilon is the max allowed Postgres vs NewAPI point drift before reconcile.
const WalletSyncDriftEpsilon = 0.01 * float64(DefaultPointsPerUnit)

// WalletSyncDebounceSecs delays wallet_sync execution after ingest/recharge bursts.
const WalletSyncDebounceSecs = 5

const NewAPIGroupPrefix = "dept-"
