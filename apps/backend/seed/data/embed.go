package data

import _ "embed"

//go:embed platform_keys.json
var PlatformKeysJSON []byte

//go:embed operation_logs.json
var OperationLogsJSON []byte

//go:embed usage_ledger.json
var UsageLedgerJSON []byte
