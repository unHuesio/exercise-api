package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gym-api/m/models"
	"gym-api/m/storage"

	"github.com/gin-gonic/gin"
)

func TestBuildRoutineRecommendationStrength(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "001", Exercise: "Barbell Squat", PrimaryMuscles: "Quads, Glutes", SecondaryMuscles: "Hamstrings", Type: "Compound", Focus: "Leg"},
		{ID: "002", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", SecondaryMuscles: "Shoulders", Type: "Compound", Focus: "Upper"},
		{ID: "003", Exercise: "Deadlift", PrimaryMuscles: "Hamstrings, Glutes", SecondaryMuscles: "Lower Back", Type: "Compound", Focus: "Leg"},
		{ID: "004", Exercise: "Pull-Up", PrimaryMuscles: "Lats", SecondaryMuscles: "Biceps", Type: "Compound", Focus: "Upper"},
		{ID: "005", Exercise: "Cable Row", PrimaryMuscles: "Back", SecondaryMuscles: "Biceps", Type: "Isolation", Focus: "Upper"},
		{ID: "006", Exercise: "Dumbbell Curl", PrimaryMuscles: "Biceps", SecondaryMuscles: "Forearms", Type: "Isolation", Focus: "Upper"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "strength", "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}
	if routine == nil {
		t.Fatal("expected routine, got nil")
	}
	// A one-day recommendation contains up to five exercises.
	expectedExercises := 5
	if len(routine.Exercises) != expectedExercises {
		t.Fatalf("len(routine.Exercises) = %d, want %d", len(routine.Exercises), expectedExercises)
	}
	// Verify each exercise has sets
	if len(routine.Exercises[0].Sets) == 0 {
		t.Fatal("first exercise has no sets")
	}
	if routine.Exercises[0].ID == "" || routine.Exercises[0].Name == "" {
		t.Fatalf("recommendation must include an exercise ID and name, got %+v", routine.Exercises[0])
	}
	if routine.Exercises[0].Day != 1 {
		t.Fatalf("first exercise day = %d, want 1", routine.Exercises[0].Day)
	}
	if routine.Days != 1 {
		t.Fatalf("routine days = %d, want 1", routine.Days)
	}
	response, err := json.Marshal(routine)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(response), `"id":"001"`) || !strings.Contains(string(response), `"name":"Barbell Squat"`) {
		t.Fatalf("response must expose usable exercise details, got %s", response)
	}
	if strings.Contains(string(response), `"id":""`) {
		t.Fatalf("response must omit the non-persisted routine ID, got %s", response)
	}
	if strings.Contains(string(response), `"weight"`) {
		t.Fatalf("response must not expose set weight, got %s", response)
	}
	// Strength goal with compound should have 4 sets
	if routine.Exercises[0].Sets[0].Reps != 6 {
		t.Fatalf("expected 6 reps for strength goal, got %d", routine.Exercises[0].Sets[0].Reps)
	}
}

