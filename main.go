package main

import (
	"context"
	"log"
	"net/http"

	"gym-api/m/config"
	"gym-api/m/db"
	"gym-api/m/handlers"
	"gym-api/m/middleware"
	"gym-api/m/storage"

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

	exerciseStore, err := storage.NewMongoExerciseStore(client, "gym-app")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize handlers
	exerciseHandler := &handlers.ExerciseHandler{Store: exerciseStore}
	permissionHandler := &handlers.PermissionHandler{DB: client, Enforcer: enforcer}
	authenticationHandler := &handlers.AuthenticationHandler{DB: client, Enforcer: enforcer}
	routineHandler := &handlers.RoutineHandler{DB: client, ExerciseStore: exerciseStore}

	// Rate limiter setup
	authRate, err := limiterlib.NewRateFromFormatted("10-M")
	if err != nil {
		log.Fatal(err)
	}
	generalRate, err := limiterlib.NewRateFromFormatted("60-M")
	if err != nil {
		log.Fatal(err)
	}
	store := memory.NewStore()
	authLimiter := limiterlib.New(store, authRate)
	generalLimiter := limiterlib.New(store, generalRate)
	authLimiterMiddleware := mgin.NewMiddleware(authLimiter)
	generalLimiterMiddleware := mgin.NewMiddleware(generalLimiter)

	// Setup router
	r := gin.Default()

	// Apply API key middleware to all routes
	r.Use(middleware.SecureHeadersMiddleware())
	r.Use(limit.MaxAllowed(200))

	// Public auth routes: keep stricter than general API read traffic
	r.POST("/register", authLimiterMiddleware, authenticationHandler.Register)
	r.POST("/login", authLimiterMiddleware, authenticationHandler.Login)

	r.Use(generalLimiterMiddleware)

	// Public routes
	r.GET("/health/storage", func(c *gin.Context) {
		if err := exerciseStore.Health(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"backend": exerciseStore.BackendName(),
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"backend": exerciseStore.BackendName(),
		})
	})

	meRouter := r.Group("/")
	meRouter.Use(middleware.Auth(middleware.JWTAuthMiddleware()))
	meRouter.GET("/me", authenticationHandler.Me)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Routes
	protected := r.Group("/")
	protected.Use(middleware.Auth(middleware.JWTAuthMiddleware()))
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
	protected.POST("/recommendations/routine", routineHandler.Recommend)

	protected.GET("/permissions", permissionHandler.GetPermissions)
	protected.GET("/permissions/role/:subject", permissionHandler.GetPermissionsBySubject)
	protected.POST("/permissions", permissionHandler.CreatePermission)
	protected.DELETE("/permissions", permissionHandler.DeletePermission)

	protected.GET("/permissions/groups", permissionHandler.GetRoles)
	protected.GET("/permissions/groups/:user", permissionHandler.GetRolesByUser)
	protected.POST("/permissions/groups", permissionHandler.AssignUserToRole)
	protected.DELETE("/permissions/groups", permissionHandler.RemoveUserFromRole)

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
