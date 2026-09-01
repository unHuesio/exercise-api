package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"gym-api/m/models"

	"github.com/gin-gonic/gin"
)

type RecommendationRequest struct {
	Goal         string   `json:"goal" binding:"required"`
	Days         int      `json:"days" binding:"required"`
	Focus        string   `json:"focus"`
	ExerciseType string   `json:"exerciseType"`
	Experience   string   `json:"experience"`
	AvoidMuscles []string `json:"avoidMuscles"`
}

type RecommendationExercise struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	PrimaryMuscles   string       `json:"primary_muscles"`
	SecondaryMuscles string       `json:"secondary_muscles,omitempty"`
	Type             string       `json:"type"`
	Focus            string       `json:"focus,omitempty"`
	Sets             []models.Set `json:"sets"`
	Order            int          `json:"order"`
	Day              int          `json:"day"`
}

type RecommendationRoutine struct {
	ID          string                   `json:"id,omitempty"`
	Goal        string                   `json:"goal"`
	Days        int                      `json:"days"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Exercises   []RecommendationExercise `json:"exercises"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// generateSetsForExercise creates concrete Set data based on goal and exercise type.
// Returns a slice of Set objects with concrete reps, weight, and rest values.
func generateSetsForExercise(exercise models.Exercise, goal string) []models.Set {
	isCompound := strings.Contains(strings.ToLower(exercise.Type), "compound")

	var numSets, reps, restSeconds int

	if goal == "powerlifting" {
		reps = 4 // 3-5 average
		if isCompound {
			numSets = 5
			restSeconds = 120 // 2 minutes
		} else {
			numSets = 4
			restSeconds = 90
		}
	} else { // strength
		reps = 6 // 5-8 average
		if isCompound {
			numSets = 4
			restSeconds = 120
		} else {
			numSets = 3
			restSeconds = 90
		}
	}

	sets := make([]models.Set, numSets)
	for i := 0; i < numSets; i++ {
		sets[i] = models.Set{
			Reps:   reps,
			Weight: 0, // Default: no weight history available
			Rest:   restSeconds,
		}
	}
	return sets
}

// BuildRoutineRecommendation creates a response with each recommended exercise and its prescription
// based on the provided exercises, goal, and number of training days.
// Ensures minimum 5 exercises per training day.
func BuildRoutineRecommendation(exercises []models.Exercise, goal string, days int, focus string) (*RecommendationRoutine, error) {
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

	filtered := filterExercisesForFocus(exercises, focus)
	if len(filtered) == 0 {
		return nil, errors.New("no exercises matched the requested focus")
	}

	// Sort by relevance score (highest first)
	sort.Slice(filtered, func(i, j int) bool {
		return scoreExercise(filtered[i], goal) > scoreExercise(filtered[j], goal)
	})

	compoundExercises, isolationExercises := partitionExercisesByType(filtered)

	// Distribute exercises across days ensuring minimum 5 per day
	minPerDay := 5
	routineExercises := make([]RecommendationExercise, 0, minPerDay*days)
	exerciseOrder := 0
	usedExerciseIDs := make(map[string]bool)

	for dayIdx := 0; dayIdx < days; dayIdx++ {
		dayExercises := selectExercisesForDay(
			compoundExercises,
			isolationExercises,
			minPerDay,
			dayIdx,
			usedExerciseIDs,
		)

		// Include the source exercise details so clients can render the recommendation directly.
		for _, ex := range dayExercises {
			sets := generateSetsForExercise(ex, goal)
			routineExercises = append(routineExercises, RecommendationExercise{
				ID:               ex.ID,
				Name:             ex.Exercise,
				PrimaryMuscles:   ex.PrimaryMuscles,
				SecondaryMuscles: ex.SecondaryMuscles,
				Type:             ex.Type,
				Focus:            ex.Focus,
				Sets:             sets,
				Order:            exerciseOrder,
				Day:              dayIdx + 1,
			})
			exerciseOrder++
		}
	}

	// Build routine name and description
	goalCapitalized := strings.ToUpper(goal[:1]) + strings.ToLower(goal[1:])
	routineName := fmt.Sprintf("Recommended %s Routine - %d Days", goalCapitalized, days)
	routineDesc := fmt.Sprintf("%d-day %s routine with %d exercises per day", days, goal, minPerDay)
	if len(routineExercises) < minPerDay*days {
		routineDesc = fmt.Sprintf("%d-day %s routine with %d unique exercises", days, goal, len(routineExercises))
	}
	if strings.TrimSpace(focus) != "" {
		routineName += fmt.Sprintf(" (%s)", focus)
		routineDesc = fmt.Sprintf("%d-day %s routine focused on %s", days, goal, focus)
		if len(routineExercises) < minPerDay*days {
			routineDesc += fmt.Sprintf(" with %d unique exercises", len(routineExercises))
		}
	}

	now := time.Now()
	routine := &RecommendationRoutine{
		Name:        routineName,
		Description: routineDesc,
		Goal:        goal,
		Days:        days,
		Exercises:   routineExercises,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return routine, nil
}

func filterExercisesForFocus(exercises []models.Exercise, focus string) []models.Exercise {
	normalizedFocus := normalizeFocus(focus)
	if normalizedFocus == "" {
		return exercises
	}

	accessoryFocus, hasAccessoryFocus := accessoryFocusFor(normalizedFocus)
	filtered := make([]models.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		if strings.TrimSpace(exercise.Exercise) == "" {
			continue
		}
		if !hasAccessoryFocus && matchesFocus(exercise, normalizedFocus) {
			filtered = append(filtered, exercise)
			continue
		}
		if isCompoundExercise(exercise) && matchesFocus(exercise, normalizedFocus) {
			filtered = append(filtered, exercise)
			continue
		}
		if isIsolationExercise(exercise) && normalizeFocus(exercise.Focus) == accessoryFocus {
			filtered = append(filtered, exercise)
		}
	}

	return filtered
}

func accessoryFocusFor(focus string) (string, bool) {
	switch normalizeFocus(focus) {
	case "chest":
		return "tricep", true
	case "back":
		return "bicep", true
	case "leg":
		return "shoulder", true
	default:
		return "", false
	}
}

func matchesFocus(exercise models.Exercise, normalizedFocus string) bool {
	return normalizeFocus(exercise.Focus) == normalizedFocus ||
		strings.Contains(strings.ToLower(exercise.Exercise), normalizedFocus)
}

func normalizeFocus(focus string) string {
	normalized := strings.ToLower(strings.TrimSpace(focus))
	if normalized == "shoulders" {
		return "shoulder"
	}
	return normalized
}

func partitionExercisesByType(exercises []models.Exercise) ([]models.Exercise, []models.Exercise) {
	compoundExercises := make([]models.Exercise, 0, len(exercises))
	isolationExercises := make([]models.Exercise, 0, len(exercises))

	for _, exercise := range exercises {
		if isCompoundExercise(exercise) {
			compoundExercises = append(compoundExercises, exercise)
		} else if isIsolationExercise(exercise) {
			isolationExercises = append(isolationExercises, exercise)
		}
	}

	return compoundExercises, isolationExercises
}

func selectExercisesForDay(compounds, isolations []models.Exercise, count, dayIndex int, used map[string]bool) []models.Exercise {
	selected := make([]models.Exercise, 0, count)
	compoundStart := dayIndex * count
	isolationStart := dayIndex * count

	for slot := 0; slot < count; slot++ {
		wantCompound := slot%2 == 0
		var exercise models.Exercise
		var found bool

		if wantCompound {
			exercise, found = nextAvailableExercise(compounds, compoundStart, used, nil)
			compoundStart++
			if !found {
				exercise, found = nextAvailableExercise(isolations, isolationStart, used, nil)
				isolationStart++
			}
		} else {
			var previousCompound *models.Exercise
			if len(selected) > 0 && isCompoundExercise(selected[len(selected)-1]) {
				previousCompound = &selected[len(selected)-1]
			}

			exercise, found = nextAvailableExercise(isolations, isolationStart, used, previousCompound)
			isolationStart++
			if !found {
				exercise, found = nextAvailableExercise(compounds, compoundStart, used, previousCompound)
				compoundStart++
			}
		}

		if !found {
			break
		}

		selected = append(selected, exercise)
		used[exercise.ID] = true
	}

	return selected
}

func nextAvailableExercise(exercises []models.Exercise, start int, used map[string]bool, previousCompound *models.Exercise) (models.Exercise, bool) {
	if len(exercises) == 0 {
		return models.Exercise{}, false
	}

	for _, requireDifferentPrimaryMuscles := range []bool{true, false} {
		for offset := 0; offset < len(exercises); offset++ {
			exercise := exercises[(start+offset)%len(exercises)]
			if used != nil && used[exercise.ID] {
				continue
			}
			if requireDifferentPrimaryMuscles && previousCompound != nil && sharesPrimaryMuscle(exercise, *previousCompound) {
				continue
			}
			return exercise, true
		}
		if previousCompound == nil {
			break
		}
	}

	return models.Exercise{}, false
}

func isCompoundExercise(exercise models.Exercise) bool {
	return strings.EqualFold(strings.TrimSpace(exercise.Type), "compound")
}

func isIsolationExercise(exercise models.Exercise) bool {
	return strings.EqualFold(strings.TrimSpace(exercise.Type), "isolation")
}

func sharesPrimaryMuscle(first, second models.Exercise) bool {
	firstMuscles := make(map[string]bool)
	for _, muscle := range strings.Split(first.PrimaryMuscles, ",") {
		muscle = strings.ToLower(strings.TrimSpace(muscle))
		if muscle != "" {
			firstMuscles[muscle] = true
		}
	}
	for _, muscle := range strings.Split(second.PrimaryMuscles, ",") {
		muscle = strings.ToLower(strings.TrimSpace(muscle))
		if firstMuscles[muscle] {
			return true
		}
	}
	return false
}

func scoreExercise(ex models.Exercise, goal string) int {
	score := 0

	// Bonus for compound exercises
	if isCompoundExercise(ex) {
		score += 3
	}

	// Goal-specific bonuses for key lift types
	if goal == "powerlifting" {
		exerciseLower := strings.ToLower(ex.Exercise)
		if strings.Contains(exerciseLower, "squat") || strings.Contains(exerciseLower, "bench") || strings.Contains(exerciseLower, "deadlift") {
			score += 5
		}
	}
	if goal == "strength" {
		exerciseLower := strings.ToLower(ex.Exercise)
		if strings.Contains(exerciseLower, "press") || strings.Contains(exerciseLower, "squat") || strings.Contains(exerciseLower, "deadlift") || strings.Contains(exerciseLower, "row") {
			score += 4
		}
	}

	// Bonus for focus alignment
	if strings.EqualFold(ex.Focus, "Leg") {
		score += 2
	}
	if strings.EqualFold(ex.Focus, "Upper") {
		score += 2
	}

	return score
}

func filterExercisesForRequest(exercises []models.Exercise, req RecommendationRequest) []models.Exercise {
	filtered := make([]models.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
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

func (h *RoutineHandler) Recommend(c *gin.Context) {
	var req RecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.ExerciseStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Exercise store is not configured"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	exercises, err := h.ExerciseStore.List(ctx, map[string]string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	exercises = filterExercisesForRequest(exercises, req)
	routine, err := BuildRoutineRecommendation(exercises, req.Goal, req.Days, req.Focus)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, routine)
}
