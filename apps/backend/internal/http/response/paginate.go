package response

import "github.com/tokenjoy/backend/internal/domain/types"

type Paginated[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

func FromPageResult[T any](page types.PageResult[T]) Paginated[T] {
	return Paginated[T]{
		Items:    page.Items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}
}
