// Package catalogsync implements an HTTP client for the platform Catalog sync API.
package catalogsync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL   string
	SyncToken string // cst_ prefixed token for authenticated endpoints (pricing)
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchVersions returns the remote catalog version from GET /api/platform/sync/versions.
func (c *Client) FetchVersions(ctx context.Context) (*CatalogVersions, error) {
	var v CatalogVersions
	if err := c.doGet(ctx, "/api/platform/sync/versions", false, &v); err != nil {
		return nil, fmt.Errorf("catalog fetch versions: %w", err)
	}
	return &v, nil
}

// FetchModels fetches the models catalog (public, no auth required).
func (c *Client) FetchModels(ctx context.Context) (*CatalogResponse[CatalogModel], error) {
	var resp CatalogResponse[CatalogModel]
	if err := c.doGet(ctx, "/api/platform/sync/catalog/models", false, &resp); err != nil {
		return nil, fmt.Errorf("catalog fetch models: %w", err)
	}
	return &resp, nil
}

// FetchPricing fetches the global pricing catalog (requires sync token).
func (c *Client) FetchPricing(ctx context.Context) (*CatalogResponse[CatalogPricing], error) {
	var resp CatalogResponse[CatalogPricing]
	if err := c.doGet(ctx, "/api/platform/sync/catalog/pricing", true, &resp); err != nil {
		return nil, fmt.Errorf("catalog fetch pricing: %w", err)
	}
	return &resp, nil
}

// FetchDiscounts fetches per-company discount coefficients (requires sync token).
func (c *Client) FetchDiscounts(ctx context.Context) (*CatalogResponse[CatalogDiscount], error) {
	var resp CatalogResponse[CatalogDiscount]
	if err := c.doGet(ctx, "/api/platform/sync/catalog/discounts", true, &resp); err != nil {
		return nil, fmt.Errorf("catalog fetch discounts: %w", err)
	}
	return &resp, nil
}

// FetchCurrencies fetches the currencies catalog (public, no auth required).
func (c *Client) FetchCurrencies(ctx context.Context) (*CatalogResponse[CatalogCurrency], error) {
	var resp CatalogResponse[CatalogCurrency]
	if err := c.doGet(ctx, "/api/platform/sync/catalog/currencies", false, &resp); err != nil {
		return nil, fmt.Errorf("catalog fetch currencies: %w", err)
	}
	return &resp, nil
}

// FetchWalletLots fetches active lots + wallet balance for the company (requires sync token).
func (c *Client) FetchWalletLots(ctx context.Context) (*CatalogLotsResponse, error) {
	var resp CatalogLotsResponse
	if err := c.doGet(ctx, "/api/platform/sync/catalog/wallet_lots", true, &resp); err != nil {
		return nil, fmt.Errorf("catalog fetch wallet_lots: %w", err)
	}
	return &resp, nil
}

func (c *Client) doGet(ctx context.Context, path string, auth bool, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(path), nil)
	if err != nil {
		return err
	}
	if auth && c.cfg.SyncToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.SyncToken)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("catalog %s: %w", path, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("catalog read %s: %w", path, err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("catalog %s: status %d body=%s", path, res.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("catalog decode %s: %w", path, err)
	}
	return nil
}

func (c *Client) url(path string) string {
	return strings.TrimRight(c.cfg.BaseURL, "/") + path
}
