package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestSetOmitsZeroWeightFromBSON(t *testing.T) {
	data, err := bson.Marshal(Set{Reps: 6, Rest: 120})
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}

	var stored bson.M
	if err := bson.Unmarshal(data, &stored); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	if _, exists := stored["weight"]; exists {
		t.Fatalf("stored set contains weight: %#v", stored)
	}
}

func TestSetStoresSpecifiedWeightInBSON(t *testing.T) {
	data, err := bson.Marshal(Set{Reps: 6, Weight: 42.5, Rest: 120})
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}

	var stored bson.M
	if err := bson.Unmarshal(data, &stored); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	if got := stored["weight"]; got != 42.5 {
		t.Fatalf("stored weight = %#v, want 42.5", got)
	}
}
