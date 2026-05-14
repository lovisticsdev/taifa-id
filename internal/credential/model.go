package credential

import (
	"errors"
	"time"
)

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusDisabled Status = "DISABLED"
	StatusLocked   Status = "LOCKED"
)

type Credential struct {
	ID           string
	PersonID     string
	Username     string
	PasswordHash string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	ErrValidation        = errors.New("credential validation failed")
	ErrNotFound          = errors.New("credential not found")
	ErrReferenceNotFound = errors.New("referenced person not found")
	ErrPersonNotActive   = errors.New("person is not active")
	ErrDuplicateUsername = errors.New("duplicate username")
	ErrHashPassword      = errors.New("hash password failed")
)

func IsValidStatus(status Status) bool {
	switch status {
	case StatusActive,
		StatusDisabled,
		StatusLocked:
		return true
	default:
		return false
	}
}
