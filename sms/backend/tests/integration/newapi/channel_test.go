package newapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sms/backend/internal/integration/newapi"
)

func TestListChannels_ParsesResponse(t *testing.T) {
	t.Parallel()

	response := map[string]any{
		"success": true,
		"data": map[string]any{
			"page":      0,
			"page_size": 100,
			"total":     2,
			"items": []map[string]any{
				{
					"id":       1,
					"name":     "OpenAI-Main",
					"type":     1,
					"status":   1,
					"models":   "gpt-4o,gpt-4o-mini",
					"base_url": "https://api.openai.com",
					"priority": 0,
					"weight":   10,
				},
				{
					"id":       2,
					"name":     "Anthropic",
					"type":     14,
					"status":   1,
					"models":   "claude-sonnet-4",
					"base_url": "https://api.anthropic.com",
					"priority": 1,
					"weight":   5,
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/channel/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := newapi.NewClientForTest(srv.URL, "test-token")
	channels, err := client.ListChannels(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}

	ch := channels[0]
	if ch.ID != 1 {
		t.Fatalf("expected ID 1, got %d", ch.ID)
	}
	if ch.Name != "OpenAI-Main" {
		t.Fatalf("expected OpenAI-Main, got %s", ch.Name)
	}
	if ch.Type != 1 {
		t.Fatalf("expected type 1, got %d", ch.Type)
	}
	if ch.Models != "gpt-4o,gpt-4o-mini" {
		t.Fatalf("expected models gpt-4o,gpt-4o-mini, got %s", ch.Models)
	}
	if ch.BaseURL != "https://api.openai.com" {
		t.Fatalf("expected base_url, got %s", ch.BaseURL)
	}

	ch2 := channels[1]
	if ch2.Name != "Anthropic" {
		t.Fatalf("expected Anthropic, got %s", ch2.Name)
	}
}

func TestListChannels_EmptyResponse(t *testing.T) {
	t.Parallel()

	response := map[string]any{
		"success": true,
		"data": map[string]any{
			"page": 0, "page_size": 100, "total": 0, "items": []any{},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := newapi.NewClientForTest(srv.URL, "test-token")
	channels, err := client.ListChannels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 0 {
		t.Fatalf("expected 0 channels, got %d", len(channels))
	}
}
