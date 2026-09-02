package handlers

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRoutineFromRecommendation(t *testing.T) {
	userID := primitive.NewObjectID()
	exerciseID := primitive.NewObjectID()
	createdAt := time.Date(2026, time.September, 2, 12, 0, 0, 0, time.UTC)
	recommendation := &RecommendationRoutine{
		Name:        "Recommended Strength Routine",
		Description: "A saved recommendation",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		Exercises: []RecommendationExercise{{
			ID:    exerciseID.Hex(),
			Order: 2,
			Sets:  []RecommendationSet{{Reps: 6, Rest: 120}},
		}},
	}

	routine, err := routineFromRecommendation(recommendation, userID)
	if err != nil {
		t.Fatalf("routineFromRecommendation() error = %v", err)
	}
	if routine.UserID != userID || routine.CreatedAt != createdAt || routine.UpdatedAt != createdAt {
		t.Fatalf("routine metadata = %+v, want owner and recommendation timestamps", routine)
	}
	if len(routine.Exercises) != 1 || routine.Exercises[0].ExerciseID != exerciseID {
		t.Fatalf("routine exercises = %+v, want source exercise ID", routine.Exercises)
	}
	if got := routine.Exercises[0].Sets[0]; got.Reps != 6 || got.Rest != 120 || got.Weight != 0 {
		t.Fatalf("routine set = %+v, want generated reps/rest and zero weight", got)
	}
}

func TestRoutineFromRecommendationRejectsInvalidExerciseID(t *testing.T) {
	_, err := routineFromRecommendation(&RecommendationRoutine{
		Exercises: []RecommendationExercise{{ID: "legacy-id"}},
	}, primitive.NewObjectID())
	if err == nil {
		t.Fatal("expected invalid recommendation exercise ID error")
	}
}

func TestRoutineOwnerFilters(t *testing.T) {
	userID := primitive.NewObjectID()
	routineID := primitive.NewObjectID()

	gotOwnerFilter := routineOwnerFilter(userID)
	wantOwnerFilter := bson.M{"user_id": userID}
	if !equalFilters(gotOwnerFilter, wantOwnerFilter) {
		t.Fatalf("routineOwnerFilter() = %#v, want %#v", gotOwnerFilter, wantOwnerFilter)
	}
	gotIDOwnerFilter := routineIDOwnerFilter(routineID, userID)
	wantIDOwnerFilter := bson.M{"_id": routineID, "user_id": userID}
	if !equalFilters(gotIDOwnerFilter, wantIDOwnerFilter) {
		t.Fatalf("routineIDOwnerFilter() = %#v, want %#v", gotIDOwnerFilter, wantIDOwnerFilter)
	}
}

func equalFilters(first, second bson.M) bool {
	return first["_id"] == second["_id"] && first["user_id"] == second["user_id"]
}
