package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/users/") ||
			strings.Contains(c.Request.URL.Path, "/webhooks/") ||
			strings.Contains(c.Request.URL.Path, "trial-expired") {
			c.Next()
			return
		}

		allowRegistration, _ := strconv.ParseBool(os.Getenv("ALLOW_REGISTER"))
		if allowRegistration {
			DB, _ := c.Get("db")
			gorm := DB.(*gorm.DB)

			var numUsers int
			gorm.Raw("SELECT count(id) FROM users").Scan(&numUsers)
			if numUsers == 0 {
				c.Redirect(http.StatusFound, "/users/register")
				c.Abort()
				return
			}
		}

		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Redirect(http.StatusFound, "/users/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
