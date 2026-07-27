//go:build testhook

package notification_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

func seedNotification(t *testing.T, router http.Handler, cookie string, entry types.NotificationLogEntry) {
	t.Helper()
	// Use admin test-send endpoint to seed a notification via dispatch.
	// But for direct DB seeding we'll use the store approach.
	// Actually, the simplest: just POST to admin/test with appropriate payload.
	body, _ := json.Marshal(map[string]any{
		"userId":    entry.UserID.String(),
		"eventType": entry.EventType,
		"title":     entry.Title,
		"body":      entry.Body,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/admin/test", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("seedNotification failed: %d %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationInbox_ListEmpty(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	// Use member2 which should have no notifications initially
	cookie := testhttp.AdminCookie(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Items   []json.RawMessage `json:"items"`
		HasMore bool              `json:"hasMore"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	// Items may or may not be empty depending on seed, but response structure should be valid
	if resp.Items == nil {
		t.Fatal("items should not be nil (should be empty array)")
	}
}

func TestNotificationInbox_UnreadCount(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Count < 0 {
		t.Fatalf("count should be >= 0, got %d", resp.Count)
	}
}

func TestNotificationInbox_FullLifecycle(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	// 1. Seed a notification via admin test endpoint
	seedNotification(t, router, cookie, types.NotificationLogEntry{
		UserID:    contract.IDMemberAdmin,
		EventType: "budget_alert_reached",
		Title:     "Test Budget Alert",
		Body:      "Budget reached 90%",
	})

	// Small delay for async processing
	time.Sleep(50 * time.Millisecond)

	// 2. List should return items
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var listResp struct {
		Items []struct {
			ID        string  `json:"id"`
			EventType string  `json:"eventType"`
			Category  string  `json:"category"`
			Title     string  `json:"title"`
			ReadAt    *string `json:"readAt"`
		} `json:"items"`
		HasMore bool `json:"hasMore"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatal(err)
	}

	if len(listResp.Items) == 0 {
		t.Fatal("expected at least 1 notification after seed")
	}

	notifID := listResp.Items[0].ID

	// Verify category was populated
	if listResp.Items[0].Category == "" {
		t.Fatal("expected category to be populated")
	}

	// 3. Mark read
	req = httptest.NewRequest(http.MethodPatch, "/api/notifications/"+notifID+"/read", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("mark read: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Archive
	req = httptest.NewRequest(http.MethodPost, "/api/notifications/"+notifID+"/archive", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("archive: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. Verify it's gone from inbox
	req = httptest.NewRequest(http.MethodGet, "/api/notifications?limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var afterArchive struct {
		Items []struct{ ID string } `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &afterArchive)
	for _, item := range afterArchive.Items {
		if item.ID == notifID {
			t.Fatal("archived notification should not appear in inbox")
		}
	}

	// 6. Verify it appears in archived tab
	req = httptest.NewRequest(http.MethodGet, "/api/notifications?archived=true&limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var archivedResp struct {
		Items []struct{ ID string } `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &archivedResp)
	found := false
	for _, item := range archivedResp.Items {
		if item.ID == notifID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("archived notification should appear in archived list")
	}

	// 7. Unarchive
	req = httptest.NewRequest(http.MethodPost, "/api/notifications/"+notifID+"/unarchive", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("unarchive: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}

	// 8. Soft delete
	req = httptest.NewRequest(http.MethodPost, "/api/notifications/"+notifID+"/delete", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("soft delete: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}

	// 9. Should not appear in inbox anymore
	req = httptest.NewRequest(http.MethodGet, "/api/notifications?limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var afterDelete struct {
		Items []struct{ ID string } `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &afterDelete)
	for _, item := range afterDelete.Items {
		if item.ID == notifID {
			t.Fatal("deleted notification should not appear")
		}
	}

	// 10. Undelete
	req = httptest.NewRequest(http.MethodPost, "/api/notifications/"+notifID+"/undelete", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("undelete: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}

	// 11. Should be back in inbox
	req = httptest.NewRequest(http.MethodGet, "/api/notifications?limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var afterUndelete struct {
		Items []struct{ ID string } `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &afterUndelete)
	found = false
	for _, item := range afterUndelete.Items {
		if item.ID == notifID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("undeleted notification should reappear in inbox")
	}
}

func TestNotificationInbox_CategoryFilter(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	// Seed notifications of different categories
	seedNotification(t, router, cookie, types.NotificationLogEntry{
		UserID:    contract.IDMemberAdmin,
		EventType: "budget_alert_reached",
		Title:     "Budget Alert",
	})
	seedNotification(t, router, cookie, types.NotificationLogEntry{
		UserID:    contract.IDMemberAdmin,
		EventType: "key_expiring_soon",
		Title:     "Key Expiring",
	})
	time.Sleep(50 * time.Millisecond)

	// Filter by budget_alert
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?category=budget_alert&limit=10", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Items []struct {
			Category string `json:"category"`
		} `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	for _, item := range resp.Items {
		if item.Category != "budget_alert" {
			t.Fatalf("expected only budget_alert items, got %q", item.Category)
		}
	}
}

func TestNotificationInbox_InvalidID(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	cookie := testhttp.AdminCookie(t)

	// Mark read with invalid ID
	req := httptest.NewRequest(http.MethodPatch, "/api/notifications/not-a-uuid/read", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", rec.Code)
	}

	// Archive with non-existent ID should still return OK (no-op)
	fakeID := uuid.Must(uuid.NewV7()).String()
	req = httptest.NewRequest(http.MethodPost, "/api/notifications/"+fakeID+"/archive", nil)
	req.Header.Set("Cookie", cookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	// Should not error — just no rows affected
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("expected 2xx for non-existent ID archive, got %d", rec.Code)
	}
}
