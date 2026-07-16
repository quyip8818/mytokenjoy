package ratelimit

import (
	"net/http"
	"strconv"
	"time"
)

// WriteHeaders sets standard X-RateLimit-* response headers.
func WriteHeaders(w http.ResponseWriter, r Result) {
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(r.Limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(r.Remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(r.ResetAt.Unix(), 10))
}

// WriteRejection sends a 429 Too Many Requests response with Retry-After header.
func WriteRejection(w http.ResponseWriter, r Result) {
	retryAfter := time.Until(r.ResetAt).Seconds()
	if retryAfter < 1 {
		retryAfter = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter)))
	http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
}
