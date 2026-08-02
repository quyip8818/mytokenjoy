//go:build testhook

package platform_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

const testRegistrationSecret = "test-local-reg-secret-32bytes!!"

// syncRouter creates a SaaS router with LocalRegistrationSecret configured.
func syncRouter(t *testing.T) http.Handler {
	t.Helper()
	mock := saas.StartNewAPIMock(t)
	application := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
		cfg.LocalRegistrationSecret = testRegistrationSecret
	}, app.WithoutWorker(), app.WithAdminPort(mock.AdminPort()))
	return application.Router
}

// --- T1: First registration creates company + issues sync token ---

func TestRegisterLocalFirstRegistration(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	body, _ := json.Marshal(map[string]string{
		"name": "Local Corp", "industry": "tech", "size": "10-50",
		"adminEmail": "admin@local-corp.test", "adminPassword": "secure-password-123",
		"idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("T1: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result struct {
		CompanyID    string `json:"companyId"`
		WalletUserID int64  `json:"walletUserId"`
		PlatformKey  string `json:"platformKey"`
		SyncToken    string `json:"syncToken"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.CompanyID == "" {
		t.Fatal("T1: companyId is empty")
	}
	if result.WalletUserID <= 0 {
		t.Fatalf("T1: walletUserId should be positive, got %d", result.WalletUserID)
	}
	if result.PlatformKey == "" {
		t.Fatal("T1: platformKey is empty")
	}
	if !strings.HasPrefix(result.SyncToken, "cst_") {
		t.Fatalf("T1: syncToken should start with cst_, got %q", result.SyncToken)
	}
	if len(result.SyncToken) != 68 {
		t.Fatalf("T1: syncToken should be 68 chars (cst_ + 64 hex), got %d", len(result.SyncToken))
	}
}

// --- T3: Same idempotencyKey within 60s returns 409 ---

func TestRegisterLocalReplayWithin60s(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)
	idemKey := uuid.Must(uuid.NewV7()).String()

	body, _ := json.Marshal(map[string]string{
		"name": "Replay Corp", "adminEmail": "replay@test.local",
		"adminPassword": "test-password-123", "idempotencyKey": idemKey,
	})

	// First call — should succeed
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("T3 first call: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Second call (immediate) — idempotent, returns 200 with same result
	body, _ = json.Marshal(map[string]string{
		"name": "Replay Corp", "adminEmail": "replay@test.local",
		"adminPassword": "test-password-123", "idempotencyKey": idemKey,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("T3 replay: expected 200 (idempotent), got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- T4: Wrong registration secret returns 403 ---

func TestRegisterLocalBadSecret(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	body, _ := json.Marshal(map[string]string{
		"name": "Bad Secret Corp", "adminEmail": "bad@test.local",
		"adminPassword": "test-password-123", "idempotencyKey": "key1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", "wrong-secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("T4: expected 403, got %d", rec.Code)
	}
}

// --- T5: Valid sync token passes middleware and returns pricing ---

func TestSyncTokenValidAccess(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	syncToken := registerAndGetToken(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/pricing", nil)
	req.Header.Set("Authorization", "Bearer "+syncToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("T5: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- T6: No Authorization header returns 401 ---

func TestSyncTokenMissingHeader(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/pricing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("T6: expected 401, got %d", rec.Code)
	}
}

// --- T7: Non-cst_ prefix returns 401 without DB lookup ---

func TestSyncTokenBadPrefix(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/pricing", nil)
	req.Header.Set("Authorization", "Bearer sk_not_a_sync_token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("T7: expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- T8: Valid format but unknown hash returns 403 ---

func TestSyncTokenUnknownHash(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	fakeToken := "cst_" + strings.Repeat("ab", 32)
	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/pricing", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("T8: expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- Helpers ---

func registerAndGetToken(t *testing.T, router http.Handler) string {
	t.Helper()
	result := registerAndGetAll(t, router)
	return result.SyncToken
}

type registerLocalTestResult struct {
	CompanyID    string `json:"companyId"`
	WalletUserID int64  `json:"walletUserId"`
	PlatformKey  string `json:"platformKey"`
	SyncToken    string `json:"syncToken"`
}

func registerAndGetAll(t *testing.T, router http.Handler) registerLocalTestResult {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"name":           "Test Corp " + uuid.Must(uuid.NewV7()).String()[:8],
		"adminEmail":     "admin-" + uuid.Must(uuid.NewV7()).String()[:8] + "@test.local",
		"adminPassword":  "test-password-123",
		"idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("registerAndGetAll: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result registerLocalTestResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}
