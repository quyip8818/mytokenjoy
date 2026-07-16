package platform_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	saas "github.com/tokenjoy/backend/tests/testutil/saas"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func provisionTwoCompanies(t *testing.T) (http.Handler, saas.ProvisionedCompany, saas.ProvisionedCompany) {
	t.Helper()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	coA := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Iso A", "admin-a@iso.example", "Admin A", "securepass123")
	coB := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Iso B", "admin-b@iso.example", "Admin B", "securepass456")
	return router, coA, coB
}

func TestCompanyIsolationMembers(t *testing.T) {
	t.Parallel()
	router, coA, coB := provisionTwoCompanies(t)

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
			t.Fatalf("expected only company A members, got companyId=%d", m.CompanyID)
		}
	}
}

func TestCompanyIsolationBudgetTree(t *testing.T) {
	t.Parallel()
	router, coA, coB := provisionTwoCompanies(t)

	req := httptest.NewRequest(http.MethodGet, "/api/budget/tree", nil)
	req.Header.Set("Cookie", coA.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var tree []types.BudgetNode
	if err := json.NewDecoder(rec.Body).Decode(&tree); err != nil {
		t.Fatal(err)
	}
	rootA := fmt.Sprintf("dept-root-%d", coA.Company.ID)
	rootB := fmt.Sprintf("dept-root-%d", coB.Company.ID)
	for _, node := range tree {
		if node.ID == rootB {
			t.Fatal("company A session must not see company B budget root")
		}
		if node.ID != rootA && len(tree) == 1 {
			t.Fatalf("expected root %s, got %s", rootA, node.ID)
		}
	}
}

func TestCompanyIsolationPlatformKeys(t *testing.T) {
	t.Parallel()
	router, coA, coB := provisionTwoCompanies(t)

	req := httptest.NewRequest(http.MethodGet, "/api/keys/platform", nil)
	req.Header.Set("Cookie", coA.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result types.PageResult[types.PlatformKey]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	for _, k := range result.Items {
		if k.MemberID != nil && *k.MemberID == coB.Member.ID {
			t.Fatal("company A session must not see company B platform keys")
		}
	}
}
