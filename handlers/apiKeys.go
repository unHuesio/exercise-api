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

	"github.com/google/uuid"
)

type APIKeyHandler struct {
	DB *mongo.Client
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	// Check for empty body
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty request body"})
		return
	}

	var apiKey models.ApiKey
	if err := c.ShouldBindJSON(&apiKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey.APIKey = uuid.New().String()
	apiKey.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	apiKey.IsValid = true

	collection := h.DB.Database("gym-app").Collection("api_keys")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

func (h *APIKeyHandler) Invalidate(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := h.DB.Database("gym-app").Collection("api_keys")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"is_valid": false}}
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key invalidated"})
}

func (h *APIKeyHandler) Validate(c *gin.Context) {
	apiKey := c.Param("api_key")

	isValid, err := h.ValidateApiKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key is valid"})
}

func (h *APIKeyHandler) ValidateApiKey(apiKey string) (bool, error) {
	collection := h.DB.Database("gym-app").Collection("api_keys")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result models.ApiKey
	err := collection.FindOne(ctx, bson.M{"api_key": apiKey, "is_valid": true}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil // API key not found or invalid
		}
		return false, err // Some other error occurred
	}

	return true, nil // API key is valid
}

func (h *APIKeyHandler) GetApiKeyUser(apiKey string) (string, error) {
	collection := h.DB.Database("gym-app").Collection("api_keys")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result models.ApiKey
	err := collection.FindOne(ctx, bson.M{"api_key": apiKey, "is_valid": true}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil // API key not found or invalid
		}
		return "", err // Some other error occurred
	}

	return result.Account, nil // Return the account associated with the API key
}

func (h *APIKeyHandler) GetByAccount(c *gin.Context) {
	account := c.Param("account")

	collection := h.DB.Database("gym-app").Collection("api_keys")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var apiKey models.ApiKey
	err := collection.FindOne(ctx, bson.M{"account": account}).Decode(&apiKey)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found for account"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

func (h *APIKeyHandler) GetAll(c *gin.Context) {
	collection := h.DB.Database("gym-app").Collection("api_keys")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var apiKeys []models.ApiKey
	if err = cursor.All(ctx, &apiKeys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := h.DB.Database("gym-app").Collection("api_keys")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted"})
}
