package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"gym-api/m/models"
	"gym-api/m/storage"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RoutineHandler struct {
	DB            *mongo.Client
	ExerciseStore storage.ExerciseStore
	saveRoutine   func(context.Context, models.Routine) (models.Routine, error)
}

func (h *RoutineHandler) GetAll(c *gin.Context) {
	userID, err := routineUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	page, limit := pagination(c)
	cursor, err := collection.Find(ctx, routineOwnerFilter(userID), options.Find().
		SetSkip((page-1)*limit).SetLimit(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			_ = c.Error(err)
		}
	}()

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
	userID, err := routineUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var routine models.Routine
	err = collection.FindOne(ctx, routineIDOwnerFilter(objectID, userID)).Decode(&routine)
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
	var input routineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := routineUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	routine := models.Routine{
		Name:        input.Name,
		Description: input.Description,
		Exercises:   input.Exercises,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	routine, err = h.persistRoutine(ctx, routine)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, routine)
}

func (h *RoutineHandler) DeleteRoutine(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	userID, err := routineUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, routineIDOwnerFilter(objectID, userID))
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

	var input routineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := routineUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("routines")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"name":        input.Name,
		"description": input.Description,
		"exercises":   input.Exercises,
		"updated_at":  time.Now(),
	}
	result, err := collection.UpdateOne(ctx, routineIDOwnerFilter(objectID, userID), bson.M{"$set": update})
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

type routineInput struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Exercises   []models.RoutineExercise `json:"exercises"`
}

func routineUserID(c *gin.Context) (primitive.ObjectID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return primitive.NilObjectID, errors.New("user ID not found in context")
	}
	userIDString, ok := userID.(string)
	if !ok {
		return primitive.NilObjectID, errors.New("user ID in context is not a string")
	}
	objectID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid user ID format")
	}
	return objectID, nil
}

func routineOwnerFilter(userID primitive.ObjectID) bson.M {
	return bson.M{"user_id": userID}
}

func routineIDOwnerFilter(routineID, userID primitive.ObjectID) bson.M {
	return bson.M{"_id": routineID, "user_id": userID}
}

func routineFromRecommendation(recommendation *RecommendationRoutine, userID primitive.ObjectID) (models.Routine, error) {
	exercises := make([]models.RoutineExercise, 0, len(recommendation.Exercises))
	for _, recommendationExercise := range recommendation.Exercises {
		exerciseID, err := primitive.ObjectIDFromHex(recommendationExercise.ID)
		if err != nil {
			return models.Routine{}, errors.New("recommended exercise has an invalid ID")
		}
		sets := make([]models.Set, len(recommendationExercise.Sets))
		for index, set := range recommendationExercise.Sets {
			sets[index] = models.Set{Reps: set.Reps, Rest: set.Rest}
		}
		exercises = append(exercises, models.RoutineExercise{
			ExerciseID: exerciseID,
			Sets:       sets,
			Order:      recommendationExercise.Order,
		})
	}

	return models.Routine{
		Name:        recommendation.Name,
		Description: recommendation.Description,
		Exercises:   exercises,
		UserID:      userID,
		CreatedAt:   recommendation.CreatedAt,
		UpdatedAt:   recommendation.UpdatedAt,
	}, nil
}

func (h *RoutineHandler) persistRoutine(ctx context.Context, routine models.Routine) (models.Routine, error) {
	if h.saveRoutine != nil {
		return h.saveRoutine(ctx, routine)
	}
	if h.DB == nil {
		return models.Routine{}, errors.New("routine database is not configured")
	}

	result, err := h.DB.Database("gym-app").Collection("routines").InsertOne(ctx, routine)
	if err != nil {
		return models.Routine{}, err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return models.Routine{}, errors.New("routine insert returned an invalid ID")
	}
	routine.ID = id
	return routine, nil
}
