package helpers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sms/backend/internal/http/helpers"
)

// buildRequestWithParam creates a request routed through chi so URLParam works.
func buildRequestWithParam(key, value string) *http.Request {
	r := chi.NewRouter()
	var captured *http.Request
	r.Get("/{"+key+"}", func(w http.ResponseWriter, req *http.Request) {
		captured = req
	})
	req := httptest.NewRequest(http.MethodGet, "/"+value, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return captured
}

func TestParamUUID_Valid(t *testing.T) {
	t.Parallel()
	id := uuid.Must(uuid.NewV7())
	req := buildRequestWithParam("id", id.String())
	got := helpers.ParamUUID(req, "id")
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestParamUUID_Invalid(t *testing.T) {
	t.Parallel()
	req := buildRequestWithParam("id", "not-a-uuid")
	got := helpers.ParamUUID(req, "id")
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

func TestParamUUID_Missing(t *testing.T) {
	t.Parallel()
	// request without chi context at all
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := helpers.ParamUUID(req, "id")
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

func TestParamInt_Valid(t *testing.T) {
	t.Parallel()
	req := buildRequestWithParam("page", "42")
	got := helpers.ParamInt(req, "page")
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestParamInt_Invalid(t *testing.T) {
	t.Parallel()
	req := buildRequestWithParam("page", "abc")
	got := helpers.ParamInt(req, "page")
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestQueryInt_Present(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/?page=5", nil)
	got := helpers.QueryInt(req, "page", 1)
	if got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestQueryInt_Missing(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := helpers.QueryInt(req, "page", 1)
	if got != 1 {
		t.Fatalf("expected default 1, got %d", got)
	}
}

func TestQueryInt_NonNumeric(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)
	got := helpers.QueryInt(req, "page", 1)
	if got != 1 {
		t.Fatalf("expected default 1, got %d", got)
	}
}

func TestQueryInt_Zero(t *testing.T) {
	t.Parallel()
	// 0 is < 1, so should return default
	req := httptest.NewRequest(http.MethodGet, "/?page=0", nil)
	got := helpers.QueryInt(req, "page", 1)
	if got != 1 {
		t.Fatalf("expected default 1 for zero value, got %d", got)
	}
}

func TestQueryInt_Negative(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/?page=-3", nil)
	got := helpers.QueryInt(req, "page", 1)
	if got != 1 {
		t.Fatalf("expected default 1 for negative value, got %d", got)
	}
}

func TestQueryUUID_Valid(t *testing.T) {
	t.Parallel()
	id := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest(http.MethodGet, "/?supplier="+id.String(), nil)
	got := helpers.QueryUUID(req, "supplier")
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestQueryUUID_Missing(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := helpers.QueryUUID(req, "supplier")
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

func TestQueryUUID_Invalid(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/?supplier=garbage", nil)
	got := helpers.QueryUUID(req, "supplier")
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}
