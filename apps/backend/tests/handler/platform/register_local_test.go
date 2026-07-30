//go:build testhook

package platform_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

// syncRouterWithStore returns both the router and the underlying store for verification.
func syncRouterWithStore(t *testing.T) (http.Handler, store.Store) {
	t.Helper()
	mock := saas.StartNewAPIMock(t)
	application := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
		cfg.LocalRegistrationSecret = testRegistrationSecret
	}, app.WithoutWorker(), app.WithAdminPort(mock.AdminPort()))
	return application.Router, application.Store
}

// --- register-local validation tests ---

func TestRegisterLocalMissingAdminEmail(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	body, _ := json.Marshal(map[string]string{
		"name": "No Email Corp", "adminPassword": "pass123",
		"idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing adminEmail, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterLocalMissingAdminPassword(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	body, _ := json.Marshal(map[string]string{
		"name": "No Pass Corp", "adminEmail": "nopass@test.local",
		"idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing adminPassword, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterLocalMissingName(t *testing.T) {
	t.Parallel()
	router := syncRouter(t)

	body, _ := json.Marshal(map[string]string{
		"adminEmail": "noname@test.local", "adminPassword": "pass123",
		"idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- register-local creates user + company + member + key on SaaS ---

func TestRegisterLocalCreatesSaaSEntities(t *testing.T) {
	t.Parallel()
	router, st := syncRouterWithStore(t)
	ctx := context.Background()

	email := "saas-verify-" + uuid.Must(uuid.NewV7()).String()[:8] + "@test.local"
	body, _ := json.Marshal(map[string]string{
		"name": "Verify Corp", "adminEmail": email, "adminPassword": "secure-pass-123",
		"adminName": "Test Admin", "idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(body))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var result registerLocalTestResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	companyID := uuid.MustParse(result.CompanyID)

	// 1. Verify company was created on SaaS
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatalf("company not found on SaaS: %v", err)
	}
	if co.Type != store.CompanyTypeSelfhosted {
		t.Fatalf("expected selfhosted type, got %q", co.Type)
	}
	if co.Name != "Verify Corp" {
		t.Fatalf("company name: got %q want 'Verify Corp'", co.Name)
	}

	// 2. Verify user was created
	user, err := st.User().GetByEmail(ctx, email)
	if err != nil || user == nil {
		t.Fatalf("user not found on SaaS: %v", err)
	}
	if user.Name != "Test Admin" {
		t.Fatalf("user name: got %q want 'Test Admin'", user.Name)
	}

	// 3. Verify member was created (user is member of company)
	companies, err := st.User().ListMemberCompanies(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, mc := range companies {
		if mc.CompanyID == companyID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("user is not a member of the created company")
	}

	// 4. Verify NewAPI wallet user was provisioned
	if co.NewAPIWalletCompanyID == nil || *co.NewAPIWalletCompanyID <= 0 {
		t.Fatal("NewAPI wallet user not provisioned")
	}

	// 5. Verify platformKey is returned (non-empty)
	if result.PlatformKey == "" {
		t.Fatal("platformKey is empty")
	}

	// 6. Verify walletUserId matches company's wallet user
	if result.WalletUserID != *co.NewAPIWalletCompanyID {
		t.Fatalf("walletUserId mismatch: response=%d company=%d", result.WalletUserID, *co.NewAPIWalletCompanyID)
	}

	// 7. Verify sync token is stored (company has sync_token_hash)
	if co.SyncTokenHash == nil || *co.SyncTokenHash == "" {
		t.Fatal("sync token not stored on company")
	}
}
