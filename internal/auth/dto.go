package auth

import "time"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	SessionID    string    `json:"session_id"`
	PersonID     string    `json:"person_id"`
	CredentialID string    `json:"credential_id"`
	Username     string    `json:"username"`
}

type IntrospectRequest struct {
	Token string `json:"token"`
}

type IntrospectResponse struct {
	Active       bool       `json:"active"`
	PersonID     string     `json:"person_id,omitempty"`
	CredentialID string     `json:"credential_id,omitempty"`
	Username     string     `json:"username,omitempty"`
	SessionID    string     `json:"session_id,omitempty"`
	IssuedAt     *time.Time `json:"issued_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

func ToLoginResponse(result LoginResult) LoginResponse {
	return LoginResponse{
		AccessToken:  result.AccessToken,
		TokenType:    result.TokenType,
		ExpiresAt:    result.ExpiresAt,
		SessionID:    result.SessionID,
		PersonID:     result.PersonID,
		CredentialID: result.CredentialID,
		Username:     result.Username,
	}
}

func ToIntrospectResponse(result IntrospectionResult) IntrospectResponse {
	return IntrospectResponse{
		Active:       result.Active,
		PersonID:     result.PersonID,
		CredentialID: result.CredentialID,
		Username:     result.Username,
		SessionID:    result.SessionID,
		IssuedAt:     result.IssuedAt,
		ExpiresAt:    result.ExpiresAt,
	}
}
