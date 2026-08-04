package app

import (
	"context"
	"log/slog"
	"strings"
	"strconv"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// syncModelsFromNewAPIOnBoot pulls models from NewAPI /api/pricing and upserts them
// into the platform models table. Best-effort — failure is non-fatal.
func syncModelsFromNewAPIOnBoot(ctx context.Context, cfg config.Config, st store.Store, port adminport.Port) {
	pricingModels, err := port.ListPricingModels(ctx)
	if err != nil {
		slog.Warn("sync models from newapi on boot: fetch failed", "error", err)
		return
	}
	if len(pricingModels) == 0 {
		return
	}

	infos := make([]types.ModelInfo, 0, len(pricingModels))
	for _, pm := range pricingModels {
		infos = append(infos, pricingModelToInfo(pm))
	}

	if err := st.Models().SyncFromPlatform(ctx, cfg.TokenJoyCompanyID, infos); err != nil {
		slog.Warn("sync models from newapi on boot: store failed", "error", err)
		return
	}
	slog.Info("sync models from newapi on boot completed", "count", len(infos))
}

// pricingModelToInfo converts a NewAPI PricingModel to a TokenJoy ModelInfo.
func pricingModelToInfo(pm adminport.PricingModel) types.ModelInfo {
	return types.ModelInfo{
		Provider:     inferProviderFromName(pm.ModelName),
		Type:         pm.ModelName,
		Name:         inferDisplayNameFromModel(pm.ModelName),
		Description:  pm.Description,
		Source:       "platform",
		Deprecated:   false,
		Capabilities: parseTagsList(pm.Tags),
		MaxContext:   extractContextFromTags(pm.Tags),
	}
}

func inferDisplayNameFromModel(modelName string) string {
	if idx := strings.Index(modelName, "/"); idx >= 0 {
		return modelName[idx+1:]
	}
	return modelName
}

func inferProviderFromName(modelName string) string {
	if idx := strings.Index(modelName, "/"); idx >= 0 {
		org := strings.ToLower(modelName[:idx])
		switch {
		case strings.Contains(org, "deepseek"):
			return "deepseek"
		case strings.Contains(org, "moonshot"):
			return "moonshot"
		case strings.Contains(org, "zai") || strings.Contains(org, "glm"):
			return "zhipu"
		default:
			return org
		}
	}
	lower := strings.ToLower(modelName)
	switch {
	case strings.HasPrefix(lower, "deepseek"):
		return "deepseek"
	case strings.HasPrefix(lower, "gpt") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3"):
		return "openai"
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	default:
		return ""
	}
}

func parseTagsList(tags string) []string {
	if tags == "" {
		return []string{"chat"}
	}
	parts := strings.Split(tags, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return []string{"chat"}
	}
	return out
}

func extractContextFromTags(tags string) int {
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		multiplier := 0
		numStr := ""
		if strings.HasSuffix(tag, "M") {
			multiplier = 1_000_000
			numStr = strings.TrimSuffix(tag, "M")
		} else if strings.HasSuffix(tag, "K") {
			multiplier = 1_000
			numStr = strings.TrimSuffix(tag, "K")
		}
		if multiplier > 0 {
			if v, err := strconv.ParseFloat(numStr, 64); err == nil {
				return int(v * float64(multiplier))
			}
		}
	}
	return 128000
}
