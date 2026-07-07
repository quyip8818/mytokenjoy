package audit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	testhttp "github.com/tokenjoy/backend/tests/testutil/http"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestSettingsUpdateHTTP(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)
	body := []byte(`{"contentRetentionEnabled":false}`)
	req := httptest.NewRequest(http.MethodPut, "/api/audit/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", testhttp.AdminCookie(t))
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
