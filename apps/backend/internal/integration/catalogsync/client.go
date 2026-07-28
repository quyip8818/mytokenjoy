// Package catalogsync implements an HTTP client for the platform Catalog sync API.
// No authentication — the Catalog API is public read-only.
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
	BaseURL string
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
	if err := c.doGet(ctx, "/api/platform/sync/versions", &v); err != nil {
		return nil, fmt.Errorf("catalog fetch versions: %w", err)
	}
	return &v, nil
}

// FetchModels fetches the models catalog.
func (c *Client) FetchModels(ctx context.Context) (*CatalogResponse[CatalogModel], error) {
	var resp CatalogResponse[CatalogModel]
	if err := c.doGet(ctx, "/api/platform/sync/catalog/models", &resp); err != nil {
		return nil, fmt.Errorf("catalog fetch models: %w", err)
	}
	return &resp, nil
}

func (c *Client) doGet(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(path), nil)
	if err != nil {
		return err
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
