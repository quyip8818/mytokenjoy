package newapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestCreateTokenFindsTokenAcrossPagesByUniqueName(t *testing.T) {
	t.Parallel()

	const tokenName = "tokenjoy:plk-page-test"
	var listCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/token/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"message":"","data":null}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/token/"):
			listCalls++
			page, _ := strconv.Atoi(r.URL.Query().Get("p"))
			var payload any
			switch page {
			case 0:
				payload = map[string]any{
					"page":      0,
					"page_size": 2,
					"total":     3,
					"items": []map[string]any{
						{"id": 1, "name": "other", "key": "sk-other"},
						{"id": 2, "name": "another", "key": "sk-another"},
					},
				}
			case 1:
				payload = map[string]any{
					"page":      1,
					"page_size": 2,
					"total":     3,
					"items": []map[string]any{
						{"id": 99, "name": tokenName, "key": "sk-found"},
					},
				}
			default:
				payload = map[string]any{"page": page, "page_size": 2, "total": 3, "items": []any{}}
			}
			raw, err := json.Marshal(map[string]any{"success": true, "message": "", "data": payload})
			if err != nil {
				t.Fatal(err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(raw)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newapi.NewClient(server.URL, "admin-token", 1)
	token, err := client.CreateToken(context.Background(), newapi.CreateTokenRequest{Name: tokenName})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if token.ID != 99 {
		t.Fatalf("expected token id 99, got %d", token.ID)
	}
	if token.Key != "sk-found" {
		t.Fatalf("expected sk-found, got %q", token.Key)
	}
	if listCalls < 2 {
		t.Fatalf("expected at least 2 list pages, got %d", listCalls)
	}
}

func TestCreateTokenRegeneratesWhenKeyMasked(t *testing.T) {
	t.Parallel()

	const tokenName = "tokenjoy:plk-masked"
	var regenerateCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/token/":
			_, _ = w.Write([]byte(`{"success":true,"message":"","data":null}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/token/"):
			raw, _ := json.Marshal(map[string]any{
				"success": true,
				"data": map[string]any{
					"page": 0, "page_size": 10, "total": 1,
					"items": []map[string]any{{"id": 7, "name": tokenName, "key": "sk-****"}},
				},
			})
			_, _ = w.Write(raw)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/regenerate"):
			regenerateCalled = true
			raw, _ := json.Marshal(map[string]any{
				"success": true,
				"data":    map[string]any{"id": 7, "name": tokenName, "key": "sk-full"},
			})
			_, _ = w.Write(raw)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newapi.NewClient(server.URL, "admin-token", 1)
	token, err := client.CreateToken(context.Background(), newapi.CreateTokenRequest{Name: tokenName})
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
