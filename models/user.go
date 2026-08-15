package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email           string             `bson:"email" json:"email" binding:"required"`
	Provider        string             `bson:"provider" json:"provider" binding:"required"`
	ProviderSubject string             `bson:"provider_subject" json:"provider_subject" binding:"required"`
}
