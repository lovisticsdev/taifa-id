package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const DefaultBcryptCost = bcrypt.DefaultCost

var ErrInvalidPassword = errors.New("invalid password")

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost == 0 {
		cost = DefaultBcryptCost
	}

	return &BcryptHasher{
		cost: cost,
	}
}

func (h *BcryptHasher) Hash(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrInvalidPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), h.cost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func (h *BcryptHasher) Verify(plaintext string, hash string) bool {
	if plaintext == "" || hash == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
