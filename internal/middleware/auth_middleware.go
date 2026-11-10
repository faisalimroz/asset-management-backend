package middleware

import (
	"net/http"
	"strings"

	"asset-management-backend/internal/config"
	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT bearer token and sets user in context
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "authorization header required", "Missing 'Authorization: Bearer <token>' header")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendError(c, http.StatusUnauthorized, "invalid authorization header format", "Format must be 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "invalid or expired token", err.Error())
			c.Abort()
			return
		}

		// Set context variables
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("department", claims.Department)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetCurrentUser extracts current authenticated user metadata from Gin context
func GetCurrentUser(c *gin.Context) *models.User {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	dept, _ := c.Get("department")

	if userID == nil {
		return nil
	}

	return &models.User{
		ID:         userID.(string),
		Email:      email.(string),
		Role:       role.(string),
		Department: dept.(string),
	}
}
