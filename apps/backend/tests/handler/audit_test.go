package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestSettingsUpdateHTTP(t *testing.T) {
	router := newTestRouter(t)
	body := []byte(`{"contentRetentionEnabled":false}`)
	req := httptest.NewRequest(http.MethodPut, "/api/audit/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", sessionCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var settings types.AuditSettings
	if err := json.NewDecoder(rec.Body).Decode(&settings); err != nil {
		t.Fatal(err)
	}
	if settings.ContentRetentionEnabled {
		t.Fatal("expected contentRetentionEnabled false")
	}
}
