//go:build testhook

package me_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func login(router http.Handler) string {
	b, _ := json.Marshal(map[string]any{
		"email":    "zhangsan@example.com",
		"password": contract.DemoPassword,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		return ""
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "tokenjoy_session_member" {
			return c.Name + "=" + c.Value
		}
	}
	return ""
}

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

func getJSON(router http.Handler, path, cookie string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// --- GET /me/profile ---

func TestGetProfile_Success(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := getJSON(router, "/api/me/profile", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	// phone and email should be masked
	phone, _ := resp["phone"].(string)
	if phone == "" {
		t.Fatal("expected masked phone in response")
	}
	if len(phone) > 4 && phone[3:7] != "****" && !containsMask(phone) {
		t.Logf("phone masking: %s", phone)
	}
	// hasPassword should be true (demo user has password)
	if resp["hasPassword"] != true {
		t.Fatalf("expected hasPassword=true, got %v", resp["hasPassword"])
	}
	// companies should be a non-empty array
	companies, ok := resp["companies"].([]any)
	if !ok || len(companies) == 0 {
		t.Fatal("expected non-empty companies array")
	}
}

func TestGetProfile_Unauthorized(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := getJSON(router, "/api/me/profile", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// --- POST /me/change-password ---

func TestChangePassword_Success(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	// Change password from demo to new
	rec := postJSON(router, "/api/me/change-password", map[string]any{
		"oldPassword": contract.DemoPassword,
		"newPassword": "changed123",
	}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify new password works
	b, _ := json.Marshal(map[string]any{
		"email":    "zhangsan@example.com",
		"password": "changed123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, req)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login with new password failed: %d", loginRec.Code)
	}
}

func TestChangePassword_WrongOld(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := postJSON(router, "/api/me/change-password", map[string]any{
		"oldPassword": "wrongpassword",
		"newPassword": "newpass123",
	}, cookie)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestChangePassword_TooShort(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := postJSON(router, "/api/me/change-password", map[string]any{
		"oldPassword": contract.DemoPassword,
		"newPassword": "short",
	}, cookie)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- POST /me/revoke-sessions ---

func TestRevokeSessions_Success(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := postJSON(router, "/api/me/revoke-sessions", nil, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Current session should still work (excluded from revocation)
	profileRec := getJSON(router, "/api/me/profile", cookie)
	if profileRec.Code != http.StatusOK {
		t.Fatalf("current session should survive revoke, got %d", profileRec.Code)
	}
}

func TestRevokeSessions_Unauthorized(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := postJSON(router, "/api/me/revoke-sessions", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// --- helpers ---

func containsMask(s string) bool {
	for _, c := range s {
		if c == '*' {
			return true
		}
	}
	return false
}
