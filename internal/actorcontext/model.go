package actorcontext

import (
	"errors"
	"time"
)

type CredentialRecord struct {
	ID               string
	PersonID         string
	Username         string
	CredentialStatus string
	PersonStatus     string
}

type MembershipContext struct {
	ID             string
	MembershipType string
}

type ActorContext struct {
	ID             string
	PersonID       string
	CredentialID   string
	Username       string
	OrganizationID string
	Memberships    []MembershipContext
	Roles          []string
	SessionID      string
	IssuedAt       *time.Time
	ExpiresAt      *time.Time
	ResolvedAt     time.Time
}

var (
	ErrValidation           = errors.New("actor context validation failed")
	ErrInvalidToken         = errors.New("invalid token")
	ErrCredentialInactive   = errors.New("credential is not active")
	ErrPersonInactive       = errors.New("person is not active")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrOrganizationInactive = errors.New("organization is not active")
	ErrNoActiveMembership   = errors.New("no active membership for organization")
)
