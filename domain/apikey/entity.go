package apikey

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApiKey struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	APIKey    string             `bson:"api_key" json:"api_key"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	IsValid   bool               `bson:"is_valid" json:"is_valid"`
	Account   string             `bson:"account" json:"account" binding:"required"`
}

func (a *ApiKey) Validate() error {
	if a == nil {
		return errors.New("api key cannot be nil")
	}
	if a.Account == "" {
		return errors.New("account is required")
	}
	return nil
}

func (a *ApiKey) Invalidate() {
	a.IsValid = false
}

func (a *ApiKey) StampCreatedAt(now time.Time) {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
}
