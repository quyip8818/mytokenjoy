package newapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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

// ListPricingModels fetches the full model list from NewAPI /api/pricing (public endpoint, no auth needed),
// then filters against /api/models/ (model metadata) to exclude models deleted from model management.
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

	// Cross-reference with model metadata to exclude deleted models.
	registered, err := c.listRegisteredModelNames(ctx)
	if err != nil {
		// ponytail: best-effort — if metadata endpoint fails, return all pricing models.
		return resp.Data, nil
	}

	filtered := make([]adminport.PricingModel, 0, len(resp.Data))
	for _, pm := range resp.Data {
		if registered[pm.ModelName] {
			filtered = append(filtered, pm)
		}
	}
	return filtered, nil
}

// listRegisteredModelNames fetches all model names from /api/models/ (paginated, admin endpoint).
// Returns a set of model_name strings that exist in model management.
func (c *Client) listRegisteredModelNames(ctx context.Context) (map[string]bool, error) {
	names := make(map[string]bool)
	page := 1
	for {
		url := c.baseURL + "/api/models/?p=" + strconv.Itoa(page) + "&page_size=100"
		data, err := c.fetchModelsPage(ctx, url)
		if err != nil {
			return nil, err
		}
		for _, item := range data.Items {
			names[item.ModelName] = true
		}
		if page*100 >= data.Total {
			break
		}
		page++
	}
	return names, nil
}

type modelsPageResp struct {
	Items []struct {
		ModelName string `json:"model_name"`
	} `json:"items"`
	Total int `json:"total"`
}

func (c *Client) fetchModelsPage(ctx context.Context, url string) (modelsPageResp, error) {
	var page modelsPageResp
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return page, err
	}
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	req.Header.Set("New-Api-User", "1")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return page, fmt.Errorf("fetch models page: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return page, fmt.Errorf("fetch models page read: %w", err)
	}
	var wrapper struct {
		Success bool           `json:"success"`
		Data    modelsPageResp `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return page, fmt.Errorf("fetch models page decode: %w", err)
	}
	return wrapper.Data, nil
}
