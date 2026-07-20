package newapi

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/adminport"
)

type pricingEntry struct {
	ModelName       string  `json:"model_name"`
	ModelRatio      float64 `json:"model_ratio"`
	CompletionRatio float64 `json:"completion_ratio"`
}

type pricingResponse struct {
	Data []pricingEntry `json:"data"`
}

func (c *Client) ListModelPricing(ctx context.Context) ([]adminport.ModelPricing, error) {
	var resp pricingResponse
	if err := c.do(ctx, "GET", "/api/pricing", nil, &resp); err != nil {
		return nil, fmt.Errorf("list model pricing: %w", err)
	}
	out := make([]adminport.ModelPricing, 0, len(resp.Data))
	for _, e := range resp.Data {
		out = append(out, adminport.ModelPricing{
			ModelName:       e.ModelName,
			ModelRatio:      e.ModelRatio,
			CompletionRatio: e.CompletionRatio,
		})
	}
	return out, nil
}
