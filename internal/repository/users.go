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

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user *domain.User
	if err := r.db.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
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
		return domain.ErrUserNotFound
	}

	return nil // Return nil if deletion was successful
}

func (r *UsersRepo) GetByID(ctx context.Context, userID primitive.ObjectID) (*domain.User, error) {
	var user *domain.User
	if err := r.db.FindOne(ctx, bson.M{
		"_id": userID,
	}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

func (r *UsersRepo) Update(ctx context.Context, user *domain.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{
		"name":              user.Name,
		"surname":           user.Surname,
		"phone":             user.Phone,
		"email":             user.Email,
		"roles":             user.Roles,
		"password":          user.Password,
		"activated":         user.Activated,
		"verification_code": user.VerificationCode,
	}}

	result, err := r.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UsersRepo) Activate(ctx context.Context, id primitive.ObjectID, activate bool) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"activated": activate,
	}}

	result, err := r.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UsersRepo) AddRole(ctx context.Context, id primitive.ObjectID, role string) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$addToSet": bson.M{"roles": role}}

	result, err := r.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UsersRepo) RemoveRole(ctx context.Context, id primitive.ObjectID, role string) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$pull": bson.M{"roles": role}}

	result, err := r.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
