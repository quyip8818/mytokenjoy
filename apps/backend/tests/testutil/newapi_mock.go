package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func StartNewAPIMockServer(t *testing.T, logs []newapi.LogEntry) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/log/" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data":    logs,
			})
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func SampleMappedLog(tokenID int64, createdAt time.Time) newapi.LogEntry {
	return newapi.LogEntry{
		ID: tokenID, TokenID: tokenID, Quota: 500000,
		ModelName: "gpt-4o", CreatedAt: createdAt.Unix(),
	}
}