func TestBuildRoutineRecommendationPowerlifting(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "001", Exercise: "Back Squat", PrimaryMuscles: "Quads, Glutes", SecondaryMuscles: "Hamstrings", Type: "Compound", Focus: "Leg"},
		{ID: "002", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", SecondaryMuscles: "Shoulders", Type: "Compound", Focus: "Upper"},
		{ID: "003", Exercise: "Deadlift", PrimaryMuscles: "Hamstrings, Glutes", SecondaryMuscles: "Lower Back", Type: "Compound", Focus: "Leg"},
		{ID: "004", Exercise: "Overhead Press", PrimaryMuscles: "Shoulders", SecondaryMuscles: "Triceps", Type: "Compound", Focus: "Upper"},
		{ID: "005", Exercise: "Romanian Deadlift", PrimaryMuscles: "Hamstrings", SecondaryMuscles: "Glutes", Type: "Compound", Focus: "Leg"},
		{ID: "006", Exercise: "Close-Grip Bench", PrimaryMuscles: "Triceps, Chest", SecondaryMuscles: "Shoulders", Type: "Compound", Focus: "Upper"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "powerlifting", "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}
	if routine == nil {
		t.Fatal("expected routine, got nil")
	}
	expectedExercises := 5
	if len(routine.Exercises) != expectedExercises {
		t.Fatalf("len(routine.Exercises) = %d, want %d", len(routine.Exercises), expectedExercises)
	}
	// Powerlifting goal should have 4 reps
	if routine.Exercises[0].Sets[0].Reps != 4 {
		t.Fatalf("expected 4 reps for powerlifting goal, got %d", routine.Exercises[0].Sets[0].Reps)
	}
}

func TestBuildRoutineRecommendationDoesNotRepeatExercises(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "001", Exercise: "Barbell Squat", PrimaryMuscles: "Quads, Glutes", SecondaryMuscles: "Hamstrings", Type: "Compound", Focus: "Leg"},
		{ID: "002", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", SecondaryMuscles: "Shoulders", Type: "Compound", Focus: "Upper"},
		{ID: "003", Exercise: "Deadlift", PrimaryMuscles: "Hamstrings, Glutes", SecondaryMuscles: "Lower Back", Type: "Compound", Focus: "Leg"},
		{ID: "004", Exercise: "Pull-Up", PrimaryMuscles: "Lats", SecondaryMuscles: "Biceps", Type: "Compound", Focus: "Upper"},
		{ID: "005", Exercise: "Cable Row", PrimaryMuscles: "Back", SecondaryMuscles: "Biceps", Type: "Isolation", Focus: "Upper"},
		{ID: "006", Exercise: "Dumbbell Curl", PrimaryMuscles: "Biceps", SecondaryMuscles: "Forearms", Type: "Isolation", Focus: "Upper"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "strength", "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}

	if len(routine.Exercises) != 5 {
		t.Fatalf("len(routine.Exercises) = %d, want 5", len(routine.Exercises))
	}
	seen := make(map[string]bool)
	for _, exercise := range routine.Exercises {
		if exercise.ID == "" {
			t.Fatal("recommendation exercise must return its source ID")
		}
		if seen[exercise.ID] {
			t.Fatalf("exercise %q appears more than once", exercise.ID)
		}
		seen[exercise.ID] = true
	}
}

func TestBuildRoutineRecommendationAlternatesCompoundAndIsolation(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "compound-1", Exercise: "Barbell Squat", PrimaryMuscles: "Quads, Glutes", Type: "Compound", Focus: "Leg"},
		{ID: "compound-2", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", Type: "Compound", Focus: "Chest"},
		{ID: "compound-3", Exercise: "Deadlift", PrimaryMuscles: "Hamstrings, Glutes", Type: "Compound", Focus: "Leg"},
		{ID: "isolation-1", Exercise: "Leg Extension", PrimaryMuscles: "Quads", Type: "Isolation", Focus: "Leg"},
		{ID: "isolation-2", Exercise: "Bicep Curl", PrimaryMuscles: "Biceps", Type: "Isolation", Focus: "Bicep"},
		{ID: "isolation-3", Exercise: "Tricep Extension", PrimaryMuscles: "Triceps", Type: "Isolation", Focus: "Tricep"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "strength", "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}

	expectedTypes := []string{"Compound", "Isolation", "Compound", "Isolation", "Compound"}
	if len(routine.Exercises) != len(expectedTypes) {
		t.Fatalf("len(routine.Exercises) = %d, want %d", len(routine.Exercises), len(expectedTypes))
	}
	for index, exercise := range routine.Exercises {
		if exercise.Type != expectedTypes[index] {
			t.Fatalf("exercise %d type = %q, want %q", index+1, exercise.Type, expectedTypes[index])
		}
	}
}

func TestBuildRoutineRecommendationAvoidsCompoundPrimaryMuscles(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "compound-1", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", Type: "Compound", Focus: "Chest"},
		{ID: "compound-2", Exercise: "Barbell Lunge", PrimaryMuscles: "Quads, Glutes", Type: "Compound", Focus: "Leg"},
		{ID: "compound-3", Exercise: "Pull-Up", PrimaryMuscles: "Lats", Type: "Compound", Focus: "Back"},
		{ID: "isolation-overlap", Exercise: "Chest Fly", PrimaryMuscles: "Chest", Type: "Isolation", Focus: "Chest"},
		{ID: "isolation-rest", Exercise: "Bicep Curl", PrimaryMuscles: "Biceps", Type: "Isolation", Focus: "Bicep"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "strength", "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}

	if routine.Exercises[1].Name != "Bicep Curl" {
		t.Fatalf("first isolation = %q, want an exercise that does not overlap Bench Press primary muscles", routine.Exercises[1].Name)
	}
}

func TestBuildRoutineRecommendationReturnsShorterResultWhenOnlyOneTypeIsAvailable(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "compound-1", Exercise: "Barbell Squat", PrimaryMuscles: "Quads, Glutes", Type: "Compound", Focus: "Leg"},
		{ID: "compound-2", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", Type: "Compound", Focus: "Chest"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "strength", "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}
	if len(routine.Exercises) != len(exercises) {
		t.Fatalf("len(routine.Exercises) = %d, want %d", len(routine.Exercises), len(exercises))
	}
	if routine.Description != "1-day strength routine with 2 unique exercises" {
		t.Fatalf("description = %q, want shortened-routine description", routine.Description)
	}
	for _, exercise := range routine.Exercises {
		if exercise.Type != "Compound" {
			t.Fatalf("fallback exercise type = %q, want Compound", exercise.Type)
		}
	}
}

func TestBuildRoutineRecommendationUsesMappedIsolationFocus(t *testing.T) {
	testCases := []struct {
		name           string
		focus          string
		compound       models.Exercise
		accessoryFocus string
	}{
		{
			name:           "chest uses tricep isolation",
			focus:          "Chest",
			compound:       models.Exercise{ID: "chest-compound", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", Type: "Compound", Focus: "Chest"},
			accessoryFocus: "Tricep",
		},
		{
			name:           "back uses bicep isolation",
			focus:          "Back",
			compound:       models.Exercise{ID: "back-compound", Exercise: "Barbell Row", PrimaryMuscles: "Back", Type: "Compound", Focus: "Back"},
			accessoryFocus: "Bicep",
		},
		{
			name:           "leg uses shoulder isolation",
			focus:          "Leg",
			compound:       models.Exercise{ID: "leg-compound", Exercise: "Barbell Squat", PrimaryMuscles: "Quads, Glutes", Type: "Compound", Focus: "Leg"},
			accessoryFocus: "Shoulder",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			exercises := []models.Exercise{
				testCase.compound,
				{ID: "mapped-isolation", Exercise: testCase.accessoryFocus + " Isolation", PrimaryMuscles: testCase.accessoryFocus, Type: "Isolation", Focus: testCase.accessoryFocus},
				{ID: "same-focus-isolation", Exercise: testCase.focus + " Isolation", PrimaryMuscles: testCase.focus, Type: "Isolation", Focus: testCase.focus},
			}

			routine, err := BuildRoutineRecommendation(exercises, "strength", testCase.focus)
			if err != nil {
				t.Fatalf("BuildRoutineRecommendation() error = %v", err)
			}
			if routine.Exercises[1].Focus != testCase.accessoryFocus {
				t.Fatalf("first isolation focus = %q, want %q", routine.Exercises[1].Focus, testCase.accessoryFocus)
			}
		})
	}
}

func TestBuildRoutineRecommendationTreatsShoulderFocusAsLegAccessory(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "leg-compound", Exercise: "Barbell Squat", PrimaryMuscles: "Quads, Glutes", Type: "Compound", Focus: "Leg"},
		{ID: "shoulders-isolation", Exercise: "Lateral Raise", PrimaryMuscles: "Shoulders", Type: "Isolation", Focus: "Shoulders"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "strength", "Leg")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}
	if routine.Exercises[1].Focus != "Shoulders" {
		t.Fatalf("first isolation focus = %q, want Shoulders", routine.Exercises[1].Focus)
	}
}

func TestBuildRoutineRecommendationFallsBackWhenMappedAccessoryIsUnavailable(t *testing.T) {
	exercises := []models.Exercise{
		{ID: "chest-compound", Exercise: "Bench Press", PrimaryMuscles: "Chest, Triceps", Type: "Compound", Focus: "Chest"},
		{ID: "bicep-isolation", Exercise: "Bicep Curl", PrimaryMuscles: "Biceps", Type: "Isolation", Focus: "Bicep"},
	}

	routine, err := BuildRoutineRecommendation(exercises, "strength", "Chest")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}
	if len(routine.Exercises) != 1 {
		t.Fatalf("len(routine.Exercises) = %d, want 1", len(routine.Exercises))
	}
	for _, exercise := range routine.Exercises {
		if exercise.Type != "Compound" || exercise.Focus != "Chest" {
			t.Fatalf("fallback exercise = %+v, want a chest compound", exercise)
		}
	}
}

