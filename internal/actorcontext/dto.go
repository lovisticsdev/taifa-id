package actorcontext

import "time"

type ResolveActorContextRequest struct {
	Token          string `json:"token"`
	OrganizationID string `json:"organization_id"`
}

type MembershipContextResponse struct {
	ID             string `json:"id"`
	MembershipType string `json:"membership_type"`
}

type ActorContextResponse struct {
	ActorContextID string                      `json:"actor_context_id"`
	PersonID       string                      `json:"person_id"`
	CredentialID   string                      `json:"credential_id"`
	Username       string                      `json:"username"`
	OrganizationID string                      `json:"organization_id"`
	Memberships    []MembershipContextResponse `json:"memberships"`
	Roles          []string                    `json:"roles"`
	SessionID      string                      `json:"session_id"`
	IssuedAt       *time.Time                  `json:"issued_at,omitempty"`
	ExpiresAt      *time.Time                  `json:"expires_at,omitempty"`
	ResolvedAt     time.Time                   `json:"resolved_at"`
}

func ToResponse(actorContext ActorContext) ActorContextResponse {
	memberships := make([]MembershipContextResponse, 0, len(actorContext.Memberships))
	for _, membership := range actorContext.Memberships {
		memberships = append(memberships, MembershipContextResponse{
			ID:             membership.ID,
			MembershipType: membership.MembershipType,
		})
	}

	return ActorContextResponse{
		ActorContextID: actorContext.ID,
		PersonID:       actorContext.PersonID,
		CredentialID:   actorContext.CredentialID,
		Username:       actorContext.Username,
		OrganizationID: actorContext.OrganizationID,
		Memberships:    memberships,
		Roles:          actorContext.Roles,
		SessionID:      actorContext.SessionID,
		IssuedAt:       actorContext.IssuedAt,
		ExpiresAt:      actorContext.ExpiresAt,
		ResolvedAt:     actorContext.ResolvedAt,
	}
}
