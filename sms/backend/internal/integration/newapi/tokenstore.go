package newapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenStore 通过登录 NewAPI 获取 session JWT，内存缓存 + 401 时刷新。
type TokenStore struct {
	pool        *pgxpool.Pool
	adminUserID int
	baseURL     string

	mu    sync.RWMutex
	token string
}

func NewTokenStore(pool *pgxpool.Pool, adminUserID int) *TokenStore {
	return &TokenStore{pool: pool, adminUserID: adminUserID}
}

func (ts *TokenStore) SetBaseURL(baseURL string) {
	ts.baseURL = baseURL
}

func (ts *TokenStore) Get(ctx context.Context) (string, error) {
	ts.mu.RLock()
	t := ts.token
	ts.mu.RUnlock()
	if t != "" {
		return t, nil
	}
	return ts.Refresh(ctx)
}

func (ts *TokenStore) Refresh(ctx context.Context) (string, error) {
	if ts.baseURL == "" {
		return "", fmt.Errorf("newapi token store: baseURL not set")
	}

	username := "admin"
	password := "admin123"
	if ts.pool != nil {
		_ = ts.pool.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, ts.adminUserID).Scan(&username)
	}

	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, ts.baseURL+"/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login to newapi: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse login response: %w", err)
	}
	if !result.Success || result.Data.AccessToken == "" {
		return "", fmt.Errorf("newapi login failed: %s", result.Message)
	}

	ts.mu.Lock()
	ts.token = result.Data.AccessToken
	ts.mu.Unlock()
	return result.Data.AccessToken, nil
}