func TestRecommendUsesOneDayRequestAndOmitsWeight(t *testing.T) {
	store := &alternativeExerciseStore{exercises: []models.Exercise{
		{ID: "001", Exercise: "Barbell Squat", PrimaryMuscles: "Quads, Glutes", Type: "Compound", Focus: "Leg"},
		{ID: "002", Exercise: "Leg Extension", PrimaryMuscles: "Quads", Type: "Isolation", Focus: "Leg"},
	}}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/recommendations/routine",
		strings.NewReader(`{"goal":"strength"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	(&RoutineHandler{ExerciseStore: store}).Recommend(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Days      int `json:"days"`
		Exercises []struct {
			Day  int `json:"day"`
			Sets []struct {
				Reps int `json:"reps"`
				Rest int `json:"rest"`
			} `json:"sets"`
		} `json:"exercises"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode recommendation response: %v", err)
	}
	if response.Days != 1 {
		t.Fatalf("response days = %d, want 1", response.Days)
	}
	if len(response.Exercises) == 0 || response.Exercises[0].Day != 1 {
		t.Fatalf("response exercises = %+v, want exercises assigned to day 1", response.Exercises)
	}
	if strings.Contains(recorder.Body.String(), `"weight"`) {
		t.Fatalf("response must not expose set weight, got %s", recorder.Body.String())
	}
}

