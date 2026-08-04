package newapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain/adminport"
)

func (c *Client) ListModelPricing(ctx context.Context) ([]adminport.ModelPricing, error) {
	// Read from global option (ModelRatio + CompletionRatio + CacheRatio maps) — same source
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
	cacheMap := map[string]float64{}
	if raw := byKey["CacheRatio"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &cacheMap)
	}

	out := make([]adminport.ModelPricing, 0, len(mrMap))
	for modelName, modelRatio := range mrMap {
		out = append(out, adminport.ModelPricing{
			ModelName:       modelName,
			ModelRatio:      modelRatio,
			CompletionRatio: crMap[modelName],
			CacheRatio:      cacheMap[modelName],
		})
	}
	return out, nil
}

// ListPricingModels fetches the full model list from NewAPI /api/pricing (public endpoint, no auth needed).
func (c *Client) ListPricingModels(ctx context.Context) ([]adminport.PricingModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/pricing", nil)
	if err != nil {
		return nil, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list pricing models: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("list pricing models read body: %w", err)
	}
	var resp struct {
		Success bool                     `json:"success"`
		Data    []adminport.PricingModel `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("list pricing models decode: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("list pricing models: upstream returned success=false")
	}
	return resp.Data, nil
}
