package sms_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/integration/sms"
)

func TestClient_FetchCatalog_WithOAuth(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oauth/token":
			tokenCalls.Add(1)
			var body struct {
				GrantType    string `json:"grant_type"`
				ClientID     string `json:"client_id"`
				ClientSecret string `json:"client_secret"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			if body.GrantType != "client_credentials" {
				http.Error(w, "bad grant_type", 400)
				return
			}
			if body.ClientID != "test-client" || body.ClientSecret != "test-secret" {
				http.Error(w, "invalid credentials", 401)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-access-token",
				"token_type":   "Bearer",
				"expires_in":   600,
				"scope":        "sync:read",
			})

		case "/api/sync/catalog":
			auth := r.Header.Get("Authorization")
			if auth != "Bearer test-access-token" {
				http.Error(w, "unauthorized", 401)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"channels": []map[string]interface{}{
					{"name": "deepseek", "type": 43, "baseUrl": "https://api.deepseek.com", "key": "sk-x", "models": []string{"deepseek-chat"}, "group": "default", "priority": 0},
				},
				"models": []map[string]interface{}{
					{"modelId": "deepseek-chat", "displayName": "DeepSeek Chat", "provider": "deepseek", "callType": "chat", "inputPrice": 1.0, "outputPrice": 2.0},
				},
				"syncedAt": time.Now().Format(time.RFC3339),
			})

		default:
			http.Error(w, "not found", 404)
		}
	}))
	defer server.Close()

	client := sms.NewClient(sms.Config{
		BaseURL:      server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	catalog, err := client.FetchCatalog(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify catalog data
	if len(catalog.Channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(catalog.Channels))
	}
	if catalog.Channels[0].Name != "deepseek" {
		t.Fatalf("expected channel name deepseek, got %s", catalog.Channels[0].Name)
	}
	if len(catalog.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(catalog.Models))
	}
	if catalog.Models[0].ModelID != "deepseek-chat" {
		t.Fatalf("expected model deepseek-chat, got %s", catalog.Models[0].ModelID)
	}
	if catalog.Models[0].InputPrice != 1.0 {
		t.Fatalf("expected input price 1.0, got %f", catalog.Models[0].InputPrice)
	}

	// Verify token was fetched
	if tokenCalls.Load() != 1 {
		t.Fatalf("expected 1 token call, got %d", tokenCalls.Load())
	}

	// Second call should reuse cached token (no new token request)
	_, err = client.FetchCatalog(context.Background())
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if tokenCalls.Load() != 1 {
		t.Fatalf("expected token to be cached (still 1 call), got %d", tokenCalls.Load())
	}
}

func TestClient_FetchCatalog_InvalidCredentials(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth/token" {
			http.Error(w, "invalid credentials", 401)
			return
		}
		http.Error(w, "not found", 404)
	}))
	defer server.Close()

	client := sms.NewClient(sms.Config{
		BaseURL:      server.URL,
		ClientID:     "bad-client",
		ClientSecret: "bad-secret",
	})

	_, err := client.FetchCatalog(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
}

func TestClient_FetchCatalog_SMSUnreachable(t *testing.T) {
	t.Parallel()

	client := sms.NewClient(sms.Config{
		BaseURL:      "http://127.0.0.1:1", // unreachable
		ClientID:     "test",
		ClientSecret: "test",
	})

	_, err := client.FetchCatalog(context.Background())
	if err == nil {
		t.Fatal("expected error for unreachable SMS")
	}
}
