package credential

import "time"

type CreateCredentialRequest struct {
	PersonID string `json:"person_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CredentialResponse struct {
	ID        string    `json:"id"`
	PersonID  string    `json:"person_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToCredentialResponse(credential Credential) CredentialResponse {
	return CredentialResponse{
		ID:        credential.ID,
		PersonID:  credential.PersonID,
		Username:  credential.Username,
		Status:    string(credential.Status),
		CreatedAt: credential.CreatedAt,
		UpdatedAt: credential.UpdatedAt,
	}
}

func ToCredentialResponses(credentials []Credential) []CredentialResponse {
	responses := make([]CredentialResponse, 0, len(credentials))
	for _, credential := range credentials {
		responses = append(responses, ToCredentialResponse(credential))
	}

	return responses
}