func TestRecommendAlternativeReturnsMatchingExercise(t *testing.T) {
	source := models.Exercise{ID: "source", Exercise: "Barbell Row", PrimaryMuscles: "Back, Biceps", Type: "Compound", Focus: "Back"}
	store := &alternativeExerciseStore{exercises: []models.Exercise{
		source,
		{ID: "alternative-z", Exercise: "Cable Row", PrimaryMuscles: "Biceps, Back", Type: "Compound", Focus: "Back"},
		{ID: "alternative-a", Exercise: "Chest Press", PrimaryMuscles: "Back, Biceps", Type: "Compound", Focus: "Back"},
		{ID: "different-focus", Exercise: "Lat Pulldown", PrimaryMuscles: "Back, Biceps", Type: "Compound", Focus: "Chest"},
	}}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/recommendations/alternative/source", nil)
	context.Params = gin.Params{{Key: "id", Value: source.ID}}

	(&RoutineHandler{ExerciseStore: store}).RecommendAlternative(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var alternative models.Exercise
	if err := json.Unmarshal(recorder.Body.Bytes(), &alternative); err != nil {
		t.Fatalf("decode alternative response: %v", err)
	}
	if alternative.ID != "alternative-a" {
		t.Fatalf("alternative ID = %q, want deterministic matching ID %q", alternative.ID, "alternative-a")
	}
}

func TestRecommendAlternativeReturnsNotFoundWithoutMatch(t *testing.T) {
	source := models.Exercise{ID: "source", Exercise: "Barbell Row", PrimaryMuscles: "Back", Type: "Compound", Focus: "Back"}
	store := &alternativeExerciseStore{exercises: []models.Exercise{source}}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/recommendations/alternative/source", nil)
	context.Params = gin.Params{{Key: "id", Value: source.ID}}

	(&RoutineHandler{ExerciseStore: store}).RecommendAlternative(context)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestRecommendAlternativeReturnsNotFoundForMissingSource(t *testing.T) {
	store := &alternativeExerciseStore{}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/recommendations/alternative/missing", nil)
	context.Params = gin.Params{{Key: "id", Value: "missing"}}

	(&RoutineHandler{ExerciseStore: store}).RecommendAlternative(context)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

type alternativeExerciseStore struct {
	exercises []models.Exercise
}

func (s *alternativeExerciseStore) List(_ context.Context, _ map[string]string) ([]models.Exercise, error) {
	return s.exercises, nil
}

func (s *alternativeExerciseStore) GetByID(_ context.Context, id string) (models.Exercise, error) {
	for _, exercise := range s.exercises {
		if exercise.ID == id {
			return exercise, nil
		}
	}
	return models.Exercise{}, storage.ErrNotFound
}

func (s *alternativeExerciseStore) Create(_ context.Context, _ models.Exercise) (string, error) {
	return "", nil
}

func (s *alternativeExerciseStore) Update(_ context.Context, _ string, _ models.Exercise) error {
	return nil
}

func (s *alternativeExerciseStore) Delete(_ context.Context, _ string) error {
	return nil
}

func (s *alternativeExerciseStore) Health(_ context.Context) error {
	return nil
}

func (s *alternativeExerciseStore) BackendName() string {
	return "test"
}

func TestBuildRoutineRecommendationRejectsInvalidGoal(t *testing.T) {
	_, err := BuildRoutineRecommendation([]models.Exercise{{ID: "001", Exercise: "Squat", PrimaryMuscles: "Quads", Type: "Compound", Focus: "Leg"}}, "cardio", "")
	if err == nil {
		t.Fatal("expected error for invalid goal")
	}
}
