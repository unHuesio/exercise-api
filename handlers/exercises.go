package handlers

import (
	"net/http"
	"strconv"

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

	page, limit := pagination(c)
	var exercises []models.Exercise
	var err error
	if pager, ok := h.Store.(storage.ExercisePager); ok {
		result, pageErr := pager.ListPage(c.Request.Context(), filter, page, limit)
		exercises = result.Items
		err = pageErr
	} else {
		exercises, err = h.Store.List(c.Request.Context(), filter)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exercises)
}

func pagination(c *gin.Context) (int64, int64) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
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
