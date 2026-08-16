package models

import "go.mongodb.org/mongo-driver/bson"

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
		ID               string `bson:"_id,omitempty"`
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
		ID:               raw.ID,
		Exercise:         raw.Exercise,
		PrimaryMuscles:   firstNonEmpty(raw.PrimaryMuscles, raw.LegacyPrimary),
		SecondaryMuscles: firstNonEmpty(raw.SecondaryMuscles, raw.LegacySecondary),
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
