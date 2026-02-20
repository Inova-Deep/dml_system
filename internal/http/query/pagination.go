package query

import (
	"net/http"
	"strconv"
)

// PaginationParams holds common listing parameters
type PaginationParams struct {
	Page   int
	Size   int
	Search string
}

// ParsePagination standardizes the extraction of ?page & ?size & ?search
func ParsePagination(r *http.Request) PaginationParams {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")
	search := r.URL.Query().Get("search")

	page := 1
	if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
		page = parsedPage
	}

	size := 50
	if parsedSize, err := strconv.Atoi(sizeStr); err == nil && parsedSize > 0 {
		if parsedSize > 1000 {
			size = 1000
		} else {
			size = parsedSize
		}
	}

	return PaginationParams{
		Page:   page,
		Size:   size,
		Search: search,
	}
}

// Offset calculates the SQL OFFSET based on Page and Size
func (p PaginationParams) Offset() int32 {
	return int32((p.Page - 1) * p.Size)
}

// Limit returns the SQL LIMIT
func (p PaginationParams) Limit() int32 {
	return int32(p.Size)
}
