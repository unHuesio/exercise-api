package handlers

import (
	"context"
	"net/http"
	"time"

	"gym-api/m/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ExerciseHandler struct {
	DB *mongo.Client
}

func (h *ExerciseHandler) GetAll(c *gin.Context) {
	collection := h.DB.Database("gym-app").Collection("exercises")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var exercises []models.Exercise
	if err := cursor.All(ctx, &exercises); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exercises)
}

func (h *ExerciseHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := h.DB.Database("gym-app").Collection("exercises")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exercise models.Exercise
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&exercise)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, exercise)
}

func (h *ExerciseHandler) Create(c *gin.Context) {
	// Check for empty body
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body is empty"})
		return
	}

	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing required fields: " + err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("exercises")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, exercise)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID})
}

func (h *ExerciseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("exercises")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": exercise})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exercise updated"})
}

func (h *ExerciseHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	collection := h.DB.Database("gym-app").Collection("exercises")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exercise deleted"})
}
