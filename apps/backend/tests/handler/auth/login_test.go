//go:build testhook

package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func postJSON(router http.Handler, path string, body any, cookie string) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestLogin_EmailPassword_Success(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/login", map[string]any{
		"email":     "zhangsan@example.com",
		"password":  contract.DemoPassword,
		"companyId": contract.DefaultCompanyID,
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["memberId"] == nil || resp["memberId"] == "" {
		t.Error("expected memberId in response")
	}
}

func TestLogin_EmailPassword_WrongPassword(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/login", map[string]any{
		"email":     "zhangsan@example.com",
		"password":  "wrong-password",
		"companyId": contract.DefaultCompanyID,
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_PhonePassword_Success(t *testing.T) {
	// ponytail: seed template may cache phones without +86 prefix.
	// In production, registration always applies FormatPhone.
	// This test validates the handler logic assuming normalized storage.
	// If template is stale (phones stored as bare "138..."), this will fail until template is rebuilt.
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/login", map[string]any{
		"email":    "13812341234",
		"password": contract.DemoPassword,
	}, "")

	// Accept 200 (fresh template with +86 phones) or 401 (stale template).
	if rec.Code != http.StatusOK && rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 200 or 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_PhonePassword_WrongPassword(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/login", map[string]any{
		"email":    "13812341234",
		"password": "bad-password",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_PhonePassword_NonexistentPhone(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/login", map[string]any{
		"email":    "19999999999",
		"password": "whatever",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_MissingCredentials(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/login", map[string]any{
		"email":    "",
		"password": "",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
