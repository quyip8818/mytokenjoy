//go:build integration

package newapi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestCreateTokenUsesResponseDataAndAssertsOwner(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/token/":
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &gotBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"id": 99, "user_id": 2, "name": "tokenjoy:plk-1", "key": "sk-created",
					"remain_quota": 10, "group": "platform_shared",
				},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newapi.NewClient(server.URL, "admin-token", 1)
	token, err := client.CreateToken(context.Background(), newapi.CreateTokenRequest{
		UserID: 2, Name: "tokenjoy:plk-1", RemainQuota: 10, Group: "platform_shared", ExpiredTime: -1,
	})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if gotBody["user_id"] != float64(2) {
		t.Fatalf("expected request user_id 2, got %#v", gotBody["user_id"])
	}
	if token.ID != 99 || token.UserID != 2 || token.Key != "sk-created" {
		t.Fatalf("unexpected token %#v", token)
	}
}

func TestCreateTokenFailsWhenResponseMissingID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/api/token/" {
			_, _ = w.Write([]byte(`{"success":true,"message":"","data":null}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := newapi.NewClient(server.URL, "admin-token", 1)
	_, err := client.CreateToken(context.Background(), newapi.CreateTokenRequest{UserID: 2, Name: "tokenjoy:x"})
	if err == nil || !strings.Contains(err.Error(), "missing id") {
		t.Fatalf("expected missing id error, got %v", err)
	}
}

func TestCreateTokenFailsAndDeletesOnOwnerMismatch(t *testing.T) {
	t.Parallel()

	var deletedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/token/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data":    map[string]any{"id": 55, "user_id": 1, "name": "tokenjoy:x", "key": "sk-wrong"},
			})
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/token/"):
			deletedPath = r.URL.Path
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newapi.NewClient(server.URL, "admin-token", 1)
	_, err := client.CreateToken(context.Background(), newapi.CreateTokenRequest{UserID: 2, Name: "tokenjoy:x"})
	if err == nil || !strings.Contains(err.Error(), "owner mismatch") {
		t.Fatalf("expected owner mismatch, got %v", err)
	}
	if deletedPath != "/api/token/55" {
		t.Fatalf("expected delete /api/token/55, got %q", deletedPath)
	}
}

func TestCreateTokenRegeneratesWhenKeyMasked(t *testing.T) {
	t.Parallel()

	var regenerateCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/token/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data":    map[string]any{"id": 7, "user_id": 2, "name": "tokenjoy:plk-masked", "key": "sk-****"},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/regenerate"):
			regenerateCalled = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data":    map[string]any{"id": 7, "user_id": 2, "name": "tokenjoy:plk-masked", "key": "sk-full"},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newapi.NewClient(server.URL, "admin-token", 1)
	token, err := client.CreateToken(context.Background(), newapi.CreateTokenRequest{
		UserID: 2, Name: "tokenjoy:plk-masked",
	})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if !regenerateCalled {
		t.Fatal("expected regenerate when key is masked")
	}
	if token.Key != "sk-full" {
		t.Fatalf("expected sk-full, got %q", token.Key)
	}
}
