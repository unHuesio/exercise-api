package middleware

import (
	"gym-api/m/handlers"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware(apiKeyHandler *handlers.APIKeyHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "API key required"})
			return
		}
		// Validate API key
		isValid, err := apiKeyHandler.ValidateApiKey(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if !isValid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid API key"})
			return
		}
		// Store API key user in context for later use
		user, err := apiKeyHandler.GetApiKeyUser(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			return
		}
		c.Set("api_key_user", user)
		c.Next()
	}
}
