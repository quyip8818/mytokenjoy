// Package sms implements an HTTP client for the SMS system's sync API.
// It handles OAuth2 client_credentials token management with caching and 401 auto-retry.
package sms

import (
	"context"
	"encoding/json"
	"errors"
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
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchVersions returns the remote partition versions from GET /api/sync/versions.
func (c *Client) FetchVersions(ctx context.Context) (*PartitionVersions, error) {
	var v PartitionVersions
	if err := c.doWithRetry(ctx, "/api/sync/versions", &v); err != nil {
		return nil, fmt.Errorf("sms fetch versions: %w", err)
	}
	return &v, nil
}

// FetchChannels fetches the channels catalog partition.
func (c *Client) FetchChannels(ctx context.Context) (*PartitionResponse[CatalogChannel], error) {
	var resp PartitionResponse[CatalogChannel]
	if err := c.doWithRetry(ctx, "/api/sync/catalog/channels", &resp); err != nil {
		return nil, fmt.Errorf("sms fetch channels: %w", err)
	}
	return &resp, nil
}

// FetchModels fetches the models catalog partition.
func (c *Client) FetchModels(ctx context.Context) (*PartitionResponse[CatalogModel], error) {
	var resp PartitionResponse[CatalogModel]
	if err := c.doWithRetry(ctx, "/api/sync/catalog/models", &resp); err != nil {
		return nil, fmt.Errorf("sms fetch models: %w", err)
	}
	return &resp, nil
}

// doWithRetry performs an authenticated GET. On 401 it invalidates the token and retries once.
func (c *Client) doWithRetry(ctx context.Context, path string, out any) error {
	err := c.doGet(ctx, path, out)
	if err == nil {
		return nil
	}
	if !isUnauthorized(err) {
		return err
	}
	// Invalidate cached token and retry once.
	c.invalidateToken()
	return c.doGet(ctx, path, out)
}

func (c *Client) doGet(ctx context.Context, path string, out any) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("oauth token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(path), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if res.StatusCode == http.StatusUnauthorized {
		return &unauthorizedError{path: path, body: string(body)}
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("sms %s: status %d body=%s", path, res.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

// invalidateToken clears the cached access token, forcing a re-fetch on next call.
func (c *Client) invalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = ""
	c.tokenExpiry = time.Time{}
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Add(30*time.Second).Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

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

// unauthorizedError signals a 401 response for retry logic.
type unauthorizedError struct {
	path string
	body string
}

func (e *unauthorizedError) Error() string {
	return fmt.Sprintf("sms %s: 401 unauthorized body=%s", e.path, e.body)
}

func isUnauthorized(err error) bool {
	var ue *unauthorizedError
	return errors.As(err, &ue)
}
