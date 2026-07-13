package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

const (
	RootDepartmentExternalID = "0"
	defaultPageSize          = 50
)

type Department struct {
	ExternalID       string
	Name             string
	ParentExternalID string
	LeaderUserID     string
}

type Member struct {
	ExternalID           string
	Name                 string
	Email                string
	Mobile               string
	DepartmentExternalID string
	EmployeeNo           string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseURL    string
	credential types.FeishuCredential
	httpClient HTTPClient

	tokenMu     sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewClient(baseURL string, credential types.FeishuCredential, httpClient HTTPClient) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		credential: credential,
		httpClient: httpClient,
	}
}

type apiResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.ensureToken(ctx)
	return err
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, out)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, out any) error {
	token, err := c.ensureToken(ctx)
	if err != nil {
		return err
	}
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
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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
		return fmt.Errorf("feishu request failed: status %d", res.StatusCode)
	}
	var wrapped apiResponse
	if err := json.Unmarshal(payload, &wrapped); err != nil {
		return err
	}
	if wrapped.Code != 0 {
		if wrapped.Msg == "" {
			wrapped.Msg = "request failed"
		}
		return fmt.Errorf("%s", wrapped.Msg)
	}
	if out == nil || len(wrapped.Data) == 0 {
		return nil
	}
	return json.Unmarshal(wrapped.Data, out)
}
