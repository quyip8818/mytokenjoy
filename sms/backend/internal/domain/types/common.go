package types

import "github.com/google/uuid"

// ========== 通用工具类型 ==========

// IDName 用于下拉选项等场景
type IDName struct {
	ID   uuid.UUID `json:"id"   db:"id"`
	Name string    `json:"name" db:"name"`
}

// PagedResult 通用分页返回结构
type PagedResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}
