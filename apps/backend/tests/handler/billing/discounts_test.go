package billing_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	saas "github.com/tokenjoy/backend/tests/testutil/saas"
)

func TestGetDiscountsEmpty(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Discount Co", "discount@example.com", "Discount Admin", "securepass123")

	// No discounts configured → should return empty array
	req := httptest.NewRequest(http.MethodGet, "/api/billing/discounts", nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var discounts []struct {
		ModelType string  `json:"modelType"`
		Discount  float64 `json:"discount"`
		Note      string  `json:"note"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&discounts); err != nil {
		t.Fatal(err)
	}
	if len(discounts) != 0 {
		t.Fatalf("expected empty discounts, got %d", len(discounts))
	}
}

func TestGetDiscountsAfterPlatformSet(t *testing.T) {
	t.Parallel()
	mock := saas.StartNewAPIMock(t)
	router := saas.NewRouter(t, mock)
	platformCookie := saas.LoginPlatform(t, router)
	provisioned := saas.ProvisionCompanyHTTP(t, router, platformCookie,
		"Discount Co2", "discount2@example.com", "Discount Admin2", "securepass123")

	// Platform admin sets a discount
	saas.SetCompanyDiscountHTTP(t, router, platformCookie, provisioned.Company.ID, "gpt-4o", 0.8, "contract")

	// Tenant member reads their own discounts
	req := httptest.NewRequest(http.MethodGet, "/api/billing/discounts", nil)
	req.Header.Set("Cookie", provisioned.MemberCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var discounts []struct {
		ModelType string  `json:"modelType"`
		Discount  float64 `json:"discount"`
		Note      string  `json:"note"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&discounts); err != nil {
		t.Fatal(err)
	}
	if len(discounts) != 1 {
		t.Fatalf("expected 1 discount, got %d", len(discounts))
	}
	if discounts[0].ModelType != "gpt-4o" {
		t.Errorf("expected modelType gpt-4o, got %s", discounts[0].ModelType)
	}
	if discounts[0].Discount != 0.8 {
		t.Errorf("expected discount 0.8, got %f", discounts[0].Discount)
	}
	if discounts[0].Note != "contract" {
		t.Errorf("expected note 'contract', got %q", discounts[0].Note)
	}
}
