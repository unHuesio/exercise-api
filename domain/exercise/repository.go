package exercise

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Filter struct {
	Focus  string
	Type   string
	Muscle string
}

type Repository interface {
	FindAll(ctx context.Context, filter Filter) ([]Exercise, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*Exercise, error)
	Create(ctx context.Context, exercise *Exercise) (primitive.ObjectID, error)
	Update(ctx context.Context, exercise *Exercise) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
