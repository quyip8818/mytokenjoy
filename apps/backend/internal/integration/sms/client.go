// Package sms implements an HTTP client for the SMS system's sync API.
// It handles OAuth2 client_credentials token management with caching.
package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
}

type CatalogChannel struct {
	Name     string            `json:"name"`
	Type     int               `json:"type"`
	BaseURL  string            `json:"baseUrl"`
	Key      string            `json:"key"`
	Models   []string          `json:"models"`
	Group    string            `json:"group"`
	Priority int               `json:"priority"`
	Settings map[string]string `json:"settings,omitempty"`
}

type CatalogModel struct {
	ModelID     string  `json:"modelId"`
	DisplayName string  `json:"displayName"`
	Provider    string  `json:"provider"`
	CallType    string  `json:"callType"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
}

type Catalog struct {
	Channels []CatalogChannel `json:"channels"`
	Models   []CatalogModel   `json:"models"`
	SyncedAt time.Time        `json:"syncedAt"`
}

type Client struct {
	cfg        Config
	httpClient *http.Client

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) FetchCatalog(ctx context.Context) (*Catalog, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("sms oauth token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/api/sync/catalog"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sms catalog request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read sms catalog: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sms catalog: status %d body=%s", res.StatusCode, string(body))
	}

	var catalog Catalog
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("decode sms catalog: %w", err)
	}
	return &catalog, nil
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return cached token if still valid (with 30s buffer)
	if c.accessToken != "" && time.Now().Add(30*time.Second).Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	// Request new token
	payload := fmt.Sprintf(`{"grant_type":"client_credentials","client_id":%q,"client_secret":%q}`,
		c.cfg.ClientID, c.cfg.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/oauth/token"),
		strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sms token request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("sms token: status %d body=%s", res.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode sms token: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return c.accessToken, nil
}

func (c *Client) url(path string) string {
	return strings.TrimRight(c.cfg.BaseURL, "/") + path
}
