package middleware

import (
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

func SetupSession() gin.HandlerFunc {

	store, _ := redis.NewStore(10, "tcp", os.Getenv("REDIS_DSN"), "", []byte("zxfS223334Dkq"))
	return sessions.Sessions("scriptable", store)
}
