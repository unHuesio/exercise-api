package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth(jwtMiddleware gin.HandlerFunc, apiKeyMiddleware gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		apiKeyHeader := c.GetHeader("x-api-key")

		jwtTried := false
		apiKeyTried := false

		if authHeader != "" {
			jwtMiddleware(c)
			jwtTried = true

			if c.IsAborted() {
				return
			}
		}

		if apiKeyHeader != "" {
			apiKeyMiddleware(c)
			apiKeyTried = true

			if c.IsAborted() {
				return
			}
		}

		if !jwtTried && !apiKeyTried {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header or API key required"})
			return
		}
		fmt.Printf("Auth middleware - continue to next middleware\n")
		c.Next()
	}
}
