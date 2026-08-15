package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth(jwtMiddleware gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		jwtMiddleware(c)
		if c.IsAborted() {
			return
		}

		c.Next()
	}
}
