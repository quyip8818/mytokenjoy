//go:build testhook && integration

package localsaas_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	catalog "github.com/tokenjoy/backend/internal/integration/catalogsync"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/worker/catalogsync"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytfpkg "github.com/tokenjoy/backend/tests/testutil/gateway"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

const testRegistrationSecret = "test-local-reg-secret-32bytes!!"

// TestLocalSaaSRegisterAndSyncLots is an end-to-end test:
// 1. Register a selfhosted company on SaaS (creates company + user + wallet + token)
// 2. Recharge the company on SaaS
// 3. Run catalog sync (lots) from "Local" perspective
// 4. Verify Local DB has the lots and wallet balance
func TestLocalSaaSRegisterAndSyncLots(t *testing.T) {
	t.Parallel()

	// --- Setup SaaS router (simulates the SaaS platform) ---
	mock := saas.StartNewAPIMock(t)
	saasApp := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
		cfg.LocalRegistrationSecret = testRegistrationSecret
	}, app.WithoutWorker(), app.WithAdminPort(mock.AdminPort()))
	saasRouter := saasApp.Router

	// --- Step 1: Register a selfhosted company ---
	regBody, _ := json.Marshal(map[string]string{
		"name":           "E2E Local Corp",
		"adminEmail":     "e2e-admin@local.test",
		"adminPassword":  "e2e-password-123",
		"idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(regBody))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	saasRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var regResult struct {
		CompanyID    string `json:"companyId"`
		WalletUserID int64  `json:"walletUserId"`
		PlatformKey  string `json:"platformKey"`
		SyncToken    string `json:"syncToken"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&regResult); err != nil {
		t.Fatal(err)
	}
	if regResult.CompanyID == "" || regResult.PlatformKey == "" || regResult.SyncToken == "" {
		t.Fatalf("register result missing fields: %+v", regResult)
	}
	companyID := uuid.MustParse(regResult.CompanyID)

	// --- Step 2: Recharge the company ---
	platformCookie := saas.LoginPlatform(t, saasRouter)
	rechargeBody, _ := json.Marshal(map[string]float64{"amount": 200.0})
	req = httptest.NewRequest(http.MethodPost, "/api/platform/companies/"+regResult.CompanyID+"/recharge", bytes.NewReader(rechargeBody))
	req.Header.Set("Cookie", platformCookie)
	rec = httptest.NewRecorder()
	saasRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("recharge: expected success, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify SaaS company now has positive wallet.
	saasCo, err := saasApp.Store.Company().GetByID(context.Background(), companyID)
	if err != nil || saasCo == nil {
		t.Fatalf("get company from SaaS: %v", err)
	}
	if saasCo.WalletRemainQuota <= 0 {
		t.Fatalf("SaaS wallet should be positive after recharge, got %d", saasCo.WalletRemainQuota)
	}

	// --- Step 3: Setup "Local" store and run catalog sync ---
	// Create a separate local store that will sync from the SaaS HTTP server.
	saasServer := httptest.NewServer(saasRouter)
	defer saasServer.Close()

	localCfg, localStore := testutil.NewFreshTestStore(t,
		testutil.WithIngestEnabled(true),
		func(cfg *config.Config) {
			cfg.SupportSaas = false
			cfg.CompanyID = companyID
		},
	)
	_ = localCfg

	// Seed the company row in local DB (mimics what bootstrap would do after setup).
	// Bootstrap already creates this company via insertCompanies when cfg.CompanyID is set,
	// so we only update it if needed.
	now := time.Now().UTC()
	localCo := store.Company{
		ID: companyID, Name: "E2E Local Corp", Type: store.CompanyTypeSelfhosted,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := localStore.Company().Create(context.Background(), localCo); err != nil {
		// Bootstrap may have already created it — ignore duplicate.
		if existing, getErr := localStore.Company().GetByID(context.Background(), companyID); getErr != nil || existing == nil {
			t.Fatal(err)
		}
	}

	// Run catalog sync executor (lots sync).
	client := catalog.NewClient(catalog.Config{
		BaseURL:   saasServer.URL,
		SyncToken: regResult.SyncToken,
	})
	globalCompanyID := uuid.MustParse("00000000-0000-7000-8000-000000000001") // tokenjoy global
	executor := catalogsync.NewExecutor(client, mock.AdminPort(), localStore, globalCompanyID, companyID)

	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("catalog sync execute: %v", err)
	}

	// --- Step 4: Verify Local has lots + wallet balance ---
	localCoAfter, err := localStore.Company().GetByID(context.Background(), companyID)
	if err != nil || localCoAfter == nil {
		t.Fatalf("get local company: %v", err)
	}
	if localCoAfter.WalletRemainQuota != saasCo.WalletRemainQuota {
		t.Fatalf("local wallet should match SaaS: local=%d saas=%d",
			localCoAfter.WalletRemainQuota, saasCo.WalletRemainQuota)
	}

	// Verify lots were synced.
	lots, err := localStore.Billing().ListActiveLotsFIFO(context.Background(), companyID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lots) == 0 {
		t.Fatal("expected lots to be synced to local DB")
	}
	t.Logf("e2e: synced %d lots, wallet=%d", len(lots), localCoAfter.WalletRemainQuota)
}

// TestWalletLotsVersionBump verifies the version-gated sync flow:
// 1. Register + recharge → wallet_lots_version bumps
// 2. First catalog sync picks up lots (version differs)
// 3. Second catalog sync is a no-op (version matches)
func TestWalletLotsVersionBump(t *testing.T) {
	t.Parallel()

	mock := saas.StartNewAPIMock(t)
	saasApp := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
		cfg.LocalRegistrationSecret = testRegistrationSecret
	}, app.WithoutWorker(), app.WithAdminPort(mock.AdminPort()))
	saasRouter := saasApp.Router

	// Register company
	regBody, _ := json.Marshal(map[string]string{
		"name": "Version Bump Corp", "adminEmail": "bump@test.local",
		"adminPassword": "test-password-123", "idempotencyKey": uuid.Must(uuid.NewV7()).String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/register-local", bytes.NewReader(regBody))
	req.Header.Set("X-Registration-Secret", testRegistrationSecret)
	rec := httptest.NewRecorder()
	saasRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rec.Code, rec.Body.String())
	}
	var regResult struct {
		CompanyID string `json:"companyId"`
		SyncToken string `json:"syncToken"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&regResult)
	companyID := uuid.MustParse(regResult.CompanyID)

	// Check initial wallet_lots_version (should be 0 — no lot activity yet)
	ctx := context.Background()
	v0, _ := saasApp.Store.SystemSettings().Get(ctx, "catalog.wallet_lots_version")

	// Recharge → should bump version
	platformCookie := saas.LoginPlatform(t, saasRouter)
	rechargeBody, _ := json.Marshal(map[string]float64{"amount": 50.0})
	req = httptest.NewRequest(http.MethodPost, "/api/platform/companies/"+regResult.CompanyID+"/recharge", bytes.NewReader(rechargeBody))
	req.Header.Set("Cookie", platformCookie)
	rec = httptest.NewRecorder()
	saasRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("recharge: %d %s", rec.Code, rec.Body.String())
	}

	v1, _ := saasApp.Store.SystemSettings().Get(ctx, "catalog.wallet_lots_version")
	if v1 == v0 {
		t.Fatalf("expected wallet_lots_version to bump after recharge: before=%q after=%q", v0, v1)
	}

	// Setup local store + run catalog sync
	saasServer := httptest.NewServer(saasRouter)
	defer saasServer.Close()

	_, localStore := testutil.NewFreshTestStore(t, testutil.WithIngestEnabled(true), func(cfg *config.Config) {
		cfg.CompanyID = companyID
	})
	now := time.Now().UTC()
	_ = localStore.Company().Create(ctx, store.Company{
		ID: companyID, Name: "Version Bump Corp", Type: store.CompanyTypeSelfhosted,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	})

	client := catalog.NewClient(catalog.Config{BaseURL: saasServer.URL, SyncToken: regResult.SyncToken})
	globalID := uuid.MustParse("00000000-0000-7000-8000-000000000001")
	executor := catalogsync.NewExecutor(client, mock.AdminPort(), localStore, globalID, companyID)

	// First sync — should fetch lots (versions differ)
	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("first sync: %v", err)
	}
	localV, _ := localStore.SystemSettings().Get(ctx, "catalog.wallet_lots_version")
	if localV != v1 {
		t.Fatalf("local wallet_lots_version should match SaaS: local=%q saas=%q", localV, v1)
	}
	lots, _ := localStore.Billing().ListActiveLotsFIFO(ctx, companyID, nil)
	if len(lots) == 0 {
		t.Fatal("expected lots after first sync")
	}

	// Second sync (no change on SaaS) — should be a no-op
	lotsBefore := len(lots)
	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("second sync: %v", err)
	}
	lotsAfter, _ := localStore.Billing().ListActiveLotsFIFO(ctx, companyID, nil)
	if len(lotsAfter) != lotsBefore {
		t.Fatalf("expected no new lots on second sync: before=%d after=%d", lotsBefore, len(lotsAfter))
	}
}

// TestGatewaySelfhostedWalletZeroHTTP tests that a selfhosted company with wallet=0
// can still make requests through the gateway (wallet check skipped).
func TestGatewaySelfhostedWalletZeroHTTP(t *testing.T) {
	t.Parallel()

	// Selfhosted + wallet=0 + budget > 0 → request passes precheck.
	zeroWallet := float64(0)
	scenario := gatewaytfpkg.BuildGatewayScenario(t, gatewaytfpkg.GatewayScenarioOpts{
		Budget:             10000,
		WalletBalancePoint: &zeroWallet,
		CompanyType:        store.CompanyTypeSelfhosted,
	})

	// Make a gateway request — should pass (wallet skip for selfhosted).
	body := []byte(`{"model":"deepseek-v4-pro","messages":[{"role":"user","content":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+scenario.FullKey)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, req)

	// Should NOT be 403 (would be 403 if wallet check triggered).
	// Expect 200 (backend mock) or at least not 403.
	if rec.Code == http.StatusForbidden {
		t.Fatalf("selfhosted wallet=0 request was rejected: %d %s", rec.Code, rec.Body.String())
	}
}
