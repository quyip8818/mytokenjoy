package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// PricingResponse is the expected response from the platform GET /api/v1/pricing/latest.
type PricingResponse struct {
	Version             string `json:"version"`
	ModelRatioJSON      string `json:"model_ratio"`
	CompletionRatioJSON string `json:"completion_ratio"`
}

// Client calls the official management platform APIs.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetLatestPricing fetches the latest pricing snapshot from the platform.
func (c *Client) GetLatestPricing(ctx context.Context) (*PricingResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/pricing/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("platform pricing request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read platform pricing response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("platform pricing: status %d body=%s", res.StatusCode, string(body))
	}

	var resp PricingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode platform pricing: %w", err)
	}
	return &resp, nil
}
