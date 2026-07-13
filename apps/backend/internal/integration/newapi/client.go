package newapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AdminClient interface {
	CreateToken(ctx context.Context, req CreateTokenRequest) (Token, error)
	UpdateToken(ctx context.Context, req UpdateTokenRequest) (Token, error)
	GetToken(ctx context.Context, tokenID int64) (Token, error)
	GetTokenKey(ctx context.Context, tokenID int64) (string, error)
	DeleteToken(ctx context.Context, tokenID int64) error
	RegenerateToken(ctx context.Context, tokenID int64) (Token, error)
	CreateUser(ctx context.Context, req CreateUserRequest) (User, error)
	GetUserQuota(ctx context.Context, userID int64) (int64, error)
	TopUp(ctx context.Context, req TopUpRequest) error
	UpsertChannel(ctx context.Context, req UpsertChannelRequest) (Channel, error)
	RebuildAbilities(ctx context.Context) error
	EnsureGroup(ctx context.Context, group, displayName string) error
}

type Client struct {
	baseURL     string
	adminToken  string
	adminUserID int64
	httpClient  *http.Client
}

func NewClient(baseURL, adminToken string, adminUserID int64) *Client {
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		adminToken:  adminToken,
		adminUserID: adminUserID,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

type apiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	if c.adminUserID > 0 {
		req.Header.Set("New-Api-User", strconv.FormatInt(c.adminUserID, 10))
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("newapi %s %s: status %d body=%s", method, path, res.StatusCode, string(payload))
	}
	var wrapped apiResponse
	if err := json.Unmarshal(payload, &wrapped); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !wrapped.Success {
		if wrapped.Message == "" {
			wrapped.Message = "request failed"
		}
		return fmt.Errorf("newapi: %s", wrapped.Message)
	}
	if out == nil || len(wrapped.Data) == 0 || string(wrapped.Data) == "null" {
		return nil
	}
	return json.Unmarshal(wrapped.Data, out)
}
