//go:build testhook

package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func extractSessionCookie(rec *httptest.ResponseRecorder) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == "tokenjoy_session_member" {
			return c.Name + "=" + c.Value
		}
	}
	return ""
}

func TestSetPassword_Success(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	// First login to get a real session cookie (with userID in claims).
	loginRec := postJSON(router, "/api/auth/login", map[string]any{
		"email":     "zhangsan@example.com",
		"password":  contract.DemoPassword,
		"companyId": contract.DefaultCompanyID,
	}, "")
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", loginRec.Code, loginRec.Body.String())
	}
	// Extract session cookie from login response.
	cookie := extractSessionCookie(loginRec)
	if cookie == "" {
		t.Fatal("no session cookie in login response")
	}

	rec := postJSON(router, "/api/auth/set-password", map[string]any{
		"password": "newpass123",
	}, cookie)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify new password works for login.
	rec2 := postJSON(router, "/api/auth/login", map[string]any{
		"email":     "zhangsan@example.com",
		"password":  "newpass123",
		"companyId": contract.DefaultCompanyID,
	}, "")
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 after password change, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestSetPassword_TooShort(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	// Login first to get a valid cookie.
	loginRec := postJSON(router, "/api/auth/login", map[string]any{
		"email":     "zhangsan@example.com",
		"password":  contract.DemoPassword,
		"companyId": contract.DefaultCompanyID,
	}, "")
	cookie := extractSessionCookie(loginRec)
	if cookie == "" {
		t.Fatal("no session cookie")
	}

	rec := postJSON(router, "/api/auth/set-password", map[string]any{
		"password": "short",
	}, cookie)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSetPassword_Unauthorized(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/set-password", map[string]any{
		"password": "newpass123",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestResetPassword_MissingFields(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/reset-password", map[string]any{
		"phone":       "",
		"code":        "",
		"newPassword": "",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestResetPassword_InvalidCode(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/auth/reset-password", map[string]any{
		"phone":       "13812341234",
		"code":        "000000",
		"newPassword": "newpass123",
	}, "")

	// Should fail because code is invalid (no SMS was sent).
	if rec.Code == http.StatusNoContent {
		t.Fatal("expected failure for invalid code, got 204")
	}
	// Accept 400 or 503 (if SMS not configured in test).
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 400 or 503, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_PhoneDetection(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	// An email-like string should require companyId in SaaS mode.
	// A phone-like string should use the phone path.
	cases := []struct {
		name    string
		email   string
		isPhone bool
	}{
		{"bare 11-digit", "13812341234", true},
		{"with +86", "+8613812341234", true},
		{"email format", "test@example.com", false},
		{"short number", "1234567", false},
		{"mixed chars", "138abc12341", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := postJSON(router, "/api/auth/login", map[string]any{
				"email":    tc.email,
				"password": "irrelevant",
			}, "")
			var resp map[string]any
			_ = json.Unmarshal(rec.Body.Bytes(), &resp)

			if tc.isPhone {
				// Phone path: should get 401 (invalid credentials, not "company id required").
				if rec.Code != http.StatusUnauthorized {
					t.Errorf("[%s] expected 401 for phone path, got %d body=%s", tc.name, rec.Code, rec.Body.String())
				}
			} else {
				// Email path in local mode: attempts email auth (401 for wrong creds).
				// In SaaS mode would require companyId → 400. In local mode: 401.
				if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusBadRequest {
					t.Errorf("[%s] expected 401 or 400 for email path, got %d body=%s", tc.name, rec.Code, rec.Body.String())
				}
			}
		})
	}
}
