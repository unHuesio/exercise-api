package main

import (
	"log"
	"net/http"

	"gym-api/m/config"
	"gym-api/m/db"
	"gym-api/m/handlers"
	"gym-api/m/middleware"

	limit "github.com/aviddiviner/gin-limit"
	"github.com/casbin/casbin/v2"
	mongodbadapter "github.com/casbin/mongodb-adapter/v4"
	"github.com/gin-gonic/gin"
	limiterlib "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
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
	routineHandler := &handlers.RoutineHandler{DB: client}

	// Rate limiter setup
	rate, err := limiterlib.NewRateFromFormatted("1-S")
	if err != nil {
		log.Fatal(err)
	}
	store := memory.NewStore()
	rateLimiter := limiterlib.New(store, rate)
	rateLimiterMiddleware := mgin.NewMiddleware(rateLimiter)

	// Setup router
	r := gin.Default()

	// Apply API key middleware to all routes
	r.Use(middleware.SecureHeadersMiddleware())
	// Apply rate limiting middleware globally
	r.Use(limit.MaxAllowed(1))
	r.Use(rateLimiterMiddleware)

	// Public routes
	r.POST("/register", authenticationHandler.Register)
	r.POST("/login", authenticationHandler.Login)
	r.POST("/applications/token", authenticationHandler.GenerateApplicationJWT)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Routes
	protected := r.Group("/")
	protected.Use(middleware.Auth(middleware.JWTAuthMiddleware(), middleware.APIKeyAuthMiddleware(apiKeyHandler)))
	protected.Use(middleware.InferObjectAction())
	protected.Use(middleware.Authorize(enforcer, nil))

	protected.GET("/exercises", exerciseHandler.GetAll)
	protected.GET("/exercises/:id", exerciseHandler.GetByID)
	protected.POST("/exercises", exerciseHandler.Create)
	protected.PUT("/exercises/:id", exerciseHandler.Update)
	protected.DELETE("/exercises/:id", exerciseHandler.Delete)

	protected.GET("/routines", routineHandler.GetAll)
	protected.GET("/routines/:id", routineHandler.GetByID)
	protected.POST("/routines", routineHandler.CreateRoutine)
	protected.PUT("/routines/:id", routineHandler.UpdateRoutine)
	protected.DELETE("/routines/:id", routineHandler.DeleteRoutine)

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

	protected.GET("/applications", authenticationHandler.GetApplications)
	protected.POST("/applications", authenticationHandler.RegisterApplication)
	protected.PUT("/applications/:id/status", authenticationHandler.UpdateApplicationStatus)
	protected.DELETE("/applications/:id", authenticationHandler.DeleteApplication)

	r.Run() // listen and serve on 0.0.0.0:8080 by default
}
