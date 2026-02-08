package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func InferObjectAction() gin.HandlerFunc {
	return func(c *gin.Context) {
		var action string
		switch c.Request.Method {
		case "GET":
			action = "read"
		case "POST":
			action = "create"
		case "PUT", "PATCH":
			action = "update"
		case "DELETE":
			action = "delete"
		default:
			action = "unknown"
		}

		// Infer object from request path
		path := c.FullPath()
		object := ""
		if path != "" {
			// Extract object from path, e.g. /exercises/:id -> exercises
			parts := strings.Split(path, "/")
			if len(parts) > 1 {
				object = parts[1]
			}
		}

		if object == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "Unable to infer object from path"})
			return
		}

		// Store inferred object and action in context for later use
		c.Set("inferred_object", object)
		c.Set("inferred_action", action)
		c.Next()
	}
}
