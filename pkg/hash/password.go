package hash

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher interface {
	Hash(string) ([]byte, error)
	Matches(string, []byte) (bool, error)
}

type Hasher struct {
	cost int
}

func NewHasher(cost int) *Hasher {
	return &Hasher{cost: cost}
}

func (h *Hasher) Hash(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return nil, err
	}
	return hash, nil

}

func (h *Hasher) Matches(plaintextPassword string, hashed []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashed, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
