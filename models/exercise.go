package models

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Exercise struct {
	ID               string `bson:"_id,omitempty" json:"id,omitempty"`
	Exercise         string `bson:"Exercise" json:"Exercise" binding:"required"`
	PrimaryMuscles   string `bson:"PrimaryMuscles" json:"PrimaryMuscles" binding:"required"`
	SecondaryMuscles string `bson:"SecondaryMuscles" json:"SecondaryMuscles"`
	Type             string `bson:"Type" json:"Type" binding:"required"`
	Focus            string `bson:"Focus" json:"Focus"`
}

func (e *Exercise) UnmarshalBSON(data []byte) error {
	type exerciseAlias struct {
		ID               any    `bson:"_id,omitempty"`
		Exercise         string `bson:"Exercise"`
		PrimaryMuscles   string `bson:"PrimaryMuscles"`
		SecondaryMuscles string `bson:"SecondaryMuscles"`
		Type             string `bson:"Type"`
		Focus            string `bson:"Focus"`
		LegacyPrimary    string `bson:"Primary Muscles"`
		LegacySecondary  string `bson:"Secondary Muscles"`
	}

	var raw exerciseAlias
	if err := bson.Unmarshal(data, &raw); err != nil {
		return err
	}

	*e = Exercise{
		ID:               exerciseIDString(raw.ID),
		Exercise:         raw.Exercise,
		PrimaryMuscles:   firstNonEmpty(raw.PrimaryMuscles, raw.LegacyPrimary),
		SecondaryMuscles: firstNonEmpty(raw.SecondaryMuscles, raw.LegacySecondary),
		Type:             raw.Type,
		Focus:            raw.Focus,
	}
	return nil
}

func exerciseIDString(id any) string {
	switch value := id.(type) {
	case primitive.ObjectID:
		return value.Hex()
	case string:
		return value
	default:
		return ""
	}
}

func (e *Exercise) UnmarshalJSON(data []byte) error {
	type exerciseAlias struct {
		ID               string `json:"id,omitempty"`
		LegacyID         string `json:"_id,omitempty"`
		Exercise         string `json:"Exercise"`
		PrimaryMuscles   string `json:"PrimaryMuscles"`
		SecondaryMuscles string `json:"SecondaryMuscles"`
		Type             string `json:"Type"`
		Focus            string `json:"Focus"`
	}

	var raw exerciseAlias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*e = Exercise{
		ID:               firstNonEmpty(raw.ID, raw.LegacyID),
		Exercise:         raw.Exercise,
		PrimaryMuscles:   raw.PrimaryMuscles,
		SecondaryMuscles: raw.SecondaryMuscles,
		Type:             raw.Type,
		Focus:            raw.Focus,
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
