package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestExerciseDecodesMongoFields(t *testing.T) {
	id := primitive.NewObjectID()
	raw := bson.M{
		"_id":              id,
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
	if exercise.ID != id.Hex() {
		t.Fatalf("expected ID %q, got %q", id.Hex(), exercise.ID)
	}
}

func TestExerciseDecodesLegacyMongoFieldNames(t *testing.T) {
	raw := bson.M{
		"Exercise":          "Barbell Hack Squat",
		"Primary Muscles":   "Quads",
		"Secondary Muscles": "Glutes, Adductors",
		"Type":              "Compound",
		"Focus":             "Leg",
	}

	b, err := bson.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal legacy raw exercise: %v", err)
	}

	var exercise Exercise
	if err := bson.Unmarshal(b, &exercise); err != nil {
		t.Fatalf("unmarshal legacy exercise: %v", err)
	}

	if exercise.PrimaryMuscles != "Quads" {
		t.Fatalf("expected legacy primary muscles Quads, got %q", exercise.PrimaryMuscles)
	}
	if exercise.SecondaryMuscles != "Glutes, Adductors" {
		t.Fatalf("expected legacy secondary muscles Glutes, Adductors, got %q", exercise.SecondaryMuscles)
	}
}
