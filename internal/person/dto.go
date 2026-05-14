package person

import "time"

type CreatePersonRequest struct {
	SyntheticNIN string `json:"synthetic_nin"`
	DisplayName  string `json:"display_name"`
}

type UpdatePersonStatusRequest struct {
	Status string `json:"status"`
}

type PersonResponse struct {
	ID           string    `json:"id"`
	SyntheticNIN string    `json:"synthetic_nin"`
	DisplayName  string    `json:"display_name"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToResponse(p Person) PersonResponse {
	return PersonResponse{
		ID:           p.ID,
		SyntheticNIN: p.SyntheticNIN,
		DisplayName:  p.DisplayName,
		Status:       string(p.Status),
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}
