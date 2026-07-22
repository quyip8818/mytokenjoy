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

	loginRec := postJSON(router, "/api/auth/login", map[string]any{
		"email":    "zhangsan@example.com",
		"password": contract.DemoPassword,
	}, "")
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", loginRec.Code, loginRec.Body.String())
	}
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

	rec2 := postJSON(router, "/api/auth/login", map[string]any{
		"email":    "zhangsan@example.com",
		"password": "newpass123",
	}, "")
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 after password change, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestAuthValidation(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	t.Run("SetPassword_TooShort", func(t *testing.T) {
		t.Parallel()
		loginRec := postJSON(router, "/api/auth/login", map[string]any{
			"email":    "zhangsan@example.com",
			"password": contract.DemoPassword,
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
	})

	t.Run("SetPassword_Unauthorized", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/set-password", map[string]any{
			"password": "newpass123",
		}, "")

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("ResetPassword_MissingFields", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/reset-password", map[string]any{
			"phone":       "",
			"code":        "",
			"newPassword": "",
		}, "")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("ResetPassword_InvalidCode", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/reset-password", map[string]any{
			"phone":       "13812341234",
			"code":        "000000",
			"newPassword": "newpass123",
		}, "")

		if rec.Code == http.StatusNoContent {
			t.Fatal("expected failure for invalid code, got 204")
		}
		if rec.Code != http.StatusBadRequest && rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 400 or 503, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Login_PhoneDetection", func(t *testing.T) {
		t.Parallel()
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
					if rec.Code != http.StatusUnauthorized {
						t.Errorf("[%s] expected 401 for phone path, got %d body=%s", tc.name, rec.Code, rec.Body.String())
					}
				} else {
					if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusBadRequest {
						t.Errorf("[%s] expected 401 or 400 for email path, got %d body=%s", tc.name, rec.Code, rec.Body.String())
					}
				}
			})
		}
	})
}
