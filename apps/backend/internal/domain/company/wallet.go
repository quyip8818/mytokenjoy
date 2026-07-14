package company

import (
	"context"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
)

type WalletService interface {
	AvailableNewAPIUnits(ctx context.Context, walletUserID int64) (int64, error)
	// FreshNewAPIUnits bypasses cache, refreshes the entry, and returns authoritative quota.
	FreshNewAPIUnits(ctx context.Context, walletUserID int64) (int64, error)
	InvalidateNewAPIUnits(walletUserID int64)
}

type NewAPIWalletReader interface {
	GetUserQuota(ctx context.Context, userID int64) (int64, error)
}

type walletService struct {
	reader   NewAPIWalletReader
	cacheTTL time.Duration
	mu       sync.RWMutex
	cache    map[int64]walletCacheEntry
}

type walletCacheEntry struct {
	units     int64
	expiresAt time.Time
}

func NewWalletService(cfg config.Config, reader NewAPIWalletReader) WalletService {
	if reader == nil {
		return &noopWalletService{}
	}
	return &walletService{
		reader:   reader,
		cacheTTL: time.Duration(cfg.CompanyWalletCacheTTLSec) * time.Second,
		cache:    make(map[int64]walletCacheEntry),
	}
}

type noopWalletService struct{}

func (n *noopWalletService) AvailableNewAPIUnits(ctx context.Context, walletUserID int64) (int64, error) {
	return 0, domain.ServiceUnavailable("wallet service unavailable")
}

func (n *noopWalletService) FreshNewAPIUnits(ctx context.Context, walletUserID int64) (int64, error) {
	return n.AvailableNewAPIUnits(ctx, walletUserID)
}

func (n *noopWalletService) InvalidateNewAPIUnits(walletUserID int64) {}

func (s *walletService) AvailableNewAPIUnits(ctx context.Context, walletUserID int64) (int64, error) {
	if walletUserID <= 0 {
		return 0, domain.NewDomainError(400, "wallet account not configured")
	}
	now := time.Now()
	s.mu.RLock()
	if entry, ok := s.cache[walletUserID]; ok && now.Before(entry.expiresAt) {
		s.mu.RUnlock()
		return entry.units, nil
	}
	s.mu.RUnlock()
	return s.fetchAndCache(ctx, walletUserID, now)
}

func (s *walletService) FreshNewAPIUnits(ctx context.Context, walletUserID int64) (int64, error) {
	if walletUserID <= 0 {
		return 0, domain.NewDomainError(400, "wallet account not configured")
	}
	return s.fetchAndCache(ctx, walletUserID, time.Now())
}

func (s *walletService) fetchAndCache(ctx context.Context, walletUserID int64, now time.Time) (int64, error) {
	units, err := s.reader.GetUserQuota(ctx, walletUserID)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	s.cache[walletUserID] = walletCacheEntry{units: units, expiresAt: now.Add(s.cacheTTL)}
	s.mu.Unlock()
	return units, nil
}

func (s *walletService) InvalidateNewAPIUnits(walletUserID int64) {
	if walletUserID <= 0 {
		return
	}
	s.mu.Lock()
	delete(s.cache, walletUserID)
	s.mu.Unlock()
}
