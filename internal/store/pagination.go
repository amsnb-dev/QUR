package store

import "strconv"

const (
	DefaultLimit = 50
	MaxLimit     = 200
)

// Page holds validated pagination parameters.
type Page struct {
	Limit  int
	Offset int
}

// ParsePage parses limit/offset from raw query-string values.
// Enforces DefaultLimit and MaxLimit.
func ParsePage(limitStr, offsetStr string) Page {
	limit := DefaultLimit
	offset := 0

	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
		offset = v
	}
	return Page{Limit: limit, Offset: offset}
}

// PagedResult wraps a list with pagination metadata.
type PagedResult[T any] struct {
	Data   []T `json:"data"`
	Meta   Meta `json:"meta"`
}

// Meta holds pagination metadata returned with every list response.
type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}
