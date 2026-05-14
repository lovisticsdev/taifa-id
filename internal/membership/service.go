package membership

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-id/internal/audit"
	"taifa-id/internal/platform/clock"
	"taifa-id/internal/platform/ids"
	"taifa-id/internal/platform/postgres"
)

const activeStatus = "ACTIVE"

type Service struct {
	pool  *pgxpool.Pool
	repo  *Repository
	clock clock.Clock
}

func NewService(pool *pgxpool.Pool, repo *Repository, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.NewRealClock()
	}

	return &Service{
		pool:  pool,
		repo:  repo,
		clock: clk,
	}
}

func (s *Service) Create(ctx context.Context, req CreateMembershipRequest, correlationID string) (Membership, error) {
	personID := strings.TrimSpace(req.PersonID)
	organizationID := strings.TrimSpace(req.OrganizationID)
	membershipType := Type(strings.TrimSpace(req.MembershipType))

	if personID == "" || organizationID == "" || !IsValidType(membershipType) {
		return Membership{}, ErrValidation
	}

	membership := Membership{
		ID:             ids.NewMembershipID(),
		PersonID:       personID,
		OrganizationID: organizationID,
		MembershipType: membershipType,
		Status:         StatusActive,
	}

	var created Membership

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		personStatus, err := s.repo.PersonStatus(ctx, tx, personID)
		if err != nil {
			return err
		}

		if personStatus != activeStatus {
			return ErrPersonNotActive
		}

		organizationStatus, err := s.repo.OrganizationStatus(ctx, tx, organizationID)
		if err != nil {
			return err
		}

		if organizationStatus != activeStatus {
			return ErrOrganizationNotActive
		}

		created, err = s.repo.Create(ctx, tx, membership)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventMembershipCreated,
			SubjectID:     created.PersonID,
			ResourceType:  audit.ResourceOrganizationMembership,
			ResourceID:    created.ID,
			Action:        audit.ActionCreate,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"person_id":         created.PersonID,
				"organization_id":   created.OrganizationID,
				"membership_type":   string(created.MembershipType),
				"membership_status": string(created.Status),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write membership created audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateActiveMembership) {
			return Membership{}, ErrDuplicateActiveMembership
		}

		return Membership{}, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Membership, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Membership{}, ErrValidation
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByPerson(ctx context.Context, personID string) ([]Membership, error) {
	personID = strings.TrimSpace(personID)
	if personID == "" {
		return nil, ErrValidation
	}

	return s.repo.ListByPerson(ctx, personID)
}

func (s *Service) ListByOrganization(ctx context.Context, organizationID string) ([]Membership, error) {
	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		return nil, ErrValidation
	}

	return s.repo.ListByOrganization(ctx, organizationID)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, req UpdateMembershipStatusRequest, correlationID string) (Membership, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Membership{}, ErrValidation
	}

	status := Status(strings.TrimSpace(req.Status))
	if !IsValidStatus(status) {
		return Membership{}, ErrValidation
	}

	var updated Membership

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		updated, err = s.repo.UpdateStatus(ctx, tx, id, status)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventMembershipStatusChanged,
			SubjectID:     updated.PersonID,
			ResourceType:  audit.ResourceOrganizationMembership,
			ResourceID:    updated.ID,
			Action:        audit.ActionUpdateStatus,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"person_id":         updated.PersonID,
				"organization_id":   updated.OrganizationID,
				"membership_type":   string(updated.MembershipType),
				"membership_status": string(updated.Status),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write membership status changed audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return Membership{}, err
	}

	return updated, nil
}

func (s *Service) AddRole(ctx context.Context, membershipID string, req AddMembershipRoleRequest, correlationID string) (MembershipRole, error) {
	membershipID = strings.TrimSpace(membershipID)
	roleValue := Role(strings.TrimSpace(req.Role))

	if membershipID == "" || !IsValidRole(roleValue) {
		return MembershipRole{}, ErrValidation
	}

	role := MembershipRole{
		ID:           ids.NewMembershipRoleID(),
		MembershipID: membershipID,
		Role:         roleValue,
	}

	var created MembershipRole

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		membership, err := s.getByIDWithTx(ctx, tx, membershipID)
		if err != nil {
			return err
		}

		if membership.Status != StatusActive {
			return ErrMembershipNotActive
		}

		created, err = s.repo.AddRole(ctx, tx, role)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventMembershipRoleAdded,
			SubjectID:     membership.PersonID,
			ResourceType:  audit.ResourceMembershipRole,
			ResourceID:    created.ID,
			Action:        audit.ActionAddRole,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"membership_id":   created.MembershipID,
				"person_id":       membership.PersonID,
				"organization_id": membership.OrganizationID,
				"role":            string(created.Role),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write membership role added audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateRole) {
			return MembershipRole{}, ErrDuplicateRole
		}

		return MembershipRole{}, err
	}

	return created, nil
}

func (s *Service) ListRoles(ctx context.Context, membershipID string) ([]MembershipRole, error) {
	membershipID = strings.TrimSpace(membershipID)
	if membershipID == "" {
		return nil, ErrValidation
	}

	if _, err := s.repo.GetByID(ctx, membershipID); err != nil {
		return nil, err
	}

	return s.repo.ListRoles(ctx, membershipID)
}

func (s *Service) RemoveRole(ctx context.Context, membershipID string, role string, correlationID string) (MembershipRole, error) {
	membershipID = strings.TrimSpace(membershipID)
	roleValue := Role(strings.TrimSpace(role))

	if membershipID == "" || !IsValidRole(roleValue) {
		return MembershipRole{}, ErrValidation
	}

	var removed MembershipRole

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		membership, err := s.getByIDWithTx(ctx, tx, membershipID)
		if err != nil {
			return err
		}

		removed, err = s.repo.RemoveRole(ctx, tx, membershipID, roleValue)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventMembershipRoleRemoved,
			SubjectID:     membership.PersonID,
			ResourceType:  audit.ResourceMembershipRole,
			ResourceID:    removed.ID,
			Action:        audit.ActionRemoveRole,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"membership_id":   removed.MembershipID,
				"person_id":       membership.PersonID,
				"organization_id": membership.OrganizationID,
				"role":            string(removed.Role),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write membership role removed audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return MembershipRole{}, err
	}

	return removed, nil
}

func (s *Service) getByIDWithTx(ctx context.Context, tx pgx.Tx, id string) (Membership, error) {
	const query = `
		SELECT
			id,
			person_id,
			organization_id,
			membership_type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
		FROM organization_memberships
		WHERE id = $1
	`

	var membership Membership
	if err := scanMembership(tx.QueryRow(ctx, query, id), &membership); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Membership{}, ErrNotFound
		}

		return Membership{}, fmt.Errorf("get membership by id with tx: %w", err)
	}

	return membership, nil
}
