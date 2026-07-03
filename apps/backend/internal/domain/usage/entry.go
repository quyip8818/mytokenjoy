package usage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
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

func OccurredAtFromPayload(createdAt int64) time.Time {
	if createdAt <= 0 {
		return time.Now().UTC()
	}
	return time.Unix(createdAt, 0).UTC()
}

type EntryBuildInput struct {
	Payload     newapi.WebhookLogPayload
	Mapping     *store.RelayMapping
	Source      string
	Model       *types.ModelInfo
	Settings    types.AuditSettings
	Member      *types.Member
	PlatformKey *types.PlatformKey
}

func BuildCallSettledEntry(input EntryBuildInput) (types.UsageLedgerEntry, error) {
	modelName := ResolveWebhookModel(input.Payload)
	costCNY := CostCNYFromLog(input.Payload.Quota, modelName, modelsFromInput(input))

	occurredAt := OccurredAtFromPayload(input.Payload.CreatedAt)

	var memberID *string
	if input.Mapping.MemberID != nil {
		memberID = input.Mapping.MemberID
	}

	detail := buildCallDetail(input, memberID, modelName)

	return types.UsageLedgerEntry{
		ID:             fmt.Sprintf("ul-%d", time.Now().UnixNano()),
		EventType:      types.EventTypeCallSettled,
		IdempotencyKey: NewAPIIdempotencyKey(input.Payload.ID),
		AmountCNY:      costCNY,
		DepartmentID:   input.Mapping.DepartmentID,
		MemberID:       memberID,
		BudgetGroupID:  input.Mapping.BudgetGroupID,
		PlatformKeyID:  input.Mapping.PlatformKeyID,
		Source:         input.Source,
		OccurredAt:     occurredAt,
		Model:          modelName,
		InputTokens:    input.Payload.PromptTokens,
		OutputTokens:   input.Payload.CompletionTokens,
		CallDetail:     detail,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

func modelsFromInput(input EntryBuildInput) []types.ModelInfo {
	if input.Model == nil {
		return nil
	}
	return []types.ModelInfo{*input.Model}
}

func buildCallDetail(input EntryBuildInput, memberID *string, modelName string) types.UsageCallDetail {
	detail := types.UsageCallDetail{
		Provider:  resolveProvider(modelName, modelsFromInput(input)),
		Status:    types.CallStatusSuccess,
		LatencyMs: float64(input.Payload.UseTime),
	}
	if memberID != nil && input.Member != nil {
		detail.CallerType = types.CallerTypeMember
		detail.CallerID = *memberID
		detail.Caller = input.Member.Name
	} else {
		detail.CallerType = types.CallerTypePlatformKey
		detail.CallerID = input.Mapping.PlatformKeyID
		if input.PlatformKey != nil {
			detail.Caller = input.PlatformKey.Name
		}
	}
	if input.Settings.ContentRetentionEnabled && input.Payload.Input != "" {
		detail.PreviewSnippet = TruncatePreview(input.Payload.Input)
	}
	return detail
}

func resolveProvider(modelName string, models []types.ModelInfo) string {
	for _, model := range models {
		if model.Name == modelName {
			return model.Provider
		}
	}
	return ""
}

func LoadEntryBuildInput(ctx context.Context, deps EntryBuildReader, mapping *store.RelayMapping, payload newapi.WebhookLogPayload, source string) (EntryBuildInput, error) {
	modelName := ResolveWebhookModel(payload)
	model, err := deps.Models().ModelByName(ctx, modelName)
	if err != nil {
		return EntryBuildInput{}, err
	}
	settings, err := deps.Audit().Settings(ctx)
	if err != nil {
		return EntryBuildInput{}, err
	}
	input := EntryBuildInput{
		Payload:  payload,
		Mapping:  mapping,
		Source:   source,
		Model:    model,
		Settings: settings,
	}
	if mapping.MemberID != nil {
		member, err := deps.Org().MemberByID(ctx, *mapping.MemberID)
		if err != nil {
			return EntryBuildInput{}, err
		}
		input.Member = member
	}
	platformKey, err := deps.Keys().PlatformKeyByID(ctx, mapping.PlatformKeyID)
	if err != nil {
		return EntryBuildInput{}, err
	}
	input.PlatformKey = platformKey
	return input, nil
}
