package mongo

import (
	"context"
	"errors"

	"gym-api/m/domain/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	Collection *mongo.Collection
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (primitive.ObjectID, error) {
	if u == nil {
		return primitive.NilObjectID, errors.New("user cannot be nil")
	}
	if err := u.Validate(); err != nil {
		return primitive.NilObjectID, err
	}
	res, err := r.Collection.InsertOne(ctx, u)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.New("unexpected inserted id type")
	}
	return id, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	var u user.User
	if err := r.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}
