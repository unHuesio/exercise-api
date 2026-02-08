package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ApiKey struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	APIKey    string             `bson:"api_key" json:"api_key"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	IsValid   bool               `bson:"is_valid" json:"is_valid"`
	Account   string             `bson:"account" json:"account" binding:"required"`
}
