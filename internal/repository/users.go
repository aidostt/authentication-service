package repository

import (
	"authentication-service/internal/domain"
	"authentication-service/pkg/database/mongodb"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UsersRepo struct {
	db *mongo.Collection
}

func NewUsersRepo(db *mongo.Database) *UsersRepo {
	return &UsersRepo{
		db: db.Collection(usersCollection),
	}
}

func (r *UsersRepo) Create(ctx context.Context, user *domain.User) error {
	result, err := r.db.InsertOne(ctx, user)
	if err != nil {
		if mongodb.IsDuplicate(err) {
			return domain.ErrUserAlreadyExists
		}
		return err
	}
	// Assert that the inserted ID is an ObjectID
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("failed to convert inserted ID to ObjectID")
	}
	user.ID = id
	return nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {

	var user domain.User
	if err := r.db.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) Delete(ctx context.Context, id primitive.ObjectID, email string) error {
	// Create a filter to match the document to delete
	filter := bson.M{"_id": id, "email": email}

	// Attempt to delete the document
	result, err := r.db.DeleteOne(ctx, filter)
	if err != nil {
		return err // Return the error if deletion failed
	}

	// Check if the document was actually deleted
	if result.DeletedCount == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil // Return nil if deletion was successful
}
