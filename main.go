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
	authenticationHandler := &handlers.AuthenticationHandler{DB: client, Enforcer: enforcer}

	// Setup router
	r := gin.Default()

	// Apply API key middleware to all routes
	r.Use(middleware.SecureHeadersMiddleware())

	// Public routes
	r.POST("/register", authenticationHandler.Register)
	r.POST("/login", authenticationHandler.Login)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Routes
	protected := r.Group("/")
	protected.Use(middleware.APIKeyAuthMiddleware(apiKeyHandler))
	protected.Use(middleware.JWTAuthMiddleware())
	protected.Use(middleware.InferObjectAction())
	protected.Use(middleware.Authorize(enforcer, nil))

	protected.GET("/exercises", exerciseHandler.GetAll)
	protected.GET("/exercises/:id", exerciseHandler.GetByID)
	protected.POST("/exercises", exerciseHandler.Create)
	protected.PUT("/exercises/:id", exerciseHandler.Update)
	protected.DELETE("/exercises/:id", exerciseHandler.Delete)

	protected.GET("/api-keys", apiKeyHandler.GetAll)
	protected.GET("/api-keys/:account", apiKeyHandler.GetByAccount)
	protected.GET("/api-keys/validate/:api_key", apiKeyHandler.Validate)
	protected.POST("/api-keys", apiKeyHandler.Create)
	protected.PUT("/api-keys/:id/invalidate", apiKeyHandler.Invalidate)
	protected.DELETE("/api-keys/:id", apiKeyHandler.Delete)

	protected.GET("/permissions", permissionHandler.GetPermissions)
	protected.GET("/permissions/role/:subject", permissionHandler.GetPermissionsBySubject)
	protected.POST("/permissions", permissionHandler.CreatePermission)
	protected.DELETE("/permissions", permissionHandler.DeletePermission)

	protected.GET("/permissions/groups", permissionHandler.GetRoles)
	protected.GET("/permissions/groups/:user", permissionHandler.GetRolesByUser)
	protected.POST("/permissions/groups", permissionHandler.AssignUserToRole)
	protected.DELETE("/permissions/groups", permissionHandler.RemoveUserFromRole)

	r.Run() // listen and serve on 0.0.0.0:8080 by default
}
