package company

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/httputil"
)

type Gate struct {
	cfg config.Config
}

func NewGate(cfg config.Config) *Gate {
	return &Gate{cfg: cfg}
}

func (g *Gate) IsSuspended(ctx context.Context) bool {
	companyCtx, ok := FromContext(ctx)
	if !ok {
		return false
	}
	return companyCtx.Status == "suspended"
}

func (g *Gate) IsReadOnlyAllowed(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func CompanyReadOnlyMiddleware(gate *Gate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if gate.IsSuspended(r.Context()) && !gate.IsReadOnlyAllowed(r.Method) {
				httputil.WriteStatus(w, http.StatusForbidden, "Company suspended")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type WalletService interface {
	AvailableQuota(ctx context.Context, walletAccountID int64) (int64, error)
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

func (n *noopWalletService) AvailableQuota(ctx context.Context, walletAccountID int64) (int64, error) {
	return 0, nil
}

func (s *walletService) AvailableQuota(ctx context.Context, walletAccountID int64) (int64, error) {
	if walletAccountID <= 0 {
		return 0, domain.NewDomainError(400, "wallet account not configured")
	}
	now := time.Now()
	s.mu.RLock()
	if entry, ok := s.cache[walletAccountID]; ok && now.Before(entry.expiresAt) {
		s.mu.RUnlock()
		return entry.quota, nil
	}
	s.mu.RUnlock()
	quota, err := s.client.GetUserQuota(ctx, walletAccountID)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	s.cache[walletAccountID] = walletCacheEntry{quota: quota, expiresAt: now.Add(s.cacheTTL)}
	s.mu.Unlock()
	return quota, nil
}
