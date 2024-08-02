package pagination

import (
	"fmt"
	"net/url"
	"strconv"
)

// PagedResponse wraps data with pagination info
type PagedResponse struct {
	Pagination *Pagination `json:"pagination"`
	Data       interface{} `json:"data"`
}

// ApplyToQuery applies pagination to a SQL query
func ApplyToQuery(query string, page, pageSize int) string {
	offset := (page - 1) * pageSize
	return fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pageSize, offset)
}

// Pagination holds pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}

// NewPagination creates a new Pagination instance
func NewPagination(page, pageSize, totalItems int) *Pagination {
	totalPages := (totalItems + pageSize - 1) / pageSize
	return &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}
}

// ParsePaginationParams parses pagination parameters from URL query
func ParsePaginationParams(query url.Values) (int, int, error) {
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(query.Get("page_size"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	return page, pageSize, nil
}
