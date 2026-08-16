package models

type Exercise struct {
	ID               string `bson:"_id,omitempty" json:"id,omitempty"`
	Exercise         string `bson:"Exercise" json:"Exercise" binding:"required"`
	PrimaryMuscles   string `bson:"PrimaryMuscles" json:"PrimaryMuscles" binding:"required"`
	SecondaryMuscles string `bson:"SecondaryMuscles" json:"SecondaryMuscles"`
	Type             string `bson:"Type" json:"Type" binding:"required"`
	Focus            string `bson:"Focus" json:"Focus"`
}
