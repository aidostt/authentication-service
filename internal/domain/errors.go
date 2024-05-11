package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user doesn't exists")
	ErrUserAlreadyExists = errors.New("user with such email already exists")
	ErrWrongPassword     = errors.New("wrong password")
	ErrSessionExpired    = errors.New("session is expired")
	ErrUnathorized       = errors.New("unauthorized access")
)
