package platform

import (
	"strconv"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// splitModels splits a comma-separated model list from NewAPI channel config.
func splitModels(models string) []string {
	parts := strings.Split(models, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// pricingModelToModelInfo converts a NewAPI PricingModel to a TokenJoy ModelInfo.
func pricingModelToModelInfo(pm adminport.PricingModel) types.ModelInfo {
	modelType := pm.ModelName
	displayName := inferDisplayName(pm.ModelName)
	provider := inferProvider(pm.ModelName)
	capabilities := parseTags(pm.Tags)
	maxContext := extractMaxContext(pm.Tags)

	return types.ModelInfo{
		Provider:     provider,
		Type:         modelType,
		Name:         displayName,
		Description:  pm.Description,
		Source:       "platform",
		Deprecated:   false,
		Capabilities: capabilities,
		MaxContext:   maxContext,
	}
}

func inferDisplayName(modelName string) string {
	if idx := strings.Index(modelName, "/"); idx >= 0 {
		return modelName[idx+1:]
	}
	return modelName
}

func inferProvider(modelName string) string {
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

func parseTags(tags string) []string {
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

func extractMaxContext(tags string) int {
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
