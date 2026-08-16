package handlers

import (
	"testing"

	"gym-api/m/models"
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

	plan, err := BuildRoutineRecommendation(exercises, "strength", 3, "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}
	if len(plan.Days) != 3 {
		t.Fatalf("len(plan.Days) = %d, want 3", len(plan.Days))
	}
	if len(plan.Days[0].Exercises) == 0 {
		t.Fatal("first day has no exercises")
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

	plan, err := BuildRoutineRecommendation(exercises, "powerlifting", 4, "")
	if err != nil {
		t.Fatalf("BuildRoutineRecommendation() error = %v", err)
	}
	if len(plan.Days) != 4 {
		t.Fatalf("len(plan.Days) = %d, want 4", len(plan.Days))
	}
}

func TestBuildRoutineRecommendationRejectsInvalidGoal(t *testing.T) {
	_, err := BuildRoutineRecommendation([]models.Exercise{{ID: "001", Exercise: "Squat", PrimaryMuscles: "Quads", Type: "Compound", Focus: "Leg"}}, "cardio", 3, "")
	if err == nil {
		t.Fatal("expected error for invalid goal")
	}
}

func TestLoadSampleExercisesFallbackIncludesLegFocus(t *testing.T) {
	exercises, err := loadSampleExercises("")
	if err != nil {
		t.Fatalf("loadSampleExercises() error = %v", err)
	}
	if len(exercises) == 0 {
		t.Fatal("expected sample exercises to be loaded")
	}

	filtered := filterExercisesForRequest(exercises, RecommendationRequest{Focus: "Leg", Goal: "strength", Days: 3})
	if len(filtered) == 0 {
		t.Fatal("expected sample exercises to include leg-focused routines")
	}
}
