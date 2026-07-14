package snapshot

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/data"
)

func buildAuditSettings() types.AuditSettings {
	return types.AuditSettings{ContentRetentionEnabled: true}
}

func loadOperationLogs() []types.OperationLog {
	var logs []types.OperationLog
	if err := json.Unmarshal(data.OperationLogsJSON, &logs); err != nil {
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
	if err := json.Unmarshal(data.UsageLedgerJSON, &rows); err != nil {
		panic("seed: load usage ledger: " + err.Error())
	}
	keyScope := make(map[string]string, len(loadPlatformKeys()))
	for _, key := range loadPlatformKeys() {
		keyScope[key.ID] = key.Scope
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
		platformKeyID := contract.IDPlatformKey1
		if row.CallerType == types.CallerTypePlatformKey {
			platformKeyID = row.CallerID
		}
		scope := keyScope[platformKeyID]
		if scope == "" {
			panic(fmt.Sprintf("seed: unknown platform key %q for ledger row %s", platformKeyID, row.ID))
		}
		entries = append(entries, types.UsageLedgerEntry{
			ID:               row.ID,
			EventType:        types.EventTypeCallSettled,
			IdempotencyKey:   fmt.Sprintf("%sseed-%d", types.IdempotencyPrefixNewAPI, i+1),
			LotID:            "tu-1",
			Amount:           seedPoints(row.Cost),
			DisplayAmount:    row.Cost,
			BillingCurrency:  common.DefaultBillingCurrency,
			DepartmentID:     contract.IDDept3,
			MemberID:         memberID,
			PlatformKeyID:    platformKeyID,
			PlatformKeyScope: scope,
			Source:           types.SourceWebhook,
			OccurredAt:       occurredAt.UTC(),
			PeriodKey:        pkgbudget.OccurrenceSnapshotKey(pkgbudget.PeriodMonthly, occurredAt.UTC()).String(),
			Model:            row.Model,
			InputTokens:      int64(row.InputTokens),
			OutputTokens:     int64(row.OutputTokens),
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
