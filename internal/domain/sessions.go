package domain

import "time"

type Session struct {
	RefreshToken string    `json:"refreshToken" bson:"refreshToken"`
	ExpiredAt    time.Time `json:"expiresAt" bson:"expiresAt"`
}
