package middleware

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func Authorize(enforcer *casbin.Enforcer, getObject func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.GetString("api_key_user")
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
