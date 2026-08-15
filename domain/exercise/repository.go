package exercise

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Exercise struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Exercise         string             `bson:"Exercise" json:"Exercise" binding:"required"`
	PrimaryMuscles   string             `bson:"Primary Muscles" json:"PrimaryMuscles" binding:"required"`
	SecondaryMuscles string             `bson:"Secondary Muscles" json:"SecondaryMuscles"`
	Type             string             `bson:"Type" json:"Type" binding:"required"`
	Focus            string             `bson:"Focus" json:"Focus" binding:"required"`
}

type Filter struct {
	Focus  string
	Type   string
	Muscle string
}

func (e *Exercise) Validate() error {
	if e == nil {
		return errors.New("exercise cannot be nil")
	}
	if e.Exercise == "" {
		return errors.New("exercise is required")
	}
	if e.PrimaryMuscles == "" {
		return errors.New("primary muscles are required")
	}
	if e.Type == "" {
		return errors.New("type is required")
	}
	if e.Focus == "" {
		return errors.New("focus is required")
	}
	return nil
}

func (e *Exercise) StampCreatedAt(_ time.Time) {}

type Repository interface {
	FindAll(ctx context.Context, filter Filter) ([]Exercise, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*Exercise, error)
	Create(ctx context.Context, exercise *Exercise) (primitive.ObjectID, error)
	Update(ctx context.Context, exercise *Exercise) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
