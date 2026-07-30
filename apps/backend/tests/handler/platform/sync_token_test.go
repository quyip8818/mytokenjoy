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
		CompanyID string `json:"companyId"`
		SyncToken string `json:"syncToken"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.CompanyID == "" {
		t.Fatal("T1: companyId is empty")
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
		"name": "Replay Corp", "idempotencyKey": idemKey,
	})

	// First call — should succeed
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("T3 first call: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Second call (immediate) — should 409
	body, _ = json.Marshal(map[string]string{
		"name": "Replay Corp", "idempotencyKey": idemKey,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("T3 replay: expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- T4: Wrong registration secret returns 403 ---

func TestRegisterLocalBadSecret(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	body, _ := json.Marshal(map[string]string{
		"name": "Bad Secret Corp", "idempotencyKey": "key1",
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

// --- T10: CatalogPricing returns contract override for the token's company ---

func TestCatalogPricingContractOverride(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)
	platformCookie := saas.LoginPlatform(t, router)

	// Register a company and get token + companyID
	companyID, syncToken := registerAndGetBoth(t, router)

	// Set global pricing
	setGlobalPrice(t, router, platformCookie, "gpt-4o", 2.0, 8.0)

	// Set contract pricing for this company (lower price)
	setContractPrice(t, router, platformCookie, companyID, "gpt-4o", 1.0, 4.0)

	// Fetch pricing with sync token
	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/pricing", nil)
	req.Header.Set("Authorization", "Bearer "+syncToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("T10: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []struct {
			ModelType   string  `json:"modelType"`
			InputPrice  float64 `json:"inputPrice"`
			OutputPrice float64 `json:"outputPrice"`
			IsContract  bool    `json:"isContract"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	found := false
	for _, p := range resp.Data {
		if p.ModelType == "gpt-4o" {
			found = true
			if p.InputPrice != 1.0 {
				t.Errorf("T10: expected contract inputPrice 1.0, got %f", p.InputPrice)
			}
			if p.OutputPrice != 4.0 {
				t.Errorf("T10: expected contract outputPrice 4.0, got %f", p.OutputPrice)
			}
			if !p.IsContract {
				t.Error("T10: expected isContract=true for overridden price")
			}
		}
	}
	if !found {
		t.Fatal("T10: gpt-4o not found in pricing response")
	}
}

// --- T11: Company without contract prices gets global prices only ---

func TestCatalogPricingGlobalOnly(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)
	platformCookie := saas.LoginPlatform(t, router)

	// Register a company (no contract pricing set for it)
	_, syncToken := registerAndGetBoth(t, router)

	// Set global pricing
	setGlobalPrice(t, router, platformCookie, "claude-3", 3.0, 15.0)

	// Fetch pricing — should only see global price with isContract=false
	req := httptest.NewRequest(http.MethodGet, "/api/platform/sync/catalog/pricing", nil)
	req.Header.Set("Authorization", "Bearer "+syncToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("T11: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []struct {
			ModelType   string  `json:"modelType"`
			InputPrice  float64 `json:"inputPrice"`
			OutputPrice float64 `json:"outputPrice"`
			IsContract  bool    `json:"isContract"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	for _, p := range resp.Data {
		if p.ModelType == "claude-3" {
			if p.IsContract {
				t.Error("T11: expected isContract=false for global-only price")
			}
			if p.InputPrice != 3.0 {
				t.Errorf("T11: expected global inputPrice 3.0, got %f", p.InputPrice)
			}
			return
		}
	}
	t.Fatal("T11: claude-3 not found in pricing response")
}

// --- Helpers ---

func registerAndGetToken(t *testing.T, router http.Handler) string {
	t.Helper()
	_, token := registerAndGetBoth(t, router)
	return token
}

func registerAndGetBoth(t *testing.T, router http.Handler) (companyID string, syncToken string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"name":           "Test Corp " + uuid.Must(uuid.NewV7()).String()[:8],
		"idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("registerAndGetBoth: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result struct {
		CompanyID string `json:"companyId"`
		SyncToken string `json:"syncToken"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result.CompanyID, result.SyncToken
}

func setGlobalPrice(t *testing.T, router http.Handler, cookie, modelType string, input, output float64) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"modelType": modelType, "inputPrice": input, "outputPrice": output,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/platform/pricing", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setGlobalPrice: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func setContractPrice(t *testing.T, router http.Handler, cookie, companyID, modelType string, input, output float64) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"modelType": modelType, "inputPrice": input, "outputPrice": output,
	})
	url := "/api/platform/companies/" + companyID + "/pricing"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setContractPrice: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
