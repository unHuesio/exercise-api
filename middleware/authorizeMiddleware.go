package middleware

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func Authorize(enforcer *casbin.Enforcer, getObject func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		user_email, exists := c.Get("user_email")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User email not found in context"})
			return
		}
		user := c.GetString("api_key_user")

		if user_email != user {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User email does not match API key user"})
			return
		}
		object := c.GetString("inferred_object")
		action := c.GetString("inferred_action")

		fmt.Printf("Authorizing user '%s' for action '%s' on object '%s'\n", user, action, object)

		allowed, err := enforcer.Enforce(user, object, action)
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
