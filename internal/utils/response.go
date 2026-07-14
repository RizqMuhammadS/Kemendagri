package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/meeting-minutes-ai/internal/dto"
)

// SuccessResponse sends a success API response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, dto.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error API response
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, dto.APIResponse{
		Success: false,
		Error:   message,
	})
}