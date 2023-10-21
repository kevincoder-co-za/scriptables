package middleware

import (
	"github.com/gin-gonic/gin"
	"plexcorp.tech/scriptable/models"
)

func DBMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		models.SetDBConnection(c)
		c.Next()
	}
}
