//go:build integration

package newapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestUpstreamRefMatchesNewAPIPin(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	pinPath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../../../../apps/newapi/patches/new-api/UPSTREAM_REF"))
	raw, err := os.ReadFile(pinPath)
	if err != nil {
		t.Fatalf("read UPSTREAM_REF: %v", err)
	}
	got := strings.TrimSpace(string(raw))
	if got != newapi.UpstreamRef {
		t.Fatalf("adapter UpstreamRef=%q does not match %s (%q)", newapi.UpstreamRef, pinPath, got)
	}
}

func TestCreateUserResolvesIDWhenUpstreamReturnsEmptyData(t *testing.T) {
	t.Parallel()
	const username = "company-9"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/user/":
			_, _ = w.Write([]byte(`{"success":true,"message":""}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/user/search":
			if r.URL.Query().Get("keyword") != username {
				t.Fatalf("expected keyword %q, got %q", username, r.URL.Query().Get("keyword"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"page": 1, "page_size": 10, "total": 1,
					"items": []map[string]any{{"id": 42, "username": username, "quota": 0}},
				},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	client := newapi.NewClient(server.URL, "admin", 1)
	user, err := client.CreateUser(context.Background(), newapi.CreateUserRequest{
		Username: username, DisplayName: "Co", Password: "abcdefgh",
	})
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != 42 {
		t.Fatalf("expected id 42, got %d", user.ID)
	}
}

func TestTopUpUsesManageAddQuotaNotLegacyTopupPath(t *testing.T) {
	t.Parallel()
	var sawManage bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/topup":
			t.Fatal("must not call legacy /api/topup")
		case r.Method == http.MethodPost && r.URL.Path == "/api/user/manage":
			sawManage = true
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "add_quota" || body["mode"] != "subtract" {
				t.Fatalf("unexpected manage body %#v", body)
			}
			if body["id"] != float64(7) || body["value"] != float64(9) {
				t.Fatalf("unexpected id/value %#v", body)
			}
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	client := newapi.NewClient(server.URL, "admin", 1)
	if err := client.TopUp(context.Background(), newapi.TopUpRequest{UserID: 7, Quota: -9}); err != nil {
		t.Fatal(err)
	}
	if !sawManage {
		t.Fatal("expected /api/user/manage")
	}
}

func TestUpsertChannelCreateUsesModeSingleAndResolvesID(t *testing.T) {
	t.Parallel()
	var createBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/channel/":
			_ = json.NewDecoder(r.Body).Decode(&createBody)
			_, _ = w.Write([]byte(`{"success":true,"message":""}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/channel/"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"page": 1, "page_size": 100, "total": 1,
					"items": []map[string]any{{"id": 55, "name": "pk-openai", "type": 1, "status": 1, "group": "platform_shared"}},
				},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	client := newapi.NewClient(server.URL, "admin", 1)
	ch, err := client.UpsertChannel(context.Background(), newapi.UpsertChannelRequest{
		Type: 1, Name: "pk-openai", Key: "sk-x", Status: 1, Group: "platform_shared",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != 55 {
		t.Fatalf("expected channel id 55, got %d", ch.ID)
	}
	if createBody["mode"] != "single" {
		t.Fatalf("expected mode=single, got %#v", createBody["mode"])
	}
	channel, _ := createBody["channel"].(map[string]any)
	if channel == nil || channel["name"] != "pk-openai" || channel["key"] != "sk-x" {
		t.Fatalf("unexpected channel body %#v", createBody["channel"])
	}
}

func TestUpsertChannelUpdatePreservesUnspecifiedFields(t *testing.T) {
	t.Parallel()
	var putBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/channel/12":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"id": 12, "type": 1, "name": "pk-openai", "key": "sk-old",
					"status": 1, "group": "platform_shared",
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/channel/":
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": putBody})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	client := newapi.NewClient(server.URL, "admin", 1)
	disabled := 2
	_, err := client.UpsertChannel(context.Background(), newapi.UpsertChannelRequest{
		ID: 12, Status: disabled, Key: "sk-new",
	})
	if err != nil {
		t.Fatal(err)
	}
	if putBody["id"] != float64(12) || putBody["status"] != float64(2) || putBody["key"] != "sk-new" {
		t.Fatalf("unexpected put %#v", putBody)
	}
	if putBody["name"] != "pk-openai" || putBody["group"] != "platform_shared" {
		t.Fatalf("expected preserve name/group, got %#v", putBody)
	}
}

func TestRebuildAbilitiesPostsChannelFix(t *testing.T) {
	t.Parallel()
	var hit bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/channel/sync" {
			t.Fatal("must not call legacy GET /api/channel/sync")
		}
		if r.Method == http.MethodPost && r.URL.Path == "/api/channel/fix" {
			hit = true
			_, _ = w.Write([]byte(`{"success":true,"data":{"success":1,"fails":0}}`))
			return
		}
		t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(server.Close)

	client := newapi.NewClient(server.URL, "admin", 1)
	if err := client.RebuildAbilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("expected POST /api/channel/fix")
	}
}

func TestUpdateTokenPreservesExpiredTimeAndIdentity(t *testing.T) {
	t.Parallel()
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/token/9":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"id": 9, "name": "tokenjoy:plk-1", "status": 1,
					"remain_quota": 10, "unlimited_quota": false,
					"model_limits_enabled": true, "model_limits": "test-model",
					"group": "dept-dept-3", "expired_time": -1,
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/token/":
			_ = json.NewDecoder(r.Body).Decode(&got)
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": got})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	client := newapi.NewClient(srv.URL, "tok", 1)
	remain := int64(42)
	_, err := client.UpdateToken(t.Context(), newapi.UpdateTokenRequest{
		ID:          9,
		RemainQuota: &remain,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["expired_time"] != float64(-1) {
		t.Fatalf("expected expired_time=-1 preserved, got %#v", got["expired_time"])
	}
	if got["name"] != "tokenjoy:plk-1" {
		t.Fatalf("expected name preserved, got %#v", got["name"])
	}
	if got["group"] != "dept-dept-3" {
		t.Fatalf("expected group preserved, got %#v", got["group"])
	}
	if got["remain_quota"] != float64(42) {
		t.Fatalf("expected remain_quota=42, got %#v", got["remain_quota"])
	}
}

func TestUpdateTokenHealsZeroExpiredTime(t *testing.T) {
	t.Parallel()
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/token/3":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"id": 3, "name": "tokenjoy:plk-2", "status": 1,
					"remain_quota": 0, "expired_time": 0, "group": "dept-dept-3",
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/token/":
			_ = json.NewDecoder(r.Body).Decode(&got)
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": got})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	client := newapi.NewClient(srv.URL, "tok", 1)
	remain := int64(100)
	if _, err := client.UpdateToken(t.Context(), newapi.UpdateTokenRequest{
		ID: 3, RemainQuota: &remain,
	}); err != nil {
		t.Fatal(err)
	}
	if got["expired_time"] != float64(-1) {
		t.Fatalf("expected heal expired_time 0→-1, got %#v", got["expired_time"])
	}
}
