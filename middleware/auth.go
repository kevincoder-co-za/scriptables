package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"plexscriptables.com/scriptables/models"
)

func isPublicPath(path string) bool {
	return strings.Contains(path, "/users/") ||
		strings.Contains(path, "/webhooks/") ||
		strings.Contains(path, "trial-expired")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		if models.CountUsers(c.MustGet("db").(*gorm.DB)) == 0 {
			c.Redirect(http.StatusFound, "/users/register")
			c.Abort()
			return
		}

		if sessions.Default(c).Get("user_id") == nil {
			c.Redirect(http.StatusFound, "/users/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
