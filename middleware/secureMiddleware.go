package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecureHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// if c.Request.Host != expectedHost {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid host header"})
		// 	return
		// }
		c.Header("X-Frame-Options", "DENY")
		cspPolicy := "default-src 'self'; connect-src *; font-src *; " +
			"script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';"
		c.Header("Content-Security-Policy", cspPolicy)
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		permPolicy := "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=()," +
			"magnetometer=(),gyroscope=(),fullscreen=(self),payment=()"
		c.Header("Permissions-Policy", permPolicy)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
