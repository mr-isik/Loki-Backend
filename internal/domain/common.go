package domain

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, total, page, pageSize int) *PaginatedResponse {
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginatedResponse{
		Data:       data,
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
	}
}
