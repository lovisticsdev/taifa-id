package actorcontext

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-id/internal/audit"
	"taifa-id/internal/platform/clock"
	"taifa-id/internal/platform/ids"
	"taifa-id/internal/platform/token"
)

const activeStatus = "ACTIVE"

type TokenVerifier interface {
	Verify(rawToken string) (token.Claims, error)
}

type Service struct {
	pool   *pgxpool.Pool
	repo   *Repository
	tokens TokenVerifier
	clock  clock.Clock
}

func NewService(pool *pgxpool.Pool, repo *Repository, tokens TokenVerifier, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.NewRealClock()
	}

	return &Service{
		pool:   pool,
		repo:   repo,
		tokens: tokens,
		clock:  clk,
	}
}

func (s *Service) Resolve(ctx context.Context, req ResolveActorContextRequest, correlationID string) (ActorContext, error) {
	rawToken := strings.TrimSpace(req.Token)
	organizationID := strings.TrimSpace(req.OrganizationID)

	if rawToken == "" || organizationID == "" {
		return ActorContext{}, ErrValidation
	}

	actorContextID := ids.NewActorContextID()

	claims, err := s.tokens.Verify(rawToken)
	if err != nil {
		if auditErr := s.writeDenied(ctx, actorContextID, correlationID, "", "", organizationID, "", "invalid_token"); auditErr != nil {
			return ActorContext{}, auditErr
		}

		return ActorContext{}, ErrInvalidToken
	}

	record, err := s.repo.GetCredentialByID(ctx, claims.CredentialID)
	if err != nil {
		if auditErr := s.writeDenied(ctx, actorContextID, correlationID, claims.PersonID, claims.CredentialID, organizationID, claims.SessionID, "credential_not_found"); auditErr != nil {
			return ActorContext{}, auditErr
		}

		return ActorContext{}, ErrInvalidToken
	}

	if record.PersonID != claims.PersonID || record.Username != claims.Username {
		if auditErr := s.writeDenied(ctx, actorContextID, correlationID, claims.PersonID, claims.CredentialID, organizationID, claims.SessionID, "token_subject_mismatch"); auditErr != nil {
			return ActorContext{}, auditErr
		}

		return ActorContext{}, ErrInvalidToken
	}

	if record.CredentialStatus != activeStatus {
		if auditErr := s.writeDenied(ctx, actorContextID, correlationID, claims.PersonID, claims.CredentialID, organizationID, claims.SessionID, "credential_not_active"); auditErr != nil {
			return ActorContext{}, auditErr
		}

		return ActorContext{}, ErrCredentialInactive
	}

	if record.PersonStatus != activeStatus {
		if auditErr := s.writeDenied(ctx, actorContextID, correlationID, claims.PersonID, claims.CredentialID, organizationID, claims.SessionID, "person_not_active"); auditErr != nil {
			return ActorContext{}, auditErr
		}

		return ActorContext{}, ErrPersonInactive
	}

	organizationStatus, err := s.repo.OrganizationStatus(ctx, organizationID)
	if err != nil {
		if errors.Is(err, ErrOrganizationNotFound) {
			if auditErr := s.writeDenied(ctx, actorContextID, correlationID, claims.PersonID, claims.CredentialID, organizationID, claims.SessionID, "organization_not_found"); auditErr != nil {
				return ActorContext{}, auditErr
			}
		}

		return ActorContext{}, err
	}

	if organizationStatus != activeStatus {
		if auditErr := s.writeDenied(ctx, actorContextID, correlationID, claims.PersonID, claims.CredentialID, organizationID, claims.SessionID, "organization_not_active"); auditErr != nil {
			return ActorContext{}, auditErr
		}

		return ActorContext{}, ErrOrganizationInactive
	}

	memberships, err := s.repo.ActiveMemberships(ctx, claims.PersonID, organizationID)
	if err != nil {
		if errors.Is(err, ErrNoActiveMembership) {
			if auditErr := s.writeDenied(ctx, actorContextID, correlationID, claims.PersonID, claims.CredentialID, organizationID, claims.SessionID, "no_active_membership"); auditErr != nil {
				return ActorContext{}, auditErr
			}
		}

		return ActorContext{}, err
	}

	roles, err := s.repo.RolesForActiveMemberships(ctx, claims.PersonID, organizationID)
	if err != nil {
		return ActorContext{}, err
	}

	var issuedAt *time.Time
	var expiresAt *time.Time

	if claims.IssuedAt != nil {
		value := claims.IssuedAt.Time
		issuedAt = &value
	}

	if claims.ExpiresAt != nil {
		value := claims.ExpiresAt.Time
		expiresAt = &value
	}

	resolved := ActorContext{
		ID:             actorContextID,
		PersonID:       claims.PersonID,
		CredentialID:   claims.CredentialID,
		Username:       claims.Username,
		OrganizationID: organizationID,
		Memberships:    memberships,
		Roles:          roles,
		SessionID:      claims.SessionID,
		IssuedAt:       issuedAt,
		ExpiresAt:      expiresAt,
		ResolvedAt:     s.clock.Now(),
	}

	if err := s.writeAllowed(ctx, correlationID, resolved); err != nil {
		return ActorContext{}, err
	}

	return resolved, nil
}

