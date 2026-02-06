package main

import (
	"net/http"

	"gym-api/m/config"
	"gym-api/m/db"
	"gym-api/m/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	client := db.Connect(cfg.MongoURI)
	defer db.Disconnect(client)

	// Initialize handlers
	exerciseHandler := &handlers.ExerciseHandler{DB: client}

	// Setup router
	r := gin.Default()

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

	r.Run() // listen and serve on 0.0.0.0:8080 by default
}
