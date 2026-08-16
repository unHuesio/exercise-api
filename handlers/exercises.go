package handlers

import (
	"net/http"

	"gym-api/m/models"
	"gym-api/m/storage"

	"github.com/gin-gonic/gin"
)

type ExerciseHandler struct {
	Store storage.ExerciseStore
}

func (h *ExerciseHandler) GetAll(c *gin.Context) {
	filter := map[string]string{
		"focus":  c.Query("focus"),
		"type":   c.Query("type"),
		"muscle": c.Query("muscle"),
	}

	exercises, err := h.Store.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exercises)
}

func (h *ExerciseHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	exercise, err := h.Store.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exercise)
}

func (h *ExerciseHandler) Create(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body is empty"})
		return
	}

	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing required fields: " + err.Error()})
		return
	}

	id, err := h.Store.Create(c.Request.Context(), exercise)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *ExerciseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Store.Update(c.Request.Context(), id, exercise); err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exercise updated"})
}

func (h *ExerciseHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Store.Delete(c.Request.Context(), id); err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exercise deleted"})
}
