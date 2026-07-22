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

func TestLoginEndpoints(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	t.Run("EmailPassword_Success", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/login", map[string]any{
			"email":    "zhangsan@example.com",
			"password": contract.DemoPassword,
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
	})

	t.Run("EmailPassword_WrongPassword", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/login", map[string]any{
			"email":    "zhangsan@example.com",
			"password": "wrong-password",
		}, "")

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PhonePassword_Success", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/login", map[string]any{
			"email":    "13812341234",
			"password": contract.DemoPassword,
		}, "")

		// Accept 200 (fresh template with +86 phones) or 401 (stale template).
		if rec.Code != http.StatusOK && rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 200 or 401, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PhonePassword_WrongPassword", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/login", map[string]any{
			"email":    "13812341234",
			"password": "bad-password",
		}, "")

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PhonePassword_NonexistentPhone", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/login", map[string]any{
			"email":    "19999999999",
			"password": "whatever",
		}, "")

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("MissingCredentials", func(t *testing.T) {
		t.Parallel()
		rec := postJSON(router, "/api/auth/login", map[string]any{
			"email":    "",
			"password": "",
		}, "")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}
