package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Set struct {
	Reps   int     `json:"reps" bson:"reps"`
	Weight float64 `json:"weight" bson:"weight"`
	Rest   int     `json:"rest" bson:"rest"`
}

type RoutineExercise struct {
	ExerciseID primitive.ObjectID `json:"exercise_id" bson:"exercise_id"`
	Sets       []Set              `json:"sets" bson:"sets"`
	Order      int                `json:"order" bson:"order"`
}

type Routine struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `json:"name" bson:"name" binding:"required"`
	Description string             `json:"description" bson:"description"`
	Exercises   []RoutineExercise  `json:"exercises" bson:"exercises"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}
