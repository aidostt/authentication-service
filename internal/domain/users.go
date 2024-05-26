package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	UserRole      = "user"
	AdminRole     = "admin"
	ActivatedRole = "activated"
)

type User struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name             string             `json:"name" bson:"name"`
	Surname          string             `json:"surname" bson:"surname"`
	Phone            string             `json:"phone" bson:"phone"`
	Email            string             `json:"email" bson:"email"`
	Roles            []string           `json:"roles,omitempty" bson:"roles"`
	Password         string             `json:"password" bson:"password"`
	Activated        bool               `json:"activated" bson:"activated"`
	VerificationCode VerificationCode   `json:"verification_code" bson:"verification_code"`
}
