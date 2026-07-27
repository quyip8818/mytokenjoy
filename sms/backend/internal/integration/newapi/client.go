package newapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sms/backend/internal/domain/newapisync"
)

// Client 实现 newapisync.AdminPort，通过 /api/option/ 端点与 NewAPI 交互。
type Client struct {
	baseURL    string
	tokenStore *TokenStore
	http       *http.Client
}

func NewClient(baseURL string, tokenStore *TokenStore) *Client {
	return &Client{
		baseURL:    baseURL,
		tokenStore: tokenStore,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

// NewClientForTest creates a Client with a static token (no DB dependency). Test only.
func NewClientForTest(baseURL, staticToken string) *Client {
	ts := &TokenStore{token: staticToken}
	return &Client{
		baseURL:    baseURL,
		tokenStore: ts,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

// optionEntry 是标准版 NewAPI /api/option/ 返回的数组元素
type optionEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetOptions 读取 NewAPI 全部 options
func (c *Client) GetOptions(ctx context.Context) (map[string]string, error) {
	body, err := c.doWithRetry(ctx, http.MethodGet, "/api/option/", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []optionEntry `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse options response: %w", err)
	}
	result := make(map[string]string, len(resp.Data))
	for _, entry := range resp.Data {
		result[entry.Key] = entry.Value
	}
	return result, nil
}

// PutOption 写入单个 option key-value
func (c *Client) PutOption(ctx context.Context, key, value string) error {
	payload := map[string]string{"key": key, "value": value}
	data, _ := json.Marshal(payload)
	_, err := c.doWithRetry(ctx, http.MethodPut, "/api/option/", data)
	return err
}

// ListCurrentRatios 读取 NewAPI 当前的 ModelRatio + CompletionRatio map
func (c *Client) ListCurrentRatios(ctx context.Context) (map[string]newapisync.ModelPricing, error) {
	opts, err := c.GetOptions(ctx)
	if err != nil {
		return nil, err
	}

	modelRatioMap := parseJSONMap(opts["ModelRatio"])
	completionRatioMap := parseJSONMap(opts["CompletionRatio"])

	result := make(map[string]newapisync.ModelPricing)
	for modelID, ratio := range modelRatioMap {
		cr := completionRatioMap[modelID]
		inputPrice := ratio * 2
		outputPrice := ratio * cr * 2
		result[modelID] = newapisync.ModelPricing{
			ModelID:         modelID,
			ModelRatio:      ratio,
			CompletionRatio: cr,
			InputPrice:      inputPrice,
			OutputPrice:     outputPrice,
		}
	}
	return result, nil
}

// SyncPricing 执行 read-modify-write merge：读取当前 map → 覆盖 SMS 管理的 key → 写回
func (c *Client) SyncPricing(ctx context.Context, entries []newapisync.PricingEntry) error {
	opts, err := c.GetOptions(ctx)
	if err != nil {
		return fmt.Errorf("read current options: %w", err)
	}

	modelRatioMap := parseJSONMap(opts["ModelRatio"])
	completionRatioMap := parseJSONMap(opts["CompletionRatio"])

	for _, e := range entries {
		if e.InputPrice <= 0 {
			continue
		}
		mr := e.InputPrice / 2
		cr := e.OutputPrice / e.InputPrice
		modelRatioMap[e.ModelID] = mr
		completionRatioMap[e.ModelID] = cr
	}

	mrJSON, _ := json.Marshal(modelRatioMap)
	if err := c.PutOption(ctx, "ModelRatio", string(mrJSON)); err != nil {
		return fmt.Errorf("write ModelRatio: %w", err)
	}

	crJSON, _ := json.Marshal(completionRatioMap)
	if err := c.PutOption(ctx, "CompletionRatio", string(crJSON)); err != nil {
		return fmt.Errorf("write CompletionRatio: %w", err)
	}

	return nil
}

// UpsertModelRatio 同步单个模型的定价
func (c *Client) UpsertModelRatio(ctx context.Context, modelID string, inputPrice, outputPrice float64) error {
	return c.SyncPricing(ctx, []newapisync.PricingEntry{
		{ModelID: modelID, InputPrice: inputPrice, OutputPrice: outputPrice},
	})
}

// ListModels 返回 NewAPI 上所有模型（从 ModelRatio map 生成）
func (c *Client) ListModels(ctx context.Context) ([]newapisync.NewAPIModel, error) {
	ratios, err := c.ListCurrentRatios(ctx)
	if err != nil {
		return nil, err
	}
	models := make([]newapisync.NewAPIModel, 0, len(ratios))
	for _, p := range ratios {
		models = append(models, newapisync.NewAPIModel{
			ModelID:         p.ModelID,
			InputPrice:      p.InputPrice,
			OutputPrice:     p.OutputPrice,
			ModelRatio:      p.ModelRatio,
			CompletionRatio: p.CompletionRatio,
		})
	}
	return models, nil
}

// doWithRetry 执行 HTTP 请求，遇 401 时刷新 PAT 并 retry 一次
func (c *Client) doWithRetry(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	resp, err := c.do(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// 刷新 PAT 并重试
		if _, refreshErr := c.tokenStore.Refresh(ctx); refreshErr != nil {
			return nil, fmt.Errorf("refresh PAT after 401: %w", refreshErr)
		}
		resp, err = c.do(ctx, method, path, body)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("newapi %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func (c *Client) do(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	token, err := c.tokenStore.Get(ctx)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return c.http.Do(req)
}

func parseJSONMap(s string) map[string]float64 {
	m := make(map[string]float64)
	if s == "" {
		return m
	}
	_ = json.Unmarshal([]byte(s), &m)
	return m
}

// channelListResponse 是 GET /api/channel/ 的响应结构
type channelListResponse struct {
	Data struct {
		Items []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Type     int    `json:"type"`
			Status   int    `json:"status"`
			Models   string `json:"models"`
			BaseURL  string `json:"base_url"`
			Priority int    `json:"priority"`
			Weight   int    `json:"weight"`
		} `json:"items"`
	} `json:"data"`
}

// ListChannels 从 NewAPI 拉取所有渠道
func (c *Client) ListChannels(ctx context.Context) ([]newapisync.Channel, error) {
	body, err := c.doWithRetry(ctx, http.MethodGet, "/api/channel/?p=0&page_size=100", nil)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}

	var resp channelListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse channel response: %w", err)
	}

	channels := make([]newapisync.Channel, 0, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		channels = append(channels, newapisync.Channel{
			ID:       item.ID,
			Name:     item.Name,
			Type:     item.Type,
			Status:   item.Status,
			Models:   item.Models,
			BaseURL:  item.BaseURL,
			Priority: item.Priority,
			Weight:   item.Weight,
		})
	}
	return channels, nil
}
