package models
package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestExerciseDecodesMongoFields(t *testing.T) {
	raw := bson.M{
		"Exercise":         "Barbell Hack Squat",
		"PrimaryMuscles":   "Quads",
		"SecondaryMuscles": "Glutes, Adductors",
		"Type":             "Compound",
		"Focus":            "Leg",
	}

	b, err := bson.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal raw exercise: %v", err)
	}

	var exercise Exercise
	if err := bson.Unmarshal(b, &exercise); err != nil {
		t.Fatalf("unmarshal exercise: %v", err)
	}

	if exercise.PrimaryMuscles != "Quads" {
		t.Fatalf("expected primary muscles Quads, got %q", exercise.PrimaryMuscles)
	}
	if exercise.SecondaryMuscles != "Glutes, Adductors" {
		t.Fatalf("expected secondary muscles Glutes, Adductors, got %q", exercise.SecondaryMuscles)
	}
	if exercise.Exercise != "Barbell Hack Squat" {
		t.Fatalf("expected exercise name Barbell Hack Squat, got %q", exercise.Exercise)
	}
	if exercise.Type != "Compound" {
		t.Fatalf("expected type Compound, got %q", exercise.Type)
	}
	if exercise.Focus != "Leg" {
		t.Fatalf("expected focus Leg, got %q", exercise.Focus)
	}
}
