package user

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	Create(ctx context.Context, user *User) (primitive.ObjectID, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}
