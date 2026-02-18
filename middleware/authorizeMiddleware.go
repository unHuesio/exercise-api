package middleware

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func Authorize(enforcer *casbin.Enforcer, getObject func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// print context keys and values for dbeugging
		for k, v := range c.Keys {
			fmt.Printf("Context key: %s, value: %v\n", k, v)
		}
		user_email, exists := c.Get("user_email")
		if !exists && c.GetHeader("Authorization") != "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User email not found in context"})
			return
		}
		user := c.GetString("api_key_user")

		object := c.GetString("inferred_object")
		action := c.GetString("inferred_action")

		apiKeyAllowed, _ := enforcer.Enforce(user, object, action)
		fmt.Printf("Authorizing API key user '%s' for action '%s' on object '%s'\n", user, action, object)

		userAllowed, _ := enforcer.Enforce(user_email, object, action)
		fmt.Printf("Authorizing user '%s' for action '%s' on object '%s'\n", user_email, action, object)

		if !apiKeyAllowed || (!userAllowed && exists) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		c.Next()
	}
}
