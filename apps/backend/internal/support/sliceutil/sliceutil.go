// Package sliceutil provides generic slice helpers used across layers.
package sliceutil

// Paginate returns a page of items with safe bounds.
func Paginate[T any](items []T, page, pageSize int) (result []T, total, safePage, safeSize int) {
	safePage = page
	if safePage < 1 {
		safePage = 1
	}
	safeSize = pageSize
	if safeSize < 1 {
		safeSize = 1
	}
	total = len(items)
	start := (safePage - 1) * safeSize
	if start >= total {
		return []T{}, total, safePage, safeSize
	}
	end := start + safeSize
	if end > total {
		end = total
	}
	return items[start:end], total, safePage, safeSize
}

// HasAny reports whether have includes any of required, or includes "*".
func HasAny(have []string, required ...string) bool {
	if len(required) == 0 {
		return true
	}
	for _, p := range have {
		if p == "*" {
			return true
		}
	}
	for _, req := range required {
		if contains(have, req) {
			return true
		}
	}
	return false
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
