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

func TestWalletServiceAvailableNewAPIUnits(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 50000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	quota, err := svc.AvailableNewAPIUnits(context.Background(), 1)
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

	svc.AvailableNewAPIUnits(context.Background(), 1)
	svc.AvailableNewAPIUnits(context.Background(), 1)

	if client.callCount != 1 {
		t.Errorf("expected 1 backend call (cached), got %d", client.callCount)
	}
}

func TestWalletServiceInvalidWalletID(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 10000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	_, err := svc.AvailableNewAPIUnits(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero wallet ID")
	}

	_, err = svc.AvailableNewAPIUnits(context.Background(), -1)
	if err == nil {
		t.Fatal("expected error for negative wallet ID")
	}
}

func TestWalletServiceNilClient(t *testing.T) {
	t.Parallel()
	cfg := config.Config{}
	svc := company.NewWalletService(cfg, nil)

	_, err := svc.AvailableNewAPIUnits(context.Background(), 1)
	if err == nil {
		t.Fatal("noop wallet should return service unavailable error")
	}
}

func TestWalletServiceClientError(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{err: errors.New("network failure")}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	_, err := svc.AvailableNewAPIUnits(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when client fails")
	}
}

func TestWalletServiceFreshBypassesCache(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 10000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	if _, err := svc.AvailableNewAPIUnits(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	client.mu.Lock()
	client.quota = 42
	client.mu.Unlock()
	got, err := svc.FreshNewAPIUnits(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Fatalf("expected fresh 42, got %d", got)
	}
	if client.callCount != 2 {
		t.Fatalf("expected 2 backend calls, got %d", client.callCount)
	}
}

func TestWalletServiceInvalidate(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 10000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 60}
	svc := company.NewWalletService(cfg, client)

	svc.AvailableNewAPIUnits(context.Background(), 1)
	svc.InvalidateNewAPIUnits(1)
	svc.AvailableNewAPIUnits(context.Background(), 1)

	if client.callCount != 2 {
		t.Errorf("expected 2 backend calls after invalidate, got %d", client.callCount)
	}
}

func TestWalletServiceCacheExpires(t *testing.T) {
	t.Parallel()
	client := &stubQuotaClient{quota: 5000}
	cfg := config.Config{CompanyWalletCacheTTLSec: 0}
	svc := company.NewWalletService(cfg, client)

	svc.AvailableNewAPIUnits(context.Background(), 1)
	time.Sleep(10 * time.Millisecond)
	svc.AvailableNewAPIUnits(context.Background(), 1)

	if client.callCount < 2 {
		t.Errorf("expected at least 2 calls with expired cache, got %d", client.callCount)
	}
}


func TestIsGatewayBlockedStatus(t *testing.T) {
	t.Parallel()
	if company.IsGatewayBlocked("active") {
		t.Error("active should not be blocked")
	}
	if !company.IsGatewayBlocked("suspended") {
		t.Error("suspended should be blocked")
	}
	if !company.IsGatewayBlocked("") {
		t.Error("empty status should be blocked")
	}
}
