package repository

import (
	"authentication-service/internal/domain"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users interface {
	Create(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (domain.User, error)
	SetSession(ctx context.Context, userID primitive.ObjectID, session domain.Session) error
}

type Models struct {
	Users Users
}

func NewModels(db *mongo.Database) *Models {
	return &Models{
		Users: NewUsersRepo(db),
	}
}
