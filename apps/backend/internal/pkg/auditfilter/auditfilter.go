package auditfilter

import (
	"strings"
)

func FilterByDateRangeCreatedAt[T any](items []T, from, to string, createdAt func(T) string) []T {
	if from == "" && to == "" {
		return items
	}
	result := make([]T, 0, len(items))
	for _, item := range items {
		day := createdAt(item)
		if len(day) > 10 {
			day = day[:10]
		}
		if from != "" && day < from {
			continue
		}
		if to != "" && day > to {
			continue
		}
		result = append(result, item)
	}
	return result
}

func FilterByEquals[T any, V comparable](items []T, value V, extract func(T) V) []T {
	var zero V
	if value == zero {
		return items
	}
	result := make([]T, 0, len(items))
	for _, item := range items {
		if extract(item) == value {
			result = append(result, item)
		}
	}
	return result
}

func FilterByKeyword[T any](items []T, keyword string, extractors []func(T) string) []T {
	if strings.TrimSpace(keyword) == "" {
		return items
	}
	q := strings.ToLower(strings.TrimSpace(keyword))
	result := make([]T, 0, len(items))
	for _, item := range items {
		for _, extract := range extractors {
			if strings.Contains(strings.ToLower(extract(item)), q) {
				result = append(result, item)
				break
			}
		}
	}
	return result
}
