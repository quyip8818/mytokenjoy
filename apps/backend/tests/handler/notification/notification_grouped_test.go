//go:build testhook

package notification_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

// TestNotificationGrouped verifies that notifications with the same group_key
// are collapsed in grouped queries, showing the latest entry with a group_count badge.
// This test relies on the demo seed which already creates grouped budget_alert entries.
func TestNotificationGrouped(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	// The demo seed (ApplyNotifications) creates 2 budget_alert entries with the same
	// group_key ("budget:{ruleID}:{period}"). Query grouped to verify they collapse.
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?grouped=true&category=budget_alert&limit=20", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Items []struct {
			ID         string `json:"id"`
			EventType  string `json:"eventType"`
			Category   string `json:"category"`
			GroupKey   string `json:"groupKey"`
			GroupCount int    `json:"groupCount"`
		} `json:"items"`
		HasMore bool `json:"hasMore"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Items) == 0 {
		t.Fatal("expected at least 1 grouped budget_alert item from seed data")
	}

	// The seed creates 2 entries with the same group_key — they should collapse into 1 group.
	// Find the entry with groupCount >= 2.
	var found bool
	for _, item := range resp.Items {
		if item.GroupCount >= 2 {
			found = true
			if item.GroupKey == "" {
				t.Fatal("grouped entry should have a non-empty groupKey")
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected at least one group with count >= 2, got items: %+v", resp.Items)
	}
}

// TestNotificationAdminLog verifies that the admin log endpoint returns
// all notification entries (including failed deliveries) with proper filtering.
func TestNotificationAdminLog(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	// Seed a notification so there's at least one entry.
	seedNotification(t, router, cookie, types.NotificationLogEntry{
		UserID:    contract.IDMemberAdmin,
		EventType: "budget_alert_reached",
		Title:     "Admin Log Test",
		Body:      "Testing admin log query",
	})
	time.Sleep(50 * time.Millisecond)

	// Query admin log — no filter.
	req := httptest.NewRequest(http.MethodGet, "/api/notifications/admin/log?limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("admin/log: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var logResp []struct {
		ID        string `json:"id"`
		Channel   string `json:"channel"`
		EventType string `json:"eventType"`
		SendOK    bool   `json:"sendOk"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &logResp); err != nil {
		t.Fatal(err)
	}
	if len(logResp) == 0 {
		t.Fatal("admin/log: expected at least 1 entry")
	}

	// Query with channel filter.
	req = httptest.NewRequest(http.MethodGet, "/api/notifications/admin/log?channel=in_app&limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("admin/log channel filter: expected 200, got %d", rec.Code)
	}

	var filteredResp []struct {
		Channel string `json:"channel"`
	}
	json.Unmarshal(rec.Body.Bytes(), &filteredResp)
	for _, entry := range filteredResp {
		if entry.Channel != "in_app" {
			t.Fatalf("expected only in_app entries, got %q", entry.Channel)
		}
	}

	// Query admin stats.
	req = httptest.NewRequest(http.MethodGet, "/api/notifications/admin/stats", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("admin/stats: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var statsResp []struct {
		Channel string `json:"channel"`
		SendOK  bool   `json:"sendOk"`
		Count   int    `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &statsResp); err != nil {
		t.Fatal(err)
	}
	if len(statsResp) == 0 {
		t.Fatal("admin/stats: expected at least 1 stat row")
	}
}

// TestNotificationCursorPagination verifies cursor-based pagination returns
// consistent results without skipping or duplicating notifications.
func TestNotificationCursorPagination(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	// The demo seed provides notifications. Query with a small limit to exercise pagination.
	// Use grouped=false + no category filter to get all individual entries.

	// Page 1: limit=2
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?limit=2&grouped=false", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("page 1: expected 200, got %d", rec.Code)
	}

	var page1 struct {
		Items      []struct{ ID string } `json:"items"`
		NextCursor *string               `json:"nextCursor"`
		HasMore    bool                  `json:"hasMore"`
	}
	json.Unmarshal(rec.Body.Bytes(), &page1)

	if len(page1.Items) != 2 {
		t.Fatalf("page 1: expected 2 items, got %d", len(page1.Items))
	}
	if !page1.HasMore {
		t.Fatal("page 1: expected hasMore=true (seed has 4+ active notifications)")
	}
	if page1.NextCursor == nil {
		t.Fatal("page 1: expected nextCursor to be set")
	}

	// Page 2: use cursor from page 1.
	req = httptest.NewRequest(http.MethodGet, "/api/notifications?limit=2&grouped=false&cursor="+*page1.NextCursor, nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("page 2: expected 200, got %d", rec.Code)
	}

	var page2 struct {
		Items      []struct{ ID string } `json:"items"`
		NextCursor *string               `json:"nextCursor"`
		HasMore    bool                  `json:"hasMore"`
	}
	json.Unmarshal(rec.Body.Bytes(), &page2)

	if len(page2.Items) == 0 {
		t.Fatal("page 2: expected at least 1 item")
	}

	// Verify no duplicates between pages.
	seen := make(map[string]bool)
	for _, item := range page1.Items {
		seen[item.ID] = true
	}
	for _, item := range page2.Items {
		if seen[item.ID] {
			t.Fatalf("duplicate notification %s across pages", item.ID)
		}
	}
}
