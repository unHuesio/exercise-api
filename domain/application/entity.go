package application

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name" binding:"required"`
	Email     string             `bson:"email" json:"email" binding:"required"`
	Status    string             `bson:"status" json:"status"`
	APIKey    string             `bson:"api_key" json:"api_key" binding:"required"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

func (a *Application) Validate() error {
	if a == nil {
		return errors.New("application cannot be nil")
	}
	if a.Name == "" {
		return errors.New("name is required")
	}
	if a.Email == "" {
		return errors.New("email is required")
	}
	if a.APIKey == "" {
		return errors.New("api_key is required")
	}
	return nil
}

func (a *Application) SetStatus(status string) {
	a.Status = status
}

func (a *Application) StampCreatedAt(now time.Time) {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
}
