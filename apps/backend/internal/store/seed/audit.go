package seed

import (
	_ "embed"
	"encoding/json"

	"github.com/tokenjoy/backend/internal/domain/types"
)

//go:embed data/operation_logs.json
var operationLogsJSON []byte

//go:embed data/call_logs.json
var callLogsJSON []byte

func buildAuditSettings() types.AuditSettings {
	return types.AuditSettings{ContentRetentionEnabled: true}
}

func loadOperationLogs() []types.OperationLog {
	var logs []types.OperationLog
	if err := json.Unmarshal(operationLogsJSON, &logs); err != nil {
		panic("seed: load operation logs: " + err.Error())
	}
	return logs
}

func loadCallLogs() []types.CallLog {
	var logs []types.CallLog
	if err := json.Unmarshal(callLogsJSON, &logs); err != nil {
		panic("seed: load call logs: " + err.Error())
	}
	return logs
}
