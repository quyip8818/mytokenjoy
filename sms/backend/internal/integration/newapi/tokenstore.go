package newapi

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenStore 从 newapi 数据库读取 admin 用户的 PAT，内存缓存 + 401 时刷新。
type TokenStore struct {
	pool        *pgxpool.Pool
	adminUserID int

	mu    sync.RWMutex
	token string
}

func NewTokenStore(pool *pgxpool.Pool, adminUserID int) *TokenStore {
	return &TokenStore{pool: pool, adminUserID: adminUserID}
}

// Get 返回缓存的 PAT，首次调用自动从 DB fetch。
func (ts *TokenStore) Get(ctx context.Context) (string, error) {
	ts.mu.RLock()
	t := ts.token
	ts.mu.RUnlock()
	if t != "" {
		return t, nil
	}
	return ts.Refresh(ctx)
}

// Refresh 强制从 DB 重新读取 PAT 并更新缓存。
func (ts *TokenStore) Refresh(ctx context.Context) (string, error) {
	if ts.pool == nil {
		return "", fmt.Errorf("newapi token store: no database connection")
	}
	var token string
	err := ts.pool.QueryRow(ctx,
		`SELECT access_token FROM users WHERE id = $1`, ts.adminUserID,
	).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("read newapi admin PAT: %w", err)
	}
	if token == "" {
		return "", fmt.Errorf("newapi admin user %d has empty access_token", ts.adminUserID)
	}
	ts.mu.Lock()
	ts.token = token
	ts.mu.Unlock()
	return token, nil
}
