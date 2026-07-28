package newapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/adminport"
)

func (c *Client) ListModelPricing(ctx context.Context) ([]adminport.ModelPricing, error) {
	// Read from global option (ModelRatio + CompletionRatio maps) — same source
	// that UpsertModelRatio writes to. The /api/pricing endpoint aggregates
	// channel-level data which may differ from the global pricing intent.
	var entries []optionEntry
	if err := c.do(ctx, "GET", "/api/option/", nil, &entries); err != nil {
		return nil, fmt.Errorf("list model pricing: %w", err)
	}
	byKey := make(map[string]string, len(entries))
	for _, e := range entries {
		byKey[e.Key] = e.Value
	}

	mrMap := map[string]float64{}
	if raw := byKey["ModelRatio"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &mrMap)
	}
	crMap := map[string]float64{}
	if raw := byKey["CompletionRatio"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &crMap)
	}

	out := make([]adminport.ModelPricing, 0, len(mrMap))
	for modelName, modelRatio := range mrMap {
		out = append(out, adminport.ModelPricing{
			ModelName:       modelName,
			ModelRatio:      modelRatio,
			CompletionRatio: crMap[modelName], // defaults to 0 if not in map
		})
	}
	return out, nil
}
