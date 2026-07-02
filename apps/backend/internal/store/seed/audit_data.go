package seed

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
)

//go:embed data/operation_logs.json
var operationLogsJSON []byte

//go:embed data/usage_ledger.json
var usageLedgerJSON []byte

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

type seedLedgerRow struct {
	ID             string  `json:"id"`
	Caller         string  `json:"caller"`
	CallerID       string  `json:"callerId"`
	CallerType     string  `json:"callerType"`
	Model          string  `json:"model"`
	Provider       string  `json:"provider"`
	InputTokens    float64 `json:"inputTokens"`
	OutputTokens   float64 `json:"outputTokens"`
	LatencyMs      float64 `json:"latencyMs"`
	Status         string  `json:"status"`
	Cost           float64 `json:"cost"`
	CreatedAt      string  `json:"createdAt"`
	PreviewSnippet string  `json:"previewSnippet"`
}

func loadUsageLedger() []types.UsageLedgerEntry {
	var rows []seedLedgerRow
	if err := json.Unmarshal(usageLedgerJSON, &rows); err != nil {
		panic("seed: load usage ledger: " + err.Error())
	}
	entries := make([]types.UsageLedgerEntry, 0, len(rows))
	for i, row := range rows {
		occurredAt, err := parseSeedLedgerTime(row.CreatedAt)
		if err != nil {
			panic("seed: parse ledger occurred_at: " + err.Error())
		}
		var memberID *string
		if row.CallerType == types.CallerTypeMember {
			memberID = &row.CallerID
		}
		entries = append(entries, types.UsageLedgerEntry{
			ID:             row.ID,
			EventType:      types.EventTypeCallSettled,
			IdempotencyKey: fmt.Sprintf("%sseed-%d", types.IdempotencyPrefixNewAPI, i+1),
			AmountCNY:      row.Cost,
			DepartmentID:   IDDept3,
			MemberID:       memberID,
			PlatformKeyID:  IDPlatformKey1,
			Source:         types.SourceWebhook,
			OccurredAt:     occurredAt.UTC(),
			Model:          row.Model,
			InputTokens:    int64(row.InputTokens),
			OutputTokens:   int64(row.OutputTokens),
			CallDetail: types.UsageCallDetail{
				Caller:         row.Caller,
				CallerID:       row.CallerID,
				CallerType:     row.CallerType,
				Provider:       row.Provider,
				Status:         row.Status,
				LatencyMs:      row.LatencyMs,
				PreviewSnippet: row.PreviewSnippet,
			},
			CreatedAt: occurredAt.UTC(),
		})
	}
	return entries
}

func parseSeedLedgerTime(value string) (time.Time, error) {
	loc, err := time.LoadLocation(types.UsageDefaultTimezone)
	if err != nil {
		return time.Time{}, err
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", value, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", value, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", value, loc); err == nil {
		return t, nil
	}
	return pkgtime.Parse(value)
}
