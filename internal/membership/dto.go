package membership

import "time"

type CreateMembershipRequest struct {
	PersonID       string `json:"person_id"`
	OrganizationID string `json:"organization_id"`
	MembershipType string `json:"membership_type"`
}

type UpdateMembershipStatusRequest struct {
	Status string `json:"status"`
}

type AddMembershipRoleRequest struct {
	Role string `json:"role"`
}

type MembershipResponse struct {
	ID             string     `json:"id"`
	PersonID       string     `json:"person_id"`
	OrganizationID string     `json:"organization_id"`
	MembershipType string     `json:"membership_type"`
	Status         string     `json:"status"`
	StartsAt       time.Time  `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type MembershipRoleResponse struct {
	ID           string    `json:"id"`
	MembershipID string    `json:"membership_id"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

func ToMembershipResponse(m Membership) MembershipResponse {
	return MembershipResponse{
		ID:             m.ID,
		PersonID:       m.PersonID,
		OrganizationID: m.OrganizationID,
		MembershipType: string(m.MembershipType),
		Status:         string(m.Status),
		StartsAt:       m.StartsAt,
		EndsAt:         m.EndsAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func ToMembershipRoleResponse(role MembershipRole) MembershipRoleResponse {
	return MembershipRoleResponse{
		ID:           role.ID,
		MembershipID: role.MembershipID,
		Role:         string(role.Role),
		CreatedAt:    role.CreatedAt,
	}
}

func ToMembershipResponses(memberships []Membership) []MembershipResponse {
	responses := make([]MembershipResponse, 0, len(memberships))
	for _, membership := range memberships {
		responses = append(responses, ToMembershipResponse(membership))
	}

	return responses
}

func ToMembershipRoleResponses(roles []MembershipRole) []MembershipRoleResponse {
	responses := make([]MembershipRoleResponse, 0, len(roles))
	for _, role := range roles {
		responses = append(responses, ToMembershipRoleResponse(role))
	}

	return responses
}
