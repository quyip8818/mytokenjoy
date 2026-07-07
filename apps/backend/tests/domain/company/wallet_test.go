package company_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
)

// stubQuotaClient implements the GetUserQuota interface for testing.
type stubQuotaClient struct {
	mu        sync.Mutex
	quota     int64
	err       error
	callCount int
}

func (s *stubQuotaClient) GetUserQuota(_ context.Context, _ int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callCount++
	return s.quota, s.err
}

func TestWalletServiceAvailableQuota(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 50000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	quota, err := svc.AvailableQuota(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quota != 50000 {
		t.Errorf("expected 50000, got %d", quota)
	}
}

func TestWalletServiceCachesResult(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 10000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	// First call
	svc.AvailableQuota(context.Background(), 1)
	// Second call should use cache
	svc.AvailableQuota(context.Background(), 1)

	if client.callCount != 1 {
		t.Errorf("expected 1 backend call (cached), got %d", client.callCount)
	}
}

func TestWalletServiceInvalidWalletID(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 10000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	_, err := svc.AvailableQuota(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero wallet ID")
	}

	_, err = svc.AvailableQuota(context.Background(), -1)
	if err == nil {
		t.Fatal("expected error for negative wallet ID")
	}
}

func TestWalletServiceNilClient(t *testing.T) {
	t.Parallel()
	cfg := config.Config{}
	svc := company.NewWalletService(cfg, nil)

	quota, err := svc.AvailableQuota(context.Background(), 1)
	if err != nil {
		t.Fatalf("noop wallet should not error: %v", err)
	}
	if quota != 0 {
		t.Errorf("noop wallet should return 0, got %d", quota)
	}
}

func TestWalletServiceClientError(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{err: errors.New("network failure")}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	_, err := svc.AvailableQuota(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when client fails")
	}
}

func TestWalletServiceCacheExpires(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 5000}
	// Set TTL to 0 so cache expires immediately
	cfg := config.Config{CompanyWalletCacheTTLSec: 0}
	svc := company.NewWalletService(cfg, client)

	svc.AvailableQuota(context.Background(), 1)
	time.Sleep(10 * time.Millisecond)
	svc.AvailableQuota(context.Background(), 1)

	if client.callCount < 2 {
		t.Errorf("expected at least 2 calls with expired cache, got %d", client.callCount)
	}
}

func TestIsRelayBlockedStatus(t *testing.T) {
	t.Parallel()
	if company.IsRelayBlocked("active") {
		t.Error("active should not be blocked")
	}
	if !company.IsRelayBlocked("suspended") {
		t.Error("suspended should be blocked")
	}
	if !company.IsRelayBlocked("") {
		t.Error("empty status should be blocked")
	}
}
