package usage

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func NewAPIIdempotencyKey(logID int64) string {
	return types.IdempotencyPrefixNewAPI + fmt.Sprintf("%d", logID)
}

func ParseNewAPILogID(idempotencyKey string) (int64, bool) {
	if !strings.HasPrefix(idempotencyKey, types.IdempotencyPrefixNewAPI) {
		return 0, false
	}
	var logID int64
	if _, err := fmt.Sscanf(idempotencyKey[len(types.IdempotencyPrefixNewAPI):], "%d", &logID); err != nil {
		return 0, false
	}
	return logID, true
}

func TruncatePreview(input string) string {
	runes := []rune(input)
	if len(runes) <= types.PreviewSnippetMaxLen {
		return input
	}
	return string(runes[:types.PreviewSnippetMaxLen])
}

func OccurredAtFromPayload(createdAt int64) (time.Time, error) {
	if createdAt <= 0 {
		return time.Time{}, fmt.Errorf("usage occurred_at missing or invalid")
	}
	return time.Unix(createdAt, 0).UTC(), nil
}

type EntryBuildInput struct {
	Raw         store.RawConsumeLog
	Mapping     *store.PlatformKeyMapping
	Source      string
	Catalog     []types.ModelInfo
	AllowedIDs  []uuid.UUID
	Settings    types.AuditSettings
	Member      *types.Member
	PlatformKey *types.PlatformKey
}

func BuildCallSettledEntry(input EntryBuildInput) (types.UsageLedgerEntry, error) {
	modelName := ResolveConsumeModel(input.Raw)

	occurredAt, err := OccurredAtFromPayload(input.Raw.CreatedAt)
	if err != nil {
		return types.UsageLedgerEntry{}, err
	}

	var memberID *uuid.UUID
	if input.Mapping.MemberID != nil {
		memberID = input.Mapping.MemberID
	}

	detail := buildCallDetail(input, memberID, modelName)

	scope := ""
	if input.PlatformKey != nil {
		scope = input.PlatformKey.Scope
	}
	if scope == "" {
		return types.UsageLedgerEntry{}, fmt.Errorf("platform key scope required")
	}

	return types.UsageLedgerEntry{
		ID:               uuid.Must(uuid.NewV7()),
		EventType:        types.EventTypeCallSettled,
		IdempotencyKey:   NewAPIIdempotencyKey(input.Raw.ID),
		Amount:           input.Raw.Quota, // direct pass-through, no conversion
		DepartmentID:     input.Mapping.DepartmentID,
		MemberID:         memberID,
		ProjectID:        input.Mapping.ProjectID,
		PlatformKeyID:    input.Mapping.PlatformKeyID,
		PlatformKeyScope: scope,
		Source:           input.Source,
		OccurredAt:       occurredAt,
		Model:            modelName,
		InputTokens:      input.Raw.PromptTokens,
		OutputTokens:     input.Raw.CompletionTokens,
		CallDetail:       detail,
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func modelsFromInput(input EntryBuildInput) []types.ModelInfo {
	return input.Catalog
}

func buildCallDetail(input EntryBuildInput, memberID *uuid.UUID, modelName string) types.UsageCallDetail {
	detail := types.UsageCallDetail{
		Provider:  resolveProvider(modelName, modelsFromInput(input)),
		Status:    types.CallStatusSuccess,
		LatencyMs: float64(input.Raw.UseTime),
	}
	if memberID != nil && input.Member != nil {
		detail.CallerType = types.CallerTypeMember
		detail.CallerID = memberID.String()
		detail.Caller = input.Member.Name
	} else {
		detail.CallerType = types.CallerTypePlatformKey
		detail.CallerID = input.Mapping.PlatformKeyID.String()
		if input.PlatformKey != nil {
			detail.Caller = input.PlatformKey.Name
		}
	}
	if input.Settings.ContentRetentionEnabled && input.Raw.Content != "" {
		detail.PreviewSnippet = TruncatePreview(input.Raw.Content)
	}
	return detail
}

func resolveProvider(modelName string, models []types.ModelInfo) string {
	for _, model := range models {
		if model.Type == modelName {
			return model.Provider
		}
	}
	return ""
}
