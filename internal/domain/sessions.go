package domain

import "time"

type Session struct {
	UserID       string    `json:"userID"`
	RefreshToken string    `json:"refreshToken"`
	ExpiredAt    time.Time `json:"expiresAt"`
}

type VerificationCode struct {
	Code      string    `json:"code"`
	ExpiredAt time.Time `json:"expiresAt"`
}
