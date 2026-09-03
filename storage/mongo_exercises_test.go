package storage

import (
	"context"
	"testing"

	"gym-api/m/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNewMongoExerciseStoreUsesCanonicalCollection(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("mongo.Connect() error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Errorf("client.Disconnect() error = %v", err)
		}
	})

	store, err := NewMongoExerciseStore(client, "gym-app")
	if err != nil {
		t.Fatalf("NewMongoExerciseStore() error = %v", err)
	}

	if store.collection.Database().Name() != "gym-app" {
		t.Fatalf("database = %q, want %q", store.collection.Database().Name(), "gym-app")
	}
	if store.collection.Name() != "excercises" {
		t.Fatalf("collection = %q, want %q", store.collection.Name(), "excercises")
	}
}

func TestExerciseDocumentUsesObjectID(t *testing.T) {
	id := primitive.NewObjectID()
	document := exerciseDocument(models.Exercise{Exercise: "Squat"}, id)

	storedID, ok := document["_id"].(primitive.ObjectID)
	if !ok {
		t.Fatalf("_id type = %T, want primitive.ObjectID", document["_id"])
	}
	if storedID != id {
		t.Fatalf("_id = %s, want %s", storedID.Hex(), id.Hex())
	}
}

func TestExerciseIDFilterMatchesObjectIDAndLegacyStringID(t *testing.T) {
	id := primitive.NewObjectID()
	filter := exerciseIDFilter(id.Hex())

	idFilter, ok := filter["_id"].(bson.M)
	if !ok {
		t.Fatalf("_id filter type = %T, want bson.M", filter["_id"])
	}
	values, ok := idFilter["$in"].(bson.A)
	if !ok {
		t.Fatalf("$in type = %T, want bson.A", idFilter["$in"])
	}
	if len(values) != 2 || values[0] != id || values[1] != id.Hex() {
		t.Fatalf("$in values = %#v, want ObjectID and hex string", values)
	}
}
