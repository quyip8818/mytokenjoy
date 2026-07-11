package usage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
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
	AllowedIDs  []int64
	Settings    types.AuditSettings
	Member      *types.Member
	PlatformKey *types.PlatformKey
}

func BuildCallSettledEntry(input EntryBuildInput) (types.UsageLedgerEntry, error) {
	modelName := ResolveConsumeModel(input.Raw)
	cost := CostFromLog(input.Raw.Quota, modelName, input.Catalog, input.AllowedIDs)

	occurredAt, err := OccurredAtFromPayload(input.Raw.CreatedAt)
	if err != nil {
		return types.UsageLedgerEntry{}, err
	}

	var memberID *string
	if input.Mapping.MemberID != nil {
		memberID = input.Mapping.MemberID
	}

	detail := buildCallDetail(input, memberID, modelName)

	return types.UsageLedgerEntry{
		ID:             fmt.Sprintf("ul-%d", time.Now().UnixNano()),
		EventType:      types.EventTypeCallSettled,
		IdempotencyKey: NewAPIIdempotencyKey(input.Raw.ID),
		Amount:         cost,
		DepartmentID:   input.Mapping.DepartmentID,
		MemberID:       memberID,
		BudgetGroupID:  input.Mapping.BudgetGroupID,
		PlatformKeyID:  input.Mapping.PlatformKeyID,
		Source:         input.Source,
		OccurredAt:     occurredAt,
		Model:          modelName,
		InputTokens:    input.Raw.PromptTokens,
		OutputTokens:   input.Raw.CompletionTokens,
		CallDetail:     detail,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

func modelsFromInput(input EntryBuildInput) []types.ModelInfo {
	return input.Catalog
}

func buildCallDetail(input EntryBuildInput, memberID *string, modelName string) types.UsageCallDetail {
	detail := types.UsageCallDetail{
		Provider:  resolveProvider(modelName, modelsFromInput(input)),
		Status:    types.CallStatusSuccess,
		LatencyMs: float64(input.Raw.UseTime),
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

func LoadEntryBuildInput(ctx context.Context, deps EntryBuildReader, mapping *store.PlatformKeyMapping, raw store.RawConsumeLog, source string) (EntryBuildInput, error) {
	modelName := ResolveConsumeModel(raw)
	catalog, err := deps.Models().Models(ctx)
	if err != nil {
		return EntryBuildInput{}, err
	}
	settings, err := deps.Audit().Settings(ctx)
	if err != nil {
		return EntryBuildInput{}, err
	}
	platformKey, err := deps.Keys().PlatformKeyByID(ctx, mapping.PlatformKeyID)
	if err != nil {
		return EntryBuildInput{}, err
	}
	allowedIDs := resolveBillingAllowedIDs(ctx, deps, mapping, platformKey)
	input := EntryBuildInput{
		Raw: raw, Mapping: mapping, Source: source,
		Catalog: catalog, AllowedIDs: allowedIDs, Settings: settings,
		PlatformKey: platformKey,
	}
	if mapping.MemberID != nil {
		member, err := deps.Org().MemberByID(ctx, *mapping.MemberID)
		if err != nil {
			return EntryBuildInput{}, err
		}
		input.Member = member
	}
	_ = modelName
	return input, nil
}

func resolveBillingAllowedIDs(ctx context.Context, deps EntryBuildReader, mapping *store.PlatformKeyMapping, platformKey *types.PlatformKey) []int64 {
	if platformKey == nil {
		return nil
	}
	keyIDs := append([]int64{}, platformKey.ModelWhitelist...)
	departments, err := common.LoadDepartments(ctx, deps.Org().Nodes())
	if err != nil {
		return keyIDs
	}
	rules, err := common.LoadRoutingRules(ctx, deps.Org().Nodes(), deps.Models().Allowlist())
	if err != nil {
		return keyIDs
	}
	catalog, err := deps.Models().Models(ctx)
	if err != nil {
		return keyIDs
	}
	deptAllowed := common.ResolveDeptAllowedModelIDs(mapping.DepartmentID, departments, rules, catalog)
	return newapiunits.EffectiveWhitelistIDs(keyIDs, deptAllowed)
}
