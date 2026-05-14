package auth

import (
	"errors"
	"time"
)

type CredentialRecord struct {
	ID               string
	PersonID         string
	Username         string
	PasswordHash     string
	CredentialStatus string
	PersonStatus     string
}

type LoginResult struct {
	AccessToken  string
	TokenType    string
	ExpiresAt    time.Time
	SessionID    string
	PersonID     string
	CredentialID string
	Username     string
}

type IntrospectionResult struct {
	Active       bool
	PersonID     string
	CredentialID string
	Username     string
	SessionID    string
	IssuedAt     *time.Time
	ExpiresAt    *time.Time
}

var (
	ErrValidation         = errors.New("auth validation failed")
	ErrInvalidCredential  = errors.New("invalid username or password")
	ErrCredentialInactive = errors.New("credential is not active")
	ErrPersonInactive     = errors.New("person is not active")
	ErrTokenIssue         = errors.New("token issue failed")
)
