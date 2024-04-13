package repository

import (
	"authentication-service/internal/domain"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users interface {
	Create(context.Context, *domain.User) error
	GetByEmail(context.Context, string) (domain.User, error)
	Delete(context.Context, primitive.ObjectID, string) error
}
type Sessions interface {
	SetSession(ctx context.Context, session domain.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) (domain.User, error)
	//TODO: delete session
}

type Models struct {
	Users    Users
	Sessions Sessions
}

func NewModels(db *mongo.Database) *Models {
	return &Models{
		Users:    NewUsersRepo(db),
		Sessions: NewSessionRepo(db),
	}
}
