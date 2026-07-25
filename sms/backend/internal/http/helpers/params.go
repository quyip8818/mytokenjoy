package helpers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ParamUUID extracts a UUID path parameter. Returns uuid.Nil on parse failure.
func ParamUUID(r *http.Request, key string) uuid.UUID {
	id, _ := uuid.Parse(chi.URLParam(r, key))
	return id
}

// ParamInt extracts an integer path parameter. Returns 0 on parse failure.
func ParamInt(r *http.Request, key string) int {
	n, _ := strconv.Atoi(chi.URLParam(r, key))
	return n
}

// QueryInt extracts an integer query parameter with a default value.
// Returns def if the value is missing, non-numeric, or less than 1.
func QueryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}

// QueryUUID extracts a UUID query parameter. Returns uuid.Nil if missing or invalid.
func QueryUUID(r *http.Request, key string) uuid.UUID {
	v := r.URL.Query().Get(key)
	if v == "" {
		return uuid.Nil
	}
	id, _ := uuid.Parse(v)
	return id
}
