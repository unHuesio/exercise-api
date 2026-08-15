package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth(jwtMiddleware gin.HandlerFunc, apiKeyMiddleware gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		apiKeyHeader := c.GetHeader("x-api-key")

		if authHeader != "" {
			jwtMiddleware(c)
			if c.IsAborted() {
				return
			}
			c.Next()
			return
		}

		if apiKeyHeader != "" {
			apiKeyMiddleware(c)
			if c.IsAborted() {
				return
			}
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header or API key required"})
	}
}
