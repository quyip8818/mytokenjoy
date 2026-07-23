//go:build testhook

package billing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

func TestDemoCompanyRechargeBlocked(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	app := testhttp.NewApp(t, func(cfg *config.Config) {
		saas.ApplyConfig(cfg)
		mock.ApplyToConfig(cfg)
	})
	router := app.Router
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Demo Co", "demo-admin@example.com", "Demo Admin", "securepass123")

	// Change company type to demo directly via pool
	pool := postgres.MainPool(app.Store)
	if pool == nil {
		t.Fatal("expected pool from store")
	}
	_, err := pool.Exec(context.Background(),
		`UPDATE companies SET type = $1 WHERE id = $2`,
		store.CompanyTypeDemo, provisioned.Company.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to create a recharge — should be 403
	body, _ := json.Marshal(map[string]any{
		"amount": 50.0, "idempotencyKey": "demo-recharge-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/recharge", bytes.NewReader(body))
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for demo recharge, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Attempt confirm — should also be 403
	req = httptest.NewRequest(http.MethodPost, "/api/billing/recharge/fake-id/confirm", nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for demo confirm, got %d body=%s", rec.Code, rec.Body.String())
	}
}
