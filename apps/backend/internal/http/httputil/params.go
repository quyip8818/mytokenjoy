package httputil

import "strconv"

// ParseIntParam parses a query string integer, returning fallback on empty/invalid/non-positive values.
func ParseIntParam(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
