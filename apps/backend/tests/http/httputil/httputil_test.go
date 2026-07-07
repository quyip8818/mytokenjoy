package httputil_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/httputil"
)

func TestWriteError_DomainError(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	err := domain.NewDomainError(400, "bad request")
	httputil.WriteError(w, err)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["message"] != "bad request" {
		t.Errorf("message = %q, want 'bad request'", body["message"])
	}
}

func TestWriteError_DomainErrorWithRetryAfter(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	err := domain.NewDomainErrorWithRetryAfter(429, "too many requests", 30)
	httputil.WriteError(w, err)

	if w.Code != 429 {
		t.Errorf("status = %d, want 429", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["message"] != "too many requests" {
		t.Errorf("message = %v, want 'too many requests'", body["message"])
	}
	if body["retryAfter"] != float64(30) {
		t.Errorf("retryAfter = %v, want 30", body["retryAfter"])
	}
}

func TestWriteError_GenericError(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteError(w, errors.New("something went wrong"))

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["message"] != httputil.MsgInternal {
		t.Errorf("message = %q, want %q", body["message"], httputil.MsgInternal)
	}
}

func TestWriteError_Nil(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteError(w, nil)
	if w.Code != 200 {
		t.Errorf("nil error should not write anything, got status %d", w.Code)
	}
}

func TestDecodeJSON_Valid(t *testing.T) {
	t.Parallel()
	body := strings.NewReader(`{"name":"test"}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)
	var dst struct {
		Name string `json:"name"`
	}
	err := httputil.DecodeJSON(r, &dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.Name != "test" {
		t.Errorf("name = %q, want 'test'", dst.Name)
	}
}

func TestDecodeJSON_Invalid(t *testing.T) {
	t.Parallel()
	body := strings.NewReader(`{invalid`)
	r := httptest.NewRequest(http.MethodPost, "/", body)
	var dst struct{}
	err := httputil.DecodeJSON(r, &dst)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	var domErr *domain.DomainError
	if !errors.As(err, &domErr) {
		t.Fatal("expected DomainError")
	}
	if domErr.Status != 400 {
		t.Errorf("status = %d, want 400", domErr.Status)
	}
}

func TestWriteJSON_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteJSON(w, 200, map[string]string{"ok": "true"}, nil)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWriteJSON_Error(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteJSON(w, 200, nil, domain.NewDomainError(404, "not found"))
	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestWriteOK(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteOK(w, map[string]int{"count": 5})
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var body map[string]int
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["count"] != 5 {
		t.Errorf("count = %d, want 5", body["count"])
	}
}

func TestWriteVoid_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteVoid(w, nil)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWriteVoid_Error(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteVoid(w, domain.NewDomainError(403, "forbidden"))
	if w.Code != 403 {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestWriteStatus(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httputil.WriteStatus(w, 401, "Unauthorized")
	if w.Code != 401 {
		t.Errorf("status = %d, want 401", w.Code)
	}
}
