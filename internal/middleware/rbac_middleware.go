package middleware

import (
	"fmt"
	"net/http"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RequireRoles restricts endpoint access to specified roles
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized", "User role not found in context")
			c.Abort()
			return
		}

		userRole := roleVal.(string)
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		utils.SendError(
			c,
			http.StatusForbidden,
			"access forbidden: insufficient permissions",
			fmt.Sprintf("Your role '%s' does not have permission to access this resource", userRole),
		)
		c.Abort()
	}
}

// RequireIT shortcut for IT-only routes
func RequireIT() gin.HandlerFunc {
	return RequireRoles(models.RoleIT)
}

// RequireITOrHR shortcut for routes accessible by IT and HR
func RequireITOrHR() gin.HandlerFunc {
	return RequireRoles(models.RoleIT, models.RoleHR)
}
