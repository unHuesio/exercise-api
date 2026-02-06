package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Exercise struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Exercise         string             `bson:"Exercise" json:"Exercise" binding:"required"`
	PrimaryMuscles   string             `bson:"Primary Muscles" json:"PrimaryMuscles" binding:"required"`
	SecondaryMuscles string             `bson:"Secondary Muscles" json:"SecondaryMuscles"`
	Type             string             `bson:"Type" json:"Type" binding:"required"`
	Focus            string             `bson:"Focus" json:"Focus" binding:"required"`
}
