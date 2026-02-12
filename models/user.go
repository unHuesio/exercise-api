package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email" json:"email" binding:"required"`
	Password string             `bson:"password" json:"password" binding:"required"`
}

type Application struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name" binding:"required"`
	Email     string             `bson:"email" json:"email" binding:"required"`
	Status    string             `bson:"status" json:"status"`
	ApiKey    string             `bson:"api_key" json:"api_key" binding:"required"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
}
