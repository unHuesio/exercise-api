package handlers

import (
	"context"
	"path/filepath"
	"testing"

	"gym-api/m/models"
	"gym-api/m/storage"
)

func TestExerciseStoreFileCRUD(t *testing.T) {
	file := filepath.Join(t.TempDir(), "excercises.json")

	store, err := storage.NewFileExerciseStore(file)
	if err != nil {
		t.Fatalf("NewFileExerciseStore() error = %v", err)
	}

	if err := store.Health(context.Background()); err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	createdID, err := store.Create(context.Background(), models.Exercise{
		ID:               "001",
		Exercise:         "Test Press",
		PrimaryMuscles:   "Shoulders",
		SecondaryMuscles: "Triceps",
		Type:             "Strength",
		Focus:            "Upper",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if createdID != "001" {
		t.Fatalf("Create() id = %q, want %q", createdID, "001")
	}

	items, err := store.List(context.Background(), map[string]string{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("List() len = %d, want 1", len(items))
	}

	item, err := store.GetByID(context.Background(), "001")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if item.Exercise != "Test Press" {
		t.Fatalf("GetByID() Exercise = %q, want %q", item.Exercise, "Test Press")
	}

	if err := store.Update(context.Background(), "001", models.Exercise{
		ID:               "001",
		Exercise:         "Updated Test Press",
		PrimaryMuscles:   "Shoulders",
		SecondaryMuscles: "Triceps, Chest",
		Type:             "Strength",
		Focus:            "Upper",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := store.GetByID(context.Background(), "001")
	if err != nil {
		t.Fatalf("GetByID() after update error = %v", err)
	}
	if updated.SecondaryMuscles != "Triceps, Chest" {
		t.Fatalf("updated.SecondaryMuscles = %q, want %q", updated.SecondaryMuscles, "Triceps, Chest")
	}

	if err := store.Delete(context.Background(), "001"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := store.GetByID(context.Background(), "001"); err == nil {
		t.Fatal("GetByID() after delete returned a record, want not found")
	}
}
