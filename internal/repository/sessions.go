package repository

import (
	"authentication-service/internal/domain"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type SessionRepo struct {
	db *mongo.Collection
}

func NewSessionRepo(db *mongo.Database) *SessionRepo {
	return &SessionRepo{
		db: db.Collection(sessionCollection),
	}
}

func (r *SessionRepo) SetSession(ctx context.Context, session domain.Session) error {
	filter := bson.M{"userID": session.UserID}
	update := bson.M{
		"$set": bson.M{
			"lastVisitAt":  time.Now(),
			"refreshToken": session.RefreshToken,
			"expiresAt":    session.ExpiredAt,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updatedDoc bson.M
	err := r.db.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.ErrUserNotFound
		}
		return err
	}

	return nil
}

func (r *SessionRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	session := &domain.Session{}
	if err := r.db.FindOne(ctx, bson.M{
		"refreshToken": refreshToken,
	}).Decode(&session); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &domain.Session{}, domain.ErrUserNotFound
		}

		return &domain.Session{}, err
	}

	return session, nil
}
