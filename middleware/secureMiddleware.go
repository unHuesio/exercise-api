package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecureHeadersMiddleware sets baseline security headers and CORS handling.
// allowedOrigins is an explicit allowlist of origins permitted to make
// credentialed cross-origin requests; reflecting an arbitrary Origin header
// with Access-Control-Allow-Credentials would let any site read authenticated
// responses, so origins outside the allowlist never receive CORS headers.
func SecureHeadersMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}

	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		cspPolicy := "default-src 'self'; connect-src *; font-src *; " +
			"script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';"
		c.Header("Content-Security-Policy", cspPolicy)
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		permPolicy := "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=()," +
			"magnetometer=(),gyroscope=(),fullscreen=(self),payment=()"
		c.Header("Permissions-Policy", permPolicy)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key")
		c.Header("Vary", "Origin")

		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		if origin := c.GetHeader("Origin"); origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
