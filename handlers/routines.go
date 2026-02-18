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

type RoutineHandler struct {
	DB *mongo.Client
}

func (h *RoutineHandler) GetAll(c *gin.Context) {
	collection := h.DB.Database("gym-app").Collection("routines")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var routines []models.Routine
	if err := cursor.All(ctx, &routines); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, routines)
}

func (h *RoutineHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var routine models.Routine
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&routine)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, routine)
}

func (h *RoutineHandler) CreateRoutine(c *gin.Context) {
	var routine models.Routine
	if err := c.ShouldBindJSON(&routine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// add user_id from context to routine
	if userID, exists := c.Get("user_id"); exists {
		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
			return
		}
		userObjID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		routine.UserID = userObjID
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}
	result, err := collection.InsertOne(ctx, routine)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	routine.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, routine)
}

func (h *RoutineHandler) DeleteRoutine(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Routine deleted"})
}

func (h *RoutineHandler) UpdateRoutine(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var routine models.Routine
	if err := c.ShouldBindJSON(&routine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// add user_id from context to routine
	if userID, exists := c.Get("user_id"); exists {
		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
			return
		}
		userObjID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		routine.UserID = userObjID
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": routine})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Routine updated"})
}