func (s *Service) writeAllowed(ctx context.Context, correlationID string, actorContext ActorContext) error {
	membershipIDs := make([]string, 0, len(actorContext.Memberships))
	membershipTypes := make([]string, 0, len(actorContext.Memberships))

	for _, membership := range actorContext.Memberships {
		membershipIDs = append(membershipIDs, membership.ID)
		membershipTypes = append(membershipTypes, membership.MembershipType)
	}

	event := audit.Event{
		EventType:     audit.EventActorContextAllowed,
		SubjectID:     actorContext.PersonID,
		ActorID:       actorContext.PersonID,
		ResourceType:  audit.ResourceActorContext,
		ResourceID:    actorContext.ID,
		Action:        audit.ActionResolveActorContext,
		Result:        audit.ResultAllowed,
		CorrelationID: correlationID,
		Payload: map[string]any{
			"actor_context_id": actorContext.ID,
			"person_id":        actorContext.PersonID,
			"credential_id":    actorContext.CredentialID,
			"username":         actorContext.Username,
			"organization_id":  actorContext.OrganizationID,
			"membership_ids":   membershipIDs,
			"membership_types": membershipTypes,
			"roles":            actorContext.Roles,
			"session_id":       actorContext.SessionID,
		},
		CreatedAt: s.clock.Now(),
	}

	if err := audit.Insert(ctx, s.pool, event); err != nil {
		return fmt.Errorf("write actor context allowed audit event: %w", err)
	}

	return nil
}

func (s *Service) writeDenied(
	ctx context.Context,
	actorContextID string,
	correlationID string,
	personID string,
	credentialID string,
	organizationID string,
	sessionID string,
	reason string,
) error {
	event := audit.Event{
		EventType:     audit.EventActorContextDenied,
		SubjectID:     personID,
		ActorID:       personID,
		ResourceType:  audit.ResourceActorContext,
		ResourceID:    actorContextID,
		Action:        audit.ActionResolveActorContext,
		Result:        audit.ResultDenied,
		CorrelationID: correlationID,
		Payload: map[string]any{
			"actor_context_id": actorContextID,
			"person_id":        personID,
			"credential_id":    credentialID,
			"organization_id":  organizationID,
			"session_id":       sessionID,
			"reason":           reason,
		},
		CreatedAt: s.clock.Now(),
	}

	if err := audit.Insert(ctx, s.pool, event); err != nil {
		return fmt.Errorf("write actor context denied audit event: %w", err)
	}

	return nil
}
