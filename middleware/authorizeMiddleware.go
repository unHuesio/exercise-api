package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func Authorize(enforcer *casbin.Enforcer, getObject func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		object := c.GetString("inferred_object")
		action := c.GetString("inferred_action")
		if object == "" || action == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Authorization context missing"})
			return
		}

		subjects := make([]string, 0, 2)
		if userEmail, exists := c.Get("user_email"); exists {
			if email, ok := userEmail.(string); ok && email != "" {
				subjects = append(subjects, email)
			}
		}
		if apiKeyUser, exists := c.Get("api_key_user"); exists {
			if user, ok := apiKeyUser.(string); ok && user != "" {
				subjects = append(subjects, user)
			}
		}
		if len(subjects) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User identity not found in request context"})
			return
		}

		for _, subject := range subjects {
			allowed, err := enforcer.Enforce(subject, object, action)
			if err == nil && allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	}
}
