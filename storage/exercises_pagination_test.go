package storage

import (
	"testing"

	"gym-api/m/models"
)

func TestPageExercisesBounds(t *testing.T) {
	items := []models.Exercise{
		{ID: "1"}, {ID: "2"}, {ID: "3"},
	}

	page := pageExercises(items, 2, 2)
	if len(page.Items) != 1 || page.Items[0].ID != "3" {
		t.Fatalf("pageExercises() items = %#v, want final item", page.Items)
	}
	if page.Total != 3 {
		t.Fatalf("pageExercises() total = %d, want 3", page.Total)
	}

	empty := pageExercises(items, 4, 2)
	if len(empty.Items) != 0 {
		t.Fatalf("pageExercises() out-of-range items = %#v, want empty", empty.Items)
	}
}
