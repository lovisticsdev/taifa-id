package organization

import "time"

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	PrimaryType string `json:"primary_type"`
}

type UpdateOrganizationStatusRequest struct {
	Status string `json:"status"`
}

type AddOrganizationCapabilityRequest struct {
	Capability string `json:"capability"`
}

type OrganizationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PrimaryType string    `json:"primary_type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrganizationCapabilityResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Capability     string    `json:"capability"`
	CreatedAt      time.Time `json:"created_at"`
}

func ToOrganizationResponse(org Organization) OrganizationResponse {
	return OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		PrimaryType: string(org.PrimaryType),
		Status:      string(org.Status),
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

func ToCapabilityResponse(capability OrganizationCapability) OrganizationCapabilityResponse {
	return OrganizationCapabilityResponse{
		ID:             capability.ID,
		OrganizationID: capability.OrganizationID,
		Capability:     string(capability.Capability),
		CreatedAt:      capability.CreatedAt,
	}
}

func ToOrganizationResponses(orgs []Organization) []OrganizationResponse {
	responses := make([]OrganizationResponse, 0, len(orgs))
	for _, org := range orgs {
		responses = append(responses, ToOrganizationResponse(org))
	}

	return responses
}

func ToCapabilityResponses(capabilities []OrganizationCapability) []OrganizationCapabilityResponse {
	responses := make([]OrganizationCapabilityResponse, 0, len(capabilities))
	for _, capability := range capabilities {
		responses = append(responses, ToCapabilityResponse(capability))
	}

	return responses
}
