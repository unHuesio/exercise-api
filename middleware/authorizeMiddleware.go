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
		if user_email != user && c.GetHeader("X-API-Key") != "" && c.GetHeader("Authorization") != "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User email does not match API key user"})
			return
		}
		object := c.GetString("inferred_object")
		action := c.GetString("inferred_action")

		var inferred_user string
		if exists && user_email != "" && user_email != nil {
			inferred_user, _ = user_email.(string)
		} else {
			inferred_user = user
		}

		fmt.Printf("Authorizing user '%s' for action '%s' on object '%s'\n", inferred_user, action, object)

		allowed, err := enforcer.Enforce(inferred_user, object, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Authorization error"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}
