//go:build testhook

package me_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func putJSON(router http.Handler, path string, body any, cookie string) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestUpdateProfile_Name(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := putJSON(router, "/api/me/profile", map[string]any{
		"name": "新姓名",
	}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify profile shows updated name.
	profileRec := getJSON(router, "/api/me/profile", cookie)
	if profileRec.Code != http.StatusOK {
		t.Fatalf("get profile failed: %d", profileRec.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(profileRec.Body.Bytes(), &resp)
	if resp["name"] != "新姓名" {
		t.Fatalf("expected name '新姓名', got %v", resp["name"])
	}
}

func TestUpdateProfile_AvatarDiceBear(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := putJSON(router, "/api/me/profile", map[string]any{
		"avatar": "dicebear:adventurer:myseed",
	}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify profile shows updated avatar.
	profileRec := getJSON(router, "/api/me/profile", cookie)
	if profileRec.Code != http.StatusOK {
		t.Fatalf("get profile failed: %d", profileRec.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(profileRec.Body.Bytes(), &resp)
	if resp["avatar"] != "dicebear:adventurer:myseed" {
		t.Fatalf("expected avatar 'dicebear:adventurer:myseed', got %v", resp["avatar"])
	}
}

func TestUpdateProfile_AvatarInvalidFormat(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := putJSON(router, "/api/me/profile", map[string]any{
		"avatar": "https://evil.com/avatar.png",
	}, cookie)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProfile_EmptyBody(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	rec := putJSON(router, "/api/me/profile", map[string]any{}, cookie)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProfile_Unauthorized(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := putJSON(router, "/api/me/profile", map[string]any{"name": "x"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUpdateProfile_ClearAvatar(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := login(router)
	if cookie == "" {
		t.Fatal("login failed")
	}

	// Set avatar first
	putJSON(router, "/api/me/profile", map[string]any{
		"avatar": "dicebear:bottts:x",
	}, cookie)

	// Clear it
	empty := ""
	rec := putJSON(router, "/api/me/profile", map[string]any{
		"avatar": empty,
	}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
}
