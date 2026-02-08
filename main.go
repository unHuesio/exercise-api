package main

import (
	"log"
	"net/http"

	"gym-api/m/config"
	"gym-api/m/db"
	"gym-api/m/handlers"
	"gym-api/m/middleware"

	"github.com/casbin/casbin/v2"
	mongodbadapter "github.com/casbin/mongodb-adapter/v4"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	client := db.Connect(cfg.MongoURI)
	defer db.Disconnect(client)

	// Initialize casbin adapter and enforcer
	adapter, err := mongodbadapter.NewAdapter(cfg.MongoURI) // Your MongoDB URL.
	if err != nil {
		log.Fatal(err)
	}
	enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
	if err != nil {
		log.Fatal(err)
	}
	// Load policies from MongoDB
	if err := enforcer.LoadPolicy(); err != nil {
		log.Fatal(err)
	}

	// Initialize handlers
	exerciseHandler := &handlers.ExerciseHandler{DB: client}
	apiKeyHandler := &handlers.APIKeyHandler{DB: client}
	permissionHandler := &handlers.PermissionHandler{DB: client, Enforcer: enforcer}

	// Setup router
	r := gin.Default()

	// Apply API key middleware to all routes
	r.Use(middleware.SecureHeadersMiddleware())
	r.Use(middleware.APIKeyAuthMiddleware(apiKeyHandler))
	r.Use(middleware.InferObjectAction())
	//r.Use(middleware.Authorize(enforcer, nil))

	// Routes
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/exercises", exerciseHandler.GetAll)
	r.GET("/exercises/:id", exerciseHandler.GetByID)
	r.POST("/exercises", exerciseHandler.Create)
	r.PUT("/exercises/:id", exerciseHandler.Update)
	r.DELETE("/exercises/:id", exerciseHandler.Delete)

	r.GET("/api-keys", apiKeyHandler.GetAll)
	r.GET("/api-keys/:account", apiKeyHandler.GetByAccount)
	r.GET("/api-keys/validate/:api_key", apiKeyHandler.Validate)
	r.POST("/api-keys", apiKeyHandler.Create)
	r.PUT("/api-keys/:id/invalidate", apiKeyHandler.Invalidate)
	r.DELETE("/api-keys/:id", apiKeyHandler.Delete)

	r.GET("/permissions", permissionHandler.GetPermissions)
	r.GET("/permissions/role/:subject", permissionHandler.GetPermissionsBySubject)
	r.POST("/permissions", permissionHandler.CreatePermission)
	r.DELETE("/permissions", permissionHandler.DeletePermission)

	r.POST("/permissions/groups", permissionHandler.AssignUserToRole)
	r.DELETE("/permissions/groups", permissionHandler.RemoveUserFromRole)

	r.Run() // listen and serve on 0.0.0.0:8080 by default
}
