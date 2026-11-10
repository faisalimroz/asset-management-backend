package middleware

import (
	"log"
	"net/http"

	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ErrorMiddleware catches panics and formats centralized errors
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERED] %v", r)
				utils.SendError(c, http.StatusInternalServerError, "internal server error", "An unexpected error occurred")
				c.Abort()
			}
		}()

		c.Next()

		// Check if errors were registered on Gin context
		if len(c.Errors) > 0 {
			lastErr := c.Errors.Last().Err
			if appErr, ok := utils.AsAppError(lastErr); ok {
				utils.SendError(c, appErr.Code, appErr.Message, nil)
				return
			}
			utils.SendError(c, http.StatusInternalServerError, lastErr.Error(), nil)
		}
	}
}

// CORSMiddleware configures standard Cross-Origin Resource Sharing headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
