package utils

import (
	"math"

	"github.com/gin-gonic/gin"
)

// StandardResponse standard API response structure
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationMeta pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse standard paginated response structure
type PaginatedResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// ErrorDetail error object details
type ErrorDetail struct {
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorResponse standard error response structure
type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   ErrorDetail `json:"error"`
}

// SendSuccess sends a standard success JSON response
func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendPaginated sends a standard paginated JSON response
func SendPaginated(c *gin.Context, statusCode int, message string, data interface{}, page, limit int, total int64) {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(statusCode, PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// SendError sends a standard error JSON response
func SendError(c *gin.Context, statusCode int, message string, details interface{}) {
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Message: message,
		Error: ErrorDetail{
			Code:    statusCode,
			Details: details,
		},
	})
}
