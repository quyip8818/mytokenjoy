package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (c *Client) ensureToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	body := map[string]string{
		"app_id":     c.credential.AppID,
		"app_secret": c.credential.AppSecret,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/open-apis/auth/v3/tenant_access_token/internal",
		bytes.NewReader(raw),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("feishu token request failed: status %d", res.StatusCode)
	}

	var tokenResp struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.Unmarshal(payload, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Code != 0 {
		if tokenResp.Msg == "" {
			tokenResp.Msg = "authentication failed"
		}
		return "", fmt.Errorf("%s", tokenResp.Msg)
	}
	if tokenResp.TenantAccessToken == "" {
		return "", fmt.Errorf("empty company access token")
	}
	c.accessToken = tokenResp.TenantAccessToken
	ttl := tokenResp.Expire
	if ttl <= 60 {
		ttl = 7200
	}
	c.tokenExpiry = time.Now().Add(time.Duration(ttl-60) * time.Second)
	return c.accessToken, nil
}
