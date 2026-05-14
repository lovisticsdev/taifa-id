package person

import (
	"errors"
	"time"
)

type Status string

const (
	StatusActive          Status = "ACTIVE"
	StatusSuspended       Status = "SUSPENDED"
	StatusDeceased        Status = "DECEASED"
	StatusDuplicateReview Status = "DUPLICATE_REVIEW"
	StatusDisabled        Status = "DISABLED"
)

type Person struct {
	ID           string
	SyntheticNIN string
	DisplayName  string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	ErrValidation            = errors.New("person validation failed")
	ErrNotFound              = errors.New("person not found")
	ErrDuplicateSyntheticNIN = errors.New("duplicate synthetic NIN")
)

func IsValidStatus(status Status) bool {
	switch status {
	case StatusActive,
		StatusSuspended,
		StatusDeceased,
		StatusDuplicateReview,
		StatusDisabled:
		return true
	default:
		return false
	}
}
