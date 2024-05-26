package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Session struct {
	UserID       primitive.ObjectID `json:"userID" bson:"userID"`
	RefreshToken string             `json:"refreshToken" bson:"refreshToken"`
	ExpiredAt    time.Time          `json:"expiresAt" bson:"expiresAt"`
}

type VerificationCode struct {
	Code      string    `json:"code" bson:"code"`
	ExpiredAt time.Time `json:"expiresAt" bson:"expiresAt"`
}
