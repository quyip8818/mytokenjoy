package platform_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
	saas "github.com/tokenjoy/backend/tests/testutil/saas"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func TestPlatformCompaniesUnauthorized(t *testing.T) {
	t.Parallel()
	router := saas.NewRouter(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/platform/companies", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPlatformCompaniesForbiddenForTenantMember(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Tenant Co", "tenant@example.com", "Tenant Admin", "securepass123")

	// Use the tenant member's cookie to access platform API — should be forbidden
	req := httptest.NewRequest(http.MethodGet, "/api/platform/companies", nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for tenant member accessing platform API, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlatformLoginRejectsBadCredentials(t *testing.T) {
	t.Parallel()
	router := saas.NewRouter(t, nil)
	body, _ := json.Marshal(map[string]string{
		"email": saas.PlatformBootstrapEmail, "password": "wrong-password",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPlatformCreateCompanyAndRecharge(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)

	created := saas.CreateCompanyHTTP(t, router, platformCookie, "Acme Corp", "ceo@acme.example")
	saas.PlatformRechargeHTTP(t, router, platformCookie, created.Company.ID, 100)
}

func TestPlatformListCompaniesIncludesCreated(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	created := saas.CreateCompanyHTTP(t, router, platformCookie, "Listed Co", "admin@listed.example")

	req := httptest.NewRequest(http.MethodGet, "/api/platform/companies", nil)
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var companies []store.Company
	if err := json.NewDecoder(rec.Body).Decode(&companies); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range companies {
		if c.ID == created.Company.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected created company in platform list")
	}
}

func TestPlatformCreateChannelAndSaaSProviderForbidden(t *testing.T) {
	t.Parallel()
	router := saas.NewRouter(t, nil)
	platformCookie := saas.LoginPlatform(t, router)

	channelBody, _ := json.Marshal(map[string]string{
		"provider": "openai",
		"name":     "shared-openai",
		"key":      "sk-platform-channel",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/platform/channels", bytes.NewReader(channelBody))
	req.Header.Set("Cookie", platformCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create channel: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	providerBody, _ := json.Marshal(map[string]string{
		"provider": "openai",
		"name":     "company-key",
		"key":      "sk-company",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/keys/provider", bytes.NewReader(providerBody))
	req.Header.Set("Cookie", saas.DefaultSeedMemberCookie(t))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("company provider create: expected 403, got %d", rec.Code)
	}
}

func TestCompanyIsolationUsesSessionCompany(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)

	coA := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Company A", "admin-a@example.com", "Admin A", "securepass123")
	coB := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Company B", "admin-b@example.com", "Admin B", "securepass456")

	req := httptest.NewRequest(http.MethodGet, "/api/org/members", nil)
	req.Header.Set("Cookie", coA.MemberCookie)
	req.Header.Set("X-Company-Id", fmt.Sprintf("%d", coB.Company.ID))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var result types.PageResult[types.Member]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	for _, m := range result.Items {
		if m.CompanyID != coA.Company.ID {
			t.Fatalf("session must scope to company A (%d), saw member from company %d", coA.Company.ID, m.CompanyID)
		}
		if m.ID == coB.Member.ID {
			t.Fatal("company A session must not see company B members")
		}
	}
}

func TestSuspendedCompanyBlocksWrites(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Suspend Co", "suspend@example.com", "Suspend Admin", "securepass123")

	saas.UpdateCompanyStatusHTTP(t, router, platformCookie, provisioned.Company.ID, store.CompanyStatusSuspended)

	// Any valid budget write endpoint suffices — use a random UUID as deptID;
	// it will hit the "company suspended" guard before reaching dept lookup.
	body, _ := json.Marshal(map[string]float64{"budget": 1000})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/budget/departments/%s", uuid.Must(uuid.NewV7())), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for write on suspended company, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/budget/tree", nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for read on suspended company, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSuspendedCompanyGatewayRejected(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:        1000,
		CompanyStatus: store.CompanyStatusSuspended,
	})

	body, _ := json.Marshal(map[string]string{"model": "gpt-4o"})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+scenario.FullKey)
	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for suspended company gateway, got %d", rec.Code)
	}
}
