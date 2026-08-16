package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"gym-api/m/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type RecommendationRequest struct {
	Goal         string   `json:"goal" binding:"required"`
	Days         int      `json:"days" binding:"required"`
	Focus        string   `json:"focus"`
	ExerciseType string   `json:"exerciseType"`
	Experience   string   `json:"experience"`
	AvoidMuscles []string `json:"avoidMuscles"`
}

type RecommendationDay struct {
	Name      string                   `json:"name"`
	Focus     string                   `json:"focus"`
	Exercises []RecommendationExercise `json:"exercises"`
}

type RecommendationExercise struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Sets  int    `json:"sets"`
	Reps  string `json:"reps"`
	Type  string `json:"type"`
	Focus string `json:"focus"`
}

type RecommendationPlan struct {
	Goal string              `json:"goal"`
	Days []RecommendationDay `json:"days"`
}

func BuildRoutineRecommendation(exercises []models.Exercise, goal string, days int, focus string) (*RecommendationPlan, error) {
	goal = strings.ToLower(strings.TrimSpace(goal))
	if goal != "strength" && goal != "powerlifting" {
		return nil, errors.New("goal must be strength or powerlifting")
	}
	if days != 3 && days != 4 {
		return nil, errors.New("days must be 3 or 4")
	}
	if len(exercises) == 0 {
		return nil, errors.New("no exercises available to build a routine")
	}

	filtered := make([]models.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		if strings.TrimSpace(exercise.Exercise) == "" {
			continue
		}
		if strings.TrimSpace(focus) != "" && !strings.EqualFold(exercise.Focus, focus) && !strings.Contains(strings.ToLower(exercise.Exercise), strings.ToLower(focus)) {
			continue
		}
		filtered = append(filtered, exercise)
	}
	if len(filtered) == 0 {
		return nil, errors.New("no exercises matched the requested focus")
	}

	sort.Slice(filtered, func(i, j int) bool {
		return scoreExercise(filtered[i], goal) > scoreExercise(filtered[j], goal)
	})

	dayTemplates := []string{"Lower Body", "Upper Body", "Full Body"}
	if days == 4 {
		dayTemplates = []string{"Lower Body", "Upper Push", "Upper Pull", "Accessory"}
	}

	planDays := make([]RecommendationDay, 0, days)
	base := make([]models.Exercise, 0, len(filtered))
	for _, ex := range filtered {
		base = append(base, ex)
	}

	for i := 0; i < days; i++ {
		dayExercises := make([]RecommendationExercise, 0, 4)
		for j := 0; j < len(base) && j < 3; j++ {
			ex := base[(i+j)%len(base)]
			if len(dayExercises) == 0 || !sameExercise(dayExercises, ex.ID) {
				dayExercises = append(dayExercises, RecommendationExercise{
					ID:    ex.ID,
					Name:  ex.Exercise,
					Sets:  4,
					Reps:  defaultReps(goal),
					Type:  ex.Type,
					Focus: ex.Focus,
				})
			}
		}
		if len(dayExercises) == 0 {
			dayExercises = append(dayExercises, RecommendationExercise{
				ID:    base[0].ID,
				Name:  base[0].Exercise,
				Sets:  4,
				Reps:  defaultReps(goal),
				Type:  base[0].Type,
				Focus: base[0].Focus,
			})
		}
		planDays = append(planDays, RecommendationDay{
			Name:      fmt.Sprintf("Day %d - %s", i+1, dayTemplates[i%len(dayTemplates)]),
			Focus:     strings.TrimSpace(base[0].Focus),
			Exercises: dayExercises,
		})
	}

	return &RecommendationPlan{Goal: goal, Days: planDays}, nil
}

func scoreExercise(ex models.Exercise, goal string) int {
	score := 0
	if strings.Contains(strings.ToLower(ex.Type), "compound") {
		score += 3
	}
	if goal == "powerlifting" {
		if strings.Contains(strings.ToLower(ex.Exercise), "squat") || strings.Contains(strings.ToLower(ex.Exercise), "bench") || strings.Contains(strings.ToLower(ex.Exercise), "deadlift") {
			score += 5
		}
	}
	if goal == "strength" {
		if strings.Contains(strings.ToLower(ex.Exercise), "press") || strings.Contains(strings.ToLower(ex.Exercise), "squat") || strings.Contains(strings.ToLower(ex.Exercise), "deadlift") || strings.Contains(strings.ToLower(ex.Exercise), "row") {
			score += 4
		}
	}
	if strings.EqualFold(ex.Focus, "Leg") {
		score += 2
	}
	if strings.EqualFold(ex.Focus, "Upper") {
		score += 2
	}
	return score
}

func defaultReps(goal string) string {
	if goal == "powerlifting" {
		return "3-5"
	}
	return "5-8"
}

func sameExercise(items []RecommendationExercise, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func filterExercisesForRequest(exercises []models.Exercise, req RecommendationRequest) []models.Exercise {
	filtered := make([]models.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		if strings.TrimSpace(req.Focus) != "" && !strings.EqualFold(exercise.Focus, req.Focus) && !strings.Contains(strings.ToLower(exercise.Exercise), strings.ToLower(req.Focus)) {
			continue
		}
		if strings.TrimSpace(req.ExerciseType) != "" && !strings.EqualFold(exercise.Type, req.ExerciseType) {
			continue
		}
		if len(req.AvoidMuscles) > 0 {
			matched := false
			for _, avoided := range req.AvoidMuscles {
				avoided = strings.ToLower(strings.TrimSpace(avoided))
				if avoided == "" {
					continue
				}
				if strings.Contains(strings.ToLower(exercise.PrimaryMuscles), avoided) || strings.Contains(strings.ToLower(exercise.SecondaryMuscles), avoided) {
					matched = true
					break
				}
			}
			if matched {
				continue
			}
		}
		filtered = append(filtered, exercise)
	}
	return filtered
}

func loadSampleExercises(path string) ([]models.Exercise, error) {
	if strings.TrimSpace(path) == "" {
		path = "sample.json"
	}

	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)
	repoRoot := filepath.Clean(filepath.Join(baseDir, ".."))

	candidatePaths := []string{
		path,
		filepath.Clean(path),
		filepath.Join(baseDir, path),
		filepath.Join(repoRoot, path),
		filepath.Join(repoRoot, "sample.json"),
		filepath.Join(repoRoot, "excercises.json"),
		filepath.Join(repoRoot, "sample"),
		filepath.Join(repoRoot, "excercises"),
		filepath.Join(".", path),
		filepath.Join("..", path),
		filepath.Join("..", "..", path),
	}

	if filepath.IsAbs(path) {
		candidatePaths = []string{path}
	}

	var lastErr error
	for _, candidate := range candidatePaths {
		data, err := os.ReadFile(candidate)
		if err == nil {
			var exercises []models.Exercise
			if err := json.Unmarshal(data, &exercises); err != nil {
				return nil, err
			}
			return exercises, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (h *RoutineHandler) Recommend(c *gin.Context) {
	var req RecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("exercises")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = cursor.Close(ctx) }()

	var exercises []models.Exercise
	if err := cursor.All(ctx, &exercises); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(exercises) == 0 {
		if sample, sampleErr := loadSampleExercises("sample.json"); sampleErr == nil && len(sample) > 0 {
			exercises = sample
		}
	}

	exercises = filterExercisesForRequest(exercises, req)
	plan, err := BuildRoutineRecommendation(exercises, req.Goal, req.Days, req.Focus)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}
