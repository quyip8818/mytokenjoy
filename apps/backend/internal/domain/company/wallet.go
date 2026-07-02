package company

import (
	"context"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
)

type WalletService interface {
	AvailableQuota(ctx context.Context, walletUserID int64) (int64, error)
}

type walletService struct {
	cfg    config.Config
	client interface {
		GetUserQuota(ctx context.Context, userID int64) (int64, error)
	}
	cacheTTL time.Duration
	mu       sync.RWMutex
	cache    map[int64]walletCacheEntry
}

type walletCacheEntry struct {
	quota     int64
	expiresAt time.Time
}

func NewWalletService(cfg config.Config, client interface {
	GetUserQuota(ctx context.Context, userID int64) (int64, error)
}) WalletService {
	if client == nil {
		return &noopWalletService{}
	}
	return &walletService{
		cfg:      cfg,
		client:   client,
		cacheTTL: time.Duration(cfg.CompanyWalletCacheTTLSec) * time.Second,
		cache:    make(map[int64]walletCacheEntry),
	}
}

type noopWalletService struct{}

func (n *noopWalletService) AvailableQuota(ctx context.Context, walletUserID int64) (int64, error) {
	return 0, nil
}

func (s *walletService) AvailableQuota(ctx context.Context, walletUserID int64) (int64, error) {
	if walletUserID <= 0 {
		return 0, domain.NewDomainError(400, "wallet account not configured")
	}
	now := time.Now()
	s.mu.RLock()
	if entry, ok := s.cache[walletUserID]; ok && now.Before(entry.expiresAt) {
		s.mu.RUnlock()
		return entry.quota, nil
	}
	s.mu.RUnlock()
	quota, err := s.client.GetUserQuota(ctx, walletUserID)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	s.cache[walletUserID] = walletCacheEntry{quota: quota, expiresAt: now.Add(s.cacheTTL)}
	s.mu.Unlock()
	return quota, nil
}
