package pagination

import (
	"math"
)

type Query struct {
	Page   int
	Limit  int
	Status string
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func NewPaginationQuery(page, limit int, status string) Query {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return Query{Page: page, Limit: limit, Status: status}
}

func (p Query) Offset() int {
	return (p.Page - 1) * p.Limit
}

func NewPaginatedResponse[T any](data []T, total int64, q Query) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(q.Limit))),
	}
}
