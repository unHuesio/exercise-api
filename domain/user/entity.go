package user

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email           string             `bson:"email" json:"email" binding:"required"`
	Provider        string             `bson:"provider" json:"provider" binding:"required"`
	ProviderSubject string             `bson:"provider_subject" json:"provider_subject" binding:"required"`
}

func (u *User) Validate() error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Provider == "" {
		return errors.New("provider is required")
	}
	if u.ProviderSubject == "" {
		return errors.New("provider subject is required")
	}
	return nil
}
